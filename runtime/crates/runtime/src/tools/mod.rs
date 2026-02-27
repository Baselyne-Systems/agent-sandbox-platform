use std::collections::HashMap;

use anyhow::{anyhow, Result};
use serde_json::Value;
use tracing::info;

/// A request to execute a tool, submitted by an agent.
#[derive(Debug, Clone)]
pub struct ToolRequest {
    /// Name of the tool to invoke.
    pub tool_name: String,
    /// Parameters for the tool call.
    pub parameters: Value,
}

/// The result of a tool execution.
#[derive(Debug, Clone)]
pub struct ToolResult {
    /// Whether the tool executed successfully.
    pub success: bool,
    /// Output returned by the tool.
    pub output: Value,
}

/// Intercepts and executes tool calls on behalf of agents. Enforces allowed-tool
/// lists and passes environment variables to tool backends.
#[derive(Debug, Clone)]
pub struct ToolInterceptor;

impl ToolInterceptor {
    /// Create a new tool interceptor.
    pub fn new() -> Self {
        Self
    }

    /// Intercept a tool request, check allowed tools, and execute (stub).
    ///
    /// Returns an error if the tool is not in the allowed list.
    /// `env_vars` are passed through for future tool backends that need them.
    pub fn intercept(
        &self,
        request: ToolRequest,
        allowed_tools: &[String],
        _env_vars: &HashMap<String, String>,
    ) -> Result<ToolResult> {
        // Enforce allowed tools list (empty list means all tools allowed)
        if !allowed_tools.is_empty()
            && !allowed_tools.iter().any(|t| t == &request.tool_name)
        {
            return Err(anyhow!(
                "tool '{}' is not in the allowed tools list",
                request.tool_name
            ));
        }

        info!(
            tool_name = %request.tool_name,
            "intercepting tool call — stub execution"
        );

        // Stub execution — actual tool backends are a future phase.
        Ok(ToolResult {
            success: true,
            output: serde_json::json!({
                "status": "ok",
                "tool": request.tool_name,
                "message": "stub execution — no real tool backend"
            }),
        })
    }
}

impl Default for ToolInterceptor {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn interceptor_allows_listed_tool() {
        let interceptor = ToolInterceptor::new();
        let request = ToolRequest {
            tool_name: "read_file".to_string(),
            parameters: serde_json::json!({"path": "/etc/hosts"}),
        };

        let result = interceptor
            .intercept(
                request,
                &vec!["read_file".to_string(), "write_file".to_string()],
                &HashMap::new(),
            )
            .unwrap();
        assert!(result.success);
        assert_eq!(result.output["status"], "ok");
    }

    #[test]
    fn interceptor_denies_unlisted_tool() {
        let interceptor = ToolInterceptor::new();
        let request = ToolRequest {
            tool_name: "exec".to_string(),
            parameters: serde_json::json!({}),
        };

        let result = interceptor.intercept(
            request,
            &vec!["read_file".to_string()],
            &HashMap::new(),
        );
        assert!(result.is_err());
        assert!(result.unwrap_err().to_string().contains("not in the allowed tools list"));
    }

    #[test]
    fn interceptor_allows_all_when_empty_list() {
        let interceptor = ToolInterceptor::new();
        let request = ToolRequest {
            tool_name: "anything".to_string(),
            parameters: serde_json::json!({}),
        };

        let result = interceptor
            .intercept(request, &vec![], &HashMap::new())
            .unwrap();
        assert!(result.success);
    }
}
