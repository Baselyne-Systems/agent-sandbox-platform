use anyhow::Result;
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

/// Intercepts and executes tool calls on behalf of agents. In a production
/// system this would enforce sandboxing, capture I/O, and route to the actual
/// tool implementation.
#[derive(Debug)]
pub struct ToolInterceptor;

impl ToolInterceptor {
    /// Create a new tool interceptor.
    pub fn new() -> Self {
        Self
    }

    /// Intercept a tool request, execute it (stub), and return the result.
    ///
    /// Stub implementation: always returns success with an empty object output.
    pub fn intercept(&self, request: ToolRequest) -> Result<ToolResult> {
        info!(
            tool_name = %request.tool_name,
            "intercepting tool call — stub execution"
        );

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
    fn stub_interceptor_returns_success() {
        let interceptor = ToolInterceptor::new();
        let request = ToolRequest {
            tool_name: "read_file".to_string(),
            parameters: serde_json::json!({"path": "/etc/hosts"}),
        };

        let result = interceptor.intercept(request).unwrap();
        assert!(result.success);
        assert_eq!(result.output["status"], "ok");
    }
}
