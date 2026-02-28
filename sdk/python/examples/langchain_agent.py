"""Example: wrapping Bulkhead tools as LangChain tools.

Shows how to integrate the Bulkhead SDK with LangChain by wrapping
@tool-decorated functions so that every LangChain tool call goes through
Bulkhead's guardrail evaluation.

Requires: pip install langchain-core
"""

from __future__ import annotations

from typing import Any

from langchain_core.tools import StructuredTool

from bulkhead import BulkheadAgent, Verdict, tool


@tool("search_docs", description="Search internal documents")
def search_docs(query: str, max_results: int = 5) -> dict:
    # Your real search implementation here
    return {"results": [f"doc about {query}"], "count": 1}


@tool("send_email", description="Send an email (requires guardrail approval)")
def send_email(to: str, subject: str, body: str) -> dict:
    # Your real email implementation here
    return {"sent": True, "to": to}


def bulkhead_tool_to_langchain(
    agent: BulkheadAgent, func: Any
) -> StructuredTool:
    """Wrap a @tool-decorated function as a LangChain StructuredTool.

    Every invocation goes through Bulkhead's evaluate → execute → report
    cycle, so guardrails and budget are enforced transparently.
    """
    defn = func._bulkhead_tool

    def _run(**kwargs: Any) -> Any:
        result = agent.execute_tool(defn.name, kwargs)
        if result.verdict == Verdict.DENY:
            return f"DENIED: {result.denial_reason}"
        if result.verdict == Verdict.ESCALATE:
            return f"ESCALATED: awaiting human approval ({result.escalation_id})"
        return result.result

    return StructuredTool.from_function(
        func=_run,
        name=defn.name,
        description=defn.description,
    )


def main():
    with BulkheadAgent(tools=[search_docs, send_email]) as agent:
        # Create LangChain tools that are guardrail-aware
        lc_search = bulkhead_tool_to_langchain(agent, search_docs)
        lc_email = bulkhead_tool_to_langchain(agent, send_email)

        # Use them like any LangChain tool
        result = lc_search.invoke({"query": "Q4 revenue", "max_results": 3})
        print(f"Search result: {result}")

        result = lc_email.invoke({
            "to": "finance@example.com",
            "subject": "Q4 Report",
            "body": "Please review attached.",
        })
        print(f"Email result: {result}")


if __name__ == "__main__":
    main()
