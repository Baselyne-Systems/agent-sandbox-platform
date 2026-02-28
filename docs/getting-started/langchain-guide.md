# LangChain Integration Guide

This guide shows how to wrap Bulkhead tools as LangChain `StructuredTool` instances so that every LangChain tool call goes through Bulkhead's guardrail evaluation, budget checking, and audit trail.

> **See also:** [Agent Developer Guide](agent-guide.md) | [API Reference](../api-reference.md)

---

## Why

LangChain agents call tools autonomously. Without guardrails, a misbehaving agent can execute any tool with any parameters. By wrapping tools through Bulkhead, every invocation is evaluated against your guardrail policy before execution — and the outcome is recorded in the audit trail.

---

## Prerequisites

```bash
pip install bulkhead-sdk langchain-core
```

---

## The Adapter Pattern

The key function is `bulkhead_tool_to_langchain` — it wraps a `@tool`-decorated function into a LangChain `StructuredTool` that routes calls through the Bulkhead SDK:

```python
from typing import Any

from langchain_core.tools import StructuredTool

from bulkhead import BulkheadAgent, Verdict, tool


def bulkhead_tool_to_langchain(
    agent: BulkheadAgent, func: Any
) -> StructuredTool:
    """Wrap a @tool-decorated function as a LangChain StructuredTool.

    Every invocation goes through Bulkhead's evaluate -> execute -> report
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
```

---

## Full Working Example

```python
"""Wrapping Bulkhead tools as LangChain tools."""
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
```

---

## How Verdicts Map to LangChain Responses

| Bulkhead Verdict | LangChain Tool Return |
|------------------|-----------------------|
| `ALLOW` | The tool's actual return value |
| `DENY` | `"DENIED: <reason>"` (string) |
| `ESCALATE` | `"ESCALATED: awaiting human approval (<id>)"` (string) |

Your LangChain agent (or chain) can inspect the string prefix to detect denied/escalated calls and adjust its behavior accordingly.

---

## Limitations

- The adapter is synchronous — it blocks on the gRPC call to the Host Agent. For high-throughput scenarios, consider an async adapter using `aiogrpc`.
- LangChain's `StructuredTool.from_function` infers parameter schemas from the `_run` signature. Since the wrapper uses `**kwargs`, you may want to add explicit parameter schemas for better LLM function-calling accuracy.
- The `_bulkhead_tool` attribute is set by the `@tool` decorator — only decorated functions can be wrapped.

---

## Next Steps

- [Agent Developer Guide](agent-guide.md) — full Python SDK tutorial with `@tool`, verdicts, and human interaction
- [Operator Guide](operator-guide.md) — deploy the stack and configure guardrails
- [API Reference](../api-reference.md) — complete RPC reference
