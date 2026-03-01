"""MCP (Model Context Protocol) adapter for Bulkhead.

Exposes Bulkhead tools as an MCP-compatible tool server using JSON-RPC 2.0 over stdio.

Usage::

    from bulkhead import BulkheadAgent, tool
    from bulkhead.mcp import BulkheadMCPServer

    @tool("search", description="Search documents")
    def search(query: str) -> str:
        return f"Results for {query}"

    with BulkheadAgent(tools=[search]) as agent:
        server = BulkheadMCPServer(agent)
        server.run()  # Reads JSON-RPC from stdin, writes to stdout
"""

from __future__ import annotations

import json
import sys
from typing import TYPE_CHECKING, Any

from .types import Verdict

if TYPE_CHECKING:
    from .client import BulkheadAgent


class BulkheadMCPServer:
    """Wraps a BulkheadAgent as an MCP tool provider.

    Implements the MCP tool protocol (tools/list, tools/call)
    over JSON-RPC 2.0 on stdio.
    """

    def __init__(self, agent: BulkheadAgent) -> None:
        self._agent = agent

    def run(self) -> None:
        """Run the MCP server, reading JSON-RPC requests from stdin."""
        for line in sys.stdin:
            line = line.strip()
            if not line:
                continue
            try:
                request = json.loads(line)
            except json.JSONDecodeError:
                self._write_error(None, -32700, "Parse error")
                continue

            response = self._handle_request(request)
            if response is not None:
                self._write(response)

    def _handle_request(self, request: dict[str, Any]) -> dict[str, Any] | None:
        req_id = request.get("id")
        method = request.get("method", "")

        if method == "initialize":
            return self._result(req_id, {
                "protocolVersion": "2024-11-05",
                "capabilities": {"tools": {}},
                "serverInfo": {"name": "bulkhead", "version": "0.1.0"},
            })

        if method == "tools/list":
            tools = []
            for defn in self._agent._tools.values():
                tools.append({
                    "name": defn.name,
                    "description": defn.description,
                    "inputSchema": defn.input_schema or {"type": "object", "properties": {}},
                })
            return self._result(req_id, {"tools": tools})

        if method == "tools/call":
            params = request.get("params", {})
            tool_name = params.get("name", "")
            arguments = params.get("arguments", {})

            try:
                result = self._agent.execute_tool(tool_name, arguments)
            except Exception as exc:
                return self._result(req_id, {
                    "content": [{"type": "text", "text": str(exc)}],
                    "isError": True,
                })

            if result.verdict == Verdict.DENY:
                return self._result(req_id, {
                    "content": [{"type": "text", "text": f"DENIED: {result.denial_reason}"}],
                    "isError": True,
                })

            if result.verdict == Verdict.ESCALATE:
                return self._result(req_id, {
                    "content": [{"type": "text", "text": f"ESCALATED: {result.escalation_id}"}],
                    "isError": True,
                })

            output = result.result if result.result is not None else ""
            return self._result(req_id, {
                "content": [{"type": "text", "text": str(output)}],
            })

        # Notifications (no id) don't get responses
        if req_id is None:
            return None

        return self._error(req_id, -32601, f"Method not found: {method}")

    def _result(self, req_id: Any, result: Any) -> dict[str, Any]:
        return {"jsonrpc": "2.0", "id": req_id, "result": result}

    def _error(self, req_id: Any, code: int, message: str) -> dict[str, Any]:
        return {"jsonrpc": "2.0", "id": req_id, "error": {"code": code, "message": message}}

    def _write_error(self, req_id: Any, code: int, message: str) -> None:
        self._write(self._error(req_id, code, message))

    def _write(self, response: dict[str, Any]) -> None:
        sys.stdout.write(json.dumps(response) + "\n")
        sys.stdout.flush()
