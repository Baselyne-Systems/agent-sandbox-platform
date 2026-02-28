"""Example: bringing existing LangChain tools to Bulkhead.

Shows how to wrap standard LangChain tools with Bulkhead guardrail enforcement.
No need to rewrite tools with Bulkhead's @tool decorator — just wrap them.

Requires: pip install bulkhead-sdk langchain-core
"""

from __future__ import annotations

from langchain_core.tools import tool

from bulkhead import BulkheadAgent
from bulkhead.langchain import wrap_langchain_tools


# 1. Define tools with standard LangChain decorators (no Bulkhead dependency)

@tool
def search_docs(query: str, max_results: int = 5) -> dict:
    """Search internal documents."""
    return {"results": [f"doc about {query}"], "count": 1}


@tool
def send_email(to: str, subject: str, body: str) -> dict:
    """Send an email."""
    return {"sent": True, "to": to}


def main():
    # 2. Wrap them with Bulkhead guardrail enforcement
    with BulkheadAgent() as agent:
        guarded_tools = wrap_langchain_tools(agent, [search_docs, send_email])

        # 3. Use them like any LangChain tool — guardrails are enforced transparently
        result = guarded_tools[0].invoke({"query": "Q4 revenue", "max_results": 3})
        print(f"Search result: {result}")

        result = guarded_tools[1].invoke({
            "to": "finance@example.com",
            "subject": "Q4 Report",
            "body": "Please review attached.",
        })
        print(f"Email result: {result}")

        # guarded_tools can be passed to any LangChain agent:
        # agent_executor = create_react_agent(llm, guarded_tools)
        # agent_executor.invoke({"input": "Find Q4 revenue docs"})


if __name__ == "__main__":
    main()
