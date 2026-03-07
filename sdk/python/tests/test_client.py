"""Tests for BulkheadAgent client."""

from __future__ import annotations

from unittest.mock import MagicMock, call

import pytest

from bulkhead.client import BulkheadAgent
from bulkhead.decorators import tool
from bulkhead.types import Verdict

from .conftest import (
    make_allow_response,
    make_deny_response,
    make_escalate_response,
    sample_read_file,
    sample_failing_tool,
    sample_add,
    _FakeRequestHumanInputResponse,
    _FakeCheckHumanRequestResponse,
)


class TestBulkheadAgentInit:
    def test_registers_decorated_tools(self):
        agent = BulkheadAgent(tools=[sample_read_file, sample_add])
        assert "read_file" in agent._tools
        assert "add_numbers" in agent._tools

    def test_rejects_undecorated_function(self):
        def plain_func(x: str) -> str:
            return x

        with pytest.raises(ValueError, match="not decorated with @tool"):
            BulkheadAgent(tools=[plain_func])

    def test_env_var_defaults(self, monkeypatch):
        monkeypatch.setenv("BULKHEAD_ENDPOINT", "custom:9999")
        monkeypatch.setenv("BULKHEAD_SANDBOX_ID", "sbx-from-env")
        agent = BulkheadAgent()
        assert agent._endpoint == "custom:9999"
        assert agent._sandbox_id == "sbx-from-env"

    def test_constructor_overrides_env(self, monkeypatch):
        monkeypatch.setenv("BULKHEAD_ENDPOINT", "env:1234")
        agent = BulkheadAgent(endpoint="explicit:5678", sandbox_id="sbx-explicit")
        assert agent._endpoint == "explicit:5678"
        assert agent._sandbox_id == "sbx-explicit"

    def test_default_endpoint(self, monkeypatch):
        monkeypatch.delenv("BULKHEAD_ENDPOINT", raising=False)
        monkeypatch.delenv("BULKHEAD_SANDBOX_ID", raising=False)
        agent = BulkheadAgent()
        assert agent._endpoint == "localhost:50052"
        assert agent._sandbox_id == ""


class TestEvaluate:
    def test_not_connected_raises(self):
        agent = BulkheadAgent(tools=[sample_read_file])
        with pytest.raises(RuntimeError, match="Not connected"):
            agent.evaluate("read_file")

    def test_allow_verdict(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_allow_response("act-100")
        result = connected_agent.evaluate("read_file", {"path": "/data/test.txt"})

        assert result.verdict == Verdict.ALLOW
        assert result.action_id == "act-100"
        assert result.denial_reason == ""
        mock_stub.ExecuteTool.assert_called_once()

    def test_deny_verdict(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_deny_response("shell blocked")
        result = connected_agent.evaluate("shell", {"cmd": "rm -rf /"})

        assert result.verdict == Verdict.DENY
        assert result.denial_reason == "shell blocked"

    def test_escalate_verdict(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_escalate_response("esc-99")
        result = connected_agent.evaluate("deploy", {"env": "production"})

        assert result.verdict == Verdict.ESCALATE
        assert result.escalation_id == "esc-99"

    def test_sends_sandbox_id_metadata(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_allow_response()
        connected_agent.evaluate("read_file")

        _, kwargs = mock_stub.ExecuteTool.call_args
        assert kwargs["metadata"] == [("x-sandbox-id", "sbx-test-001")]


class TestExecuteTool:
    def test_allow_executes_handler(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_allow_response("act-200")
        result = connected_agent.execute_tool("read_file", {"path": "/data/inv.json"})

        assert result.verdict == Verdict.ALLOW
        assert result.result == {"content": "data from /data/inv.json"}
        assert result.action_id == "act-200"
        mock_stub.ReportActionResult.assert_called_once()

    def test_deny_skips_handler(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_deny_response("not allowed")
        result = connected_agent.execute_tool("read_file", {"path": "/etc/shadow"})

        assert result.verdict == Verdict.DENY
        assert result.result is None
        mock_stub.ReportActionResult.assert_not_called()

    def test_escalate_skips_handler(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_escalate_response("esc-50")
        result = connected_agent.execute_tool("read_file", {"path": "/data/test.txt"})

        assert result.verdict == Verdict.ESCALATE
        assert result.result is None
        mock_stub.ReportActionResult.assert_not_called()

    def test_handler_exception_reports_failure(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_allow_response("act-300")

        with pytest.raises(RuntimeError, match="tool exploded"):
            connected_agent.execute_tool("failing_tool", {"x": "boom"})

        mock_stub.ReportActionResult.assert_called_once()

    def test_unregistered_tool_raises(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_allow_response()

        with pytest.raises(ValueError, match="No handler registered for tool 'nonexistent'"):
            connected_agent.execute_tool("nonexistent")

    def test_report_result_fire_and_forget(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_allow_response()
        mock_stub.ReportActionResult.side_effect = Exception("gRPC down")

        # Should not raise — report_result swallows errors
        result = connected_agent.execute_tool("read_file", {"path": "/ok"})
        assert result.verdict == Verdict.ALLOW

    def test_handler_with_multiple_params(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_allow_response()
        result = connected_agent.execute_tool("add_numbers", {"a": 3, "b": 7})

        assert result.verdict == Verdict.ALLOW
        assert result.result == 10

    def test_none_params_defaults_to_empty(self, connected_agent, mock_stub):
        mock_stub.ExecuteTool.return_value = make_allow_response()
        # read_file handler expects 'path' kwarg — will fail, but evaluate should work
        with pytest.raises(TypeError):
            connected_agent.execute_tool("read_file", None)
        # evaluate was called (ALLOW), handler failed, report_result called with failure
        mock_stub.ExecuteTool.assert_called_once()


class TestReportResult:
    def test_calls_rpc(self, connected_agent, mock_stub):
        connected_agent.report_result("act-500", success=True, result={"data": "ok"})
        mock_stub.ReportActionResult.assert_called_once()

    def test_swallows_errors(self, connected_agent, mock_stub):
        mock_stub.ReportActionResult.side_effect = Exception("network error")
        # Should not raise
        connected_agent.report_result("act-500", success=False, error="something broke")


class TestHumanInteraction:
    def test_request_human_input(self, connected_agent, mock_stub):
        mock_stub.RequestHumanInput.return_value = _FakeRequestHumanInputResponse(
            request_id="req-777"
        )
        resp = connected_agent.request_human_input(
            question="Deploy to production?",
            options=["yes", "no"],
            context="Agent wants to deploy v2.1",
            timeout_seconds=600,
        )
        assert resp.request_id == "req-777"
        assert resp.status == "pending"
        mock_stub.RequestHumanInput.assert_called_once()

    def test_request_not_connected_raises(self):
        agent = BulkheadAgent()
        with pytest.raises(RuntimeError, match="Not connected"):
            agent.request_human_input("question?")

    def test_check_human_request_pending(self, connected_agent, mock_stub):
        mock_stub.CheckHumanRequest.return_value = _FakeCheckHumanRequestResponse(
            status="pending"
        )
        resp = connected_agent.check_human_request("req-777")
        assert resp.status == "pending"
        assert resp.response == ""

    def test_check_human_request_responded(self, connected_agent, mock_stub):
        mock_stub.CheckHumanRequest.return_value = _FakeCheckHumanRequestResponse(
            status="responded",
            response="approved",
            responder_id="admin-1",
        )
        resp = connected_agent.check_human_request("req-777")
        assert resp.status == "responded"
        assert resp.response == "approved"
        assert resp.responder_id == "admin-1"

    def test_check_not_connected_raises(self):
        agent = BulkheadAgent()
        with pytest.raises(RuntimeError, match="Not connected"):
            agent.check_human_request("req-000")


class TestContextManager:
    def test_context_manager_closes(self, monkeypatch):
        """Verify __exit__ closes the channel."""
        agent = BulkheadAgent(
            tools=[sample_read_file],
            endpoint="localhost:50052",
            sandbox_id="test",
        )
        mock_channel = MagicMock()
        agent._channel = mock_channel
        agent._stub = MagicMock()

        agent.close()
        mock_channel.close.assert_called_once()
        assert agent._channel is None


class TestListTools:
    def test_returns_local_tools_when_rpc_fails(self, connected_agent, mock_stub):
        mock_stub.ListTools.side_effect = Exception("not implemented")
        tools = connected_agent.list_tools()

        names = {t.name for t in tools}
        assert "read_file" in names
        assert "failing_tool" in names
        assert "add_numbers" in names

    def test_returns_local_tools_when_not_connected(self):
        agent = BulkheadAgent(tools=[sample_read_file])
        tools = agent.list_tools()
        assert len(tools) == 1
        assert tools[0].name == "read_file"
