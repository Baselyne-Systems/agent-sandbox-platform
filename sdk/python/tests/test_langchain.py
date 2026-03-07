"""Tests for LangChain integration wrapper."""

from __future__ import annotations

from unittest.mock import MagicMock, patch

import pytest

from bulkhead.types import ToolResult, Verdict


# Skip all tests if langchain is not installed
langchain = pytest.importorskip("langchain_core")

from langchain_core.tools import tool as lc_tool, BaseTool

from bulkhead.langchain import wrap_langchain_tool, wrap_langchain_tools


@lc_tool
def search_docs(query: str) -> str:
    """Search internal documents."""
    return f"Results for {query}"


@lc_tool
def send_email(to: str, body: str) -> str:
    """Send an email."""
    return f"Sent to {to}"


@lc_tool
def failing_lc_tool(x: str) -> str:
    """Always fails."""
    raise RuntimeError("LC tool exploded")


class TestWrapLangchainTool:
    def test_allow_executes_and_reports(self):
        agent = MagicMock()
        agent.evaluate.return_value = ToolResult(
            verdict=Verdict.ALLOW, action_id="act-lc-1"
        )

        wrapped = wrap_langchain_tool(agent, search_docs)
        result = wrapped.invoke({"query": "quarterly revenue"})

        assert result == "Results for quarterly revenue"
        agent.evaluate.assert_called_once_with("search_docs", {"query": "quarterly revenue"})
        agent.report_result.assert_called_once_with(
            "act-lc-1", success=True, result="Results for quarterly revenue"
        )

    def test_deny_returns_string(self):
        agent = MagicMock()
        agent.evaluate.return_value = ToolResult(
            verdict=Verdict.DENY, denial_reason="shell not allowed"
        )

        wrapped = wrap_langchain_tool(agent, search_docs)
        result = wrapped.invoke({"query": "test"})

        assert result == "DENIED: shell not allowed"
        agent.report_result.assert_not_called()

    def test_escalate_returns_string(self):
        agent = MagicMock()
        agent.evaluate.return_value = ToolResult(
            verdict=Verdict.ESCALATE, escalation_id="esc-lc-1"
        )

        wrapped = wrap_langchain_tool(agent, send_email)
        result = wrapped.invoke({"to": "boss@co.com", "body": "hi"})

        assert result == "ESCALATED: awaiting human approval (esc-lc-1)"
        agent.report_result.assert_not_called()

    def test_exception_reports_failure_and_reraises(self):
        agent = MagicMock()
        agent.evaluate.return_value = ToolResult(
            verdict=Verdict.ALLOW, action_id="act-lc-err"
        )

        wrapped = wrap_langchain_tool(agent, failing_lc_tool)
        with pytest.raises(RuntimeError, match="LC tool exploded"):
            wrapped.invoke({"x": "boom"})

        agent.report_result.assert_called_once_with(
            "act-lc-err", success=False, error="LC tool exploded"
        )

    def test_preserves_name_and_description(self):
        agent = MagicMock()
        wrapped = wrap_langchain_tool(agent, search_docs)
        assert wrapped.name == "search_docs"
        assert wrapped.description == "Search internal documents."


class TestWrapLangchainTools:
    def test_wraps_multiple_tools(self):
        agent = MagicMock()
        wrapped = wrap_langchain_tools(agent, [search_docs, send_email])

        assert len(wrapped) == 2
        assert wrapped[0].name == "search_docs"
        assert wrapped[1].name == "send_email"

    def test_empty_list(self):
        agent = MagicMock()
        wrapped = wrap_langchain_tools(agent, [])
        assert wrapped == []
