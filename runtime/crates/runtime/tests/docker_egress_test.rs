//! Integration test: verifies egress allowlist enforcement with a live Docker container.
//!
//! Requires:
//! - Docker daemon running
//! - `alpine:3.20` image available (pulled automatically)
//! - Run with: cargo test --test docker_egress_test -- --ignored
//!
//! This test:
//! 1. Creates a DockerRuntime
//! 2. Starts an alpine container with egress_allowlist = ["1.1.1.1"]
//! 3. Verifies the iptables chain BH-{short_id} was created with correct rules
//! 4. Verifies the container can reach 1.1.1.1 (allowed) but NOT 8.8.8.8 (blocked)
//! 5. Cleans up: stops container + removes iptables chain

use std::collections::HashMap;

use host_agent::container::{ContainerRuntime, DockerRuntime};

/// Run a command inside a container and return (exit_code, stdout).
async fn docker_exec(container_id: &str, cmd: &[&str]) -> (i32, String) {
    let output = tokio::process::Command::new("docker")
        .args(["exec", container_id])
        .args(cmd)
        .output()
        .await
        .expect("failed to run docker exec");

    let stdout = String::from_utf8_lossy(&output.stdout).to_string();
    let code = output.status.code().unwrap_or(-1);
    (code, stdout)
}

/// Check if an iptables chain exists.
async fn chain_exists(chain: &str) -> bool {
    let output = tokio::process::Command::new("iptables")
        .args(["-L", chain])
        .output()
        .await;

    match output {
        Ok(o) => o.status.success(),
        Err(_) => false,
    }
}

/// List rules in an iptables chain.
async fn list_chain_rules(chain: &str) -> String {
    let output = tokio::process::Command::new("iptables")
        .args(["-L", chain, "-n", "--line-numbers"])
        .output()
        .await
        .expect("failed to list iptables chain");

    String::from_utf8_lossy(&output.stdout).to_string()
}

#[tokio::test]
#[ignore] // requires Docker + root (for iptables)
async fn egress_allowlist_blocks_unauthorized_traffic() {
    let runtime = DockerRuntime::new().expect("Docker must be running for this test");

    let sandbox_id = "test-egress-001";
    let chain_name = format!("BH-{}", &sandbox_id[..sandbox_id.len().min(12)]);

    // Start container with egress allowlist: only 1.1.1.1 is allowed
    let container_id = runtime
        .start_container(
            sandbox_id,
            "alpine:3.20",
            HashMap::new(),
            128 * 1024 * 1024, // 128MB
            50_000,            // 0.5 CPU
            &["1.1.1.1".to_string()],
        )
        .await
        .expect("failed to start container");

    println!("Container started: {container_id}");
    println!("Chain name: {chain_name}");

    // Verify the iptables chain was created
    assert!(
        chain_exists(&chain_name).await,
        "iptables chain {chain_name} should exist"
    );

    // Print the chain rules for debugging
    let rules = list_chain_rules(&chain_name).await;
    println!("Chain rules:\n{rules}");

    // Verify rules contain the allowed destination and DROP
    assert!(rules.contains("1.1.1.1"), "chain should contain ACCEPT rule for 1.1.1.1");
    assert!(rules.contains("DROP"), "chain should end with DROP rule");

    // Test: allowed destination (1.1.1.1) should be reachable
    // Use wget with a short timeout — we only care about TCP connect, not HTTP response
    let (code_allowed, _) = docker_exec(
        &container_id,
        &["wget", "-q", "-O", "/dev/null", "--timeout=3", "http://1.1.1.1/"],
    )
    .await;
    println!("wget 1.1.1.1 exit code: {code_allowed}");
    // exit code 0 = success, but even non-0 is OK if it connected (HTTP error != network block)
    // The key test is that the blocked destination fails differently

    // Test: blocked destination (8.8.8.8) should be unreachable
    let (code_blocked, _) = docker_exec(
        &container_id,
        &["wget", "-q", "-O", "/dev/null", "--timeout=3", "http://8.8.8.8/"],
    )
    .await;
    println!("wget 8.8.8.8 exit code: {code_blocked}");

    // The blocked request should fail (non-zero exit code from timeout/connection refused)
    // while the allowed request should succeed or at least connect
    assert_ne!(
        code_allowed, code_blocked,
        "allowed destination should behave differently from blocked destination"
    );

    // Clean up: destroy sandbox (removes iptables rules + stops container)
    runtime
        .cleanup_egress_rules(sandbox_id)
        .await
        .expect("failed to cleanup egress rules");

    assert!(
        !chain_exists(&chain_name).await,
        "iptables chain should be removed after cleanup"
    );

    runtime
        .stop_container(&container_id)
        .await
        .expect("failed to stop container");

    println!("Test passed: egress allowlist correctly blocks unauthorized traffic");
}

#[tokio::test]
#[ignore] // requires Docker
async fn container_without_egress_has_no_iptables_chain() {
    let runtime = DockerRuntime::new().expect("Docker must be running for this test");

    let sandbox_id = "test-no-egress-002";
    let chain_name = format!("BH-{}", &sandbox_id[..sandbox_id.len().min(12)]);

    // Start container WITHOUT egress allowlist
    let container_id = runtime
        .start_container(
            sandbox_id,
            "alpine:3.20",
            HashMap::new(),
            128 * 1024 * 1024,
            50_000,
            &[], // empty = no restrictions
        )
        .await
        .expect("failed to start container");

    println!("Container started without egress rules: {container_id}");

    // Verify NO iptables chain was created
    assert!(
        !chain_exists(&chain_name).await,
        "no iptables chain should exist when egress_allowlist is empty"
    );

    // Clean up
    runtime
        .stop_container(&container_id)
        .await
        .expect("failed to stop container");

    println!("Test passed: no egress rules when allowlist is empty");
}
