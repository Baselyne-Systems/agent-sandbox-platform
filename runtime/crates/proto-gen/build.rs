use std::path::PathBuf;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let manifest_dir = PathBuf::from(std::env::var("CARGO_MANIFEST_DIR")?);
    let proto_dir = manifest_dir.join("../../../proto");
    let proto_include = proto_dir.clone();

    // Ensure the output directory exists.
    let out_dir = manifest_dir.join("src/gen");
    std::fs::create_dir_all(&out_dir)?;

    let protos = &[
        proto_dir.join("platform/runtime/v1/runtime.proto"),
        proto_dir.join("platform/identity/v1/identity.proto"),
        proto_dir.join("platform/activity/v1/activity.proto"),
        proto_dir.join("platform/guardrails/v1/guardrails.proto"),
        proto_dir.join("platform/workspace/v1/workspace.proto"),
    ];

    tonic_build::configure()
        .build_server(true)
        .build_client(true)
        .out_dir(&out_dir)
        .compile_protos(protos, &[proto_include])?;

    // Rerun if any proto file changes.
    for proto in protos {
        println!("cargo:rerun-if-changed={}", proto.display());
    }

    Ok(())
}
