"""Tests for MCP JSON-RPC server."""

from __future__ import annotations

import json
import io
from unittest.mock import MagicMock, patch

import pytest

from bulkhead.decorators import tool, ToolDefinition
from bulkhead.mcp import BulkheadMCPServer
from bulkhead.types import ToolResult, Verdict


@tool("search", description="Search documents")
def search(query: str) -> str:
    return f"Results for {query}"


@tool("calculate", description="Calculate expression")
def calculate(expr: str) -> str:
    return f"Result: {eval(expr)}"


def make_server() -> tuple[BulkheadMCPServer, MagicMock]:
    """Create an MCP server with a mocked BulkheadAgent."""
    agent = MagicMock()
    agent._tools = {
        "search": search._bulkhead_tool,
        "calculate": calculate._bulkhead_tool,
    }
    server = BulkheadMCPServer(agent)
    return server, agent


class TestInitialize:
    def test_returns_capabilities(self):
        server, _ = make_server()
        resp = server._handle_request({
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
        })

        assert resp["id"] == 1
        assert resp["result"]["protocolVersion"] == "2024-11-05"
        assert "tools" in resp["result"]["capabilities"]
        assert resp["result"]["serverInfo"]["name"] == "bulkhead"


class TestToolsList:
    def test_returns_all_tools(self):
        server, _ = make_server()
        resp = server._handle_request({
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/list",
        })

        tools = resp["result"]["tools"]
        names = {t["name"] for t in tools}
        assert names == {"search", "calculate"}

    def test_includes_schema(self):
        server, _ = make_server()
        resp = server._handle_request({
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/list",
        })

        tools = resp["result"]["tools"]
        search_tool = next(t for t in tools if t["name"] == "search")
        assert "inputSchema" in search_tool
        assert search_tool["description"] == "Search documents"


class TestToolsCall:
    def test_allow_returns_result(self):
        server, agent = make_server()
        agent.execute_tool.return_value = ToolResult(
            verdict=Verdict.ALLOW,
            result="Found 5 documents",
            action_id="act-mcp-1",
        )

        resp = server._handle_request({
            "jsonrpc": "2.0",
            "id": 4,
            "method": "tools/call",
            "params": {"name": "search", "arguments": {"query": "revenue"}},
        })

        assert resp["id"] == 4
        content = resp["result"]["content"]
        assert len(content) == 1
        assert content[0]["type"] == "text"
        assert content[0]["text"] == "Found 5 documents"
        assert "isError" not in resp["result"]

    def test_deny_returns_error(self):
        server, agent = make_server()
        agent.execute_tool.return_value = ToolResult(
            verdict=Verdict.DENY,
            denial_reason="tool blocked",
        )

        resp = server._handle_request({
            "jsonrpc": "2.0",
            "id": 5,
            "method": "tools/call",
            "params": {"name": "search", "arguments": {"query": "secret"}},
        })

        assert resp["result"]["isError"] is True
        assert "DENIED: tool blocked" in resp["result"]["content"][0]["text"]

    def test_escalate_returns_error(self):
        server, agent = make_server()
        agent.execute_tool.return_value = ToolResult(
            verdict=Verdict.ESCALATE,
            escalation_id="esc-mcp-1",
        )

        resp = server._handle_request({
            "jsonrpc": "2.0",
            "id": 6,
            "method": "tools/call",
            "params": {"name": "search", "arguments": {}},
        })

        assert resp["result"]["isError"] is True
        assert "ESCALATED" in resp["result"]["content"][0]["text"]

    def test_exception_returns_error(self):
        server, agent = make_server()
        agent.execute_tool.side_effect = ValueError("no handler")

        resp = server._handle_request({
            "jsonrpc": "2.0",
            "id": 7,
            "method": "tools/call",
            "params": {"name": "nonexistent", "arguments": {}},
        })

        assert resp["result"]["isError"] is True
        assert "no handler" in resp["result"]["content"][0]["text"]

    def test_none_result_returns_empty_string(self):
        server, agent = make_server()
        agent.execute_tool.return_value = ToolResult(
            verdict=Verdict.ALLOW,
            result=None,
        )

        resp = server._handle_request({
            "jsonrpc": "2.0",
            "id": 8,
            "method": "tools/call",
            "params": {"name": "search", "arguments": {}},
        })

        assert resp["result"]["content"][0]["text"] == ""


class TestErrorHandling:
    def test_unknown_method(self):
        server, _ = make_server()
        resp = server._handle_request({
            "jsonrpc": "2.0",
            "id": 9,
            "method": "unknown/method",
        })

        assert resp["error"]["code"] == -32601
        assert "Method not found" in resp["error"]["message"]

    def test_notification_no_response(self):
        server, _ = make_server()
        resp = server._handle_request({
            "jsonrpc": "2.0",
            "method": "some/notification",
            # No "id" field — this is a notification
        })

        assert resp is None

    def test_parse_error(self):
        server, _ = make_server()
        output = io.StringIO()

        with patch("sys.stdin", io.StringIO("not valid json\n")), \
             patch("sys.stdout", output):
            server.run()

        result = json.loads(output.getvalue().strip())
        assert result["error"]["code"] == -32700


class TestRunLoop:
    def test_processes_multiple_requests(self):
        server, agent = make_server()
        agent.execute_tool.return_value = ToolResult(
            verdict=Verdict.ALLOW, result="ok"
        )

        requests = [
            json.dumps({"jsonrpc": "2.0", "id": 1, "method": "initialize"}),
            json.dumps({"jsonrpc": "2.0", "id": 2, "method": "tools/list"}),
        ]
        stdin = io.StringIO("\n".join(requests) + "\n")
        stdout = io.StringIO()

        with patch("sys.stdin", stdin), patch("sys.stdout", stdout):
            server.run()

        lines = stdout.getvalue().strip().split("\n")
        assert len(lines) == 2
        assert json.loads(lines[0])["id"] == 1
        assert json.loads(lines[1])["id"] == 2

    def test_skips_empty_lines(self):
        server, _ = make_server()
        stdin = io.StringIO("\n\n\n")
        stdout = io.StringIO()

        with patch("sys.stdin", stdin), patch("sys.stdout", stdout):
            server.run()

        assert stdout.getvalue() == ""
