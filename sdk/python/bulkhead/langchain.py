"""LangChain integration — wrap existing LangChain tools with Bulkhead guardrails.

Usage::

    from langchain_core.tools import tool

    @tool
    def search_docs(query: str) -> str:
        \"\"\"Search internal documents.\"\"\"
        return f"Results for {query}"

    from bulkhead import BulkheadAgent
    from bulkhead.langchain import wrap_langchain_tools

    with BulkheadAgent() as agent:
        guarded = wrap_langchain_tools(agent, [search_docs])
        # Use guarded tools in your LangChain agent
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

from .types import Verdict

if TYPE_CHECKING:
    from langchain_core.tools import BaseTool

    from .client import BulkheadAgent


def wrap_langchain_tool(agent: BulkheadAgent, lc_tool: BaseTool) -> BaseTool:
    """Wrap a LangChain tool with Bulkhead guardrail enforcement.

    Every invocation goes through the evaluate → execute → report cycle:

    1. Calls the Host Agent to evaluate guardrails and budget.
    2. If ALLOW, delegates to the original tool.
    3. Reports the result for the audit trail.

    DENY returns ``"DENIED: <reason>"``. ESCALATE returns
    ``"ESCALATED: awaiting human approval (<id>)"``. The LangChain agent
    sees these as string tool outputs and can react accordingly.

    Args:
        agent: A connected :class:`~bulkhead.BulkheadAgent` instance.
        lc_tool: Any LangChain ``BaseTool`` (``@tool``, ``StructuredTool``, etc.).

    Returns:
        A new LangChain tool with the same name, description, and schema.
    """
    from langchain_core.tools import StructuredTool

    original_run = lc_tool._run

    def _guarded_run(**kwargs: Any) -> Any:
        evaluation = agent.evaluate(lc_tool.name, kwargs)

        if evaluation.verdict == Verdict.DENY:
            return f"DENIED: {evaluation.denial_reason}"

        if evaluation.verdict == Verdict.ESCALATE:
            return f"ESCALATED: awaiting human approval ({evaluation.escalation_id})"

        try:
            output = original_run(**kwargs)
        except Exception as exc:
            agent.report_result(evaluation.action_id, success=False, error=str(exc))
            raise

        agent.report_result(evaluation.action_id, success=True, result=output)
        return output

    return StructuredTool.from_function(
        func=_guarded_run,
        name=lc_tool.name,
        description=lc_tool.description,
        args_schema=getattr(lc_tool, "args_schema", None),
    )


def wrap_langchain_tools(agent: BulkheadAgent, tools: list[BaseTool]) -> list[BaseTool]:
    """Wrap multiple LangChain tools with Bulkhead guardrail enforcement.

    Args:
        agent: A connected :class:`~bulkhead.BulkheadAgent` instance.
        tools: List of LangChain tools to wrap.

    Returns:
        List of guardrail-enforced LangChain tools.
    """
    return [wrap_langchain_tool(agent, t) for t in tools]
