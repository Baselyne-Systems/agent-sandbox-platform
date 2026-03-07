"""Minimal test agent that exercises the SDK end-to-end.

Reads BULKHEAD_ENDPOINT and BULKHEAD_SANDBOX_ID from the environment,
makes a few tool calls, and prints results to stdout.
"""

import json
import sys

from bulkhead import BulkheadAgent, tool


@tool("read_file", description="Read a file from disk")
def read_file(path: str) -> dict:
    return {"content": f"data from {path}"}


@tool("shell", description="Execute a shell command")
def shell(cmd: str) -> str:
    return f"executed: {cmd}"


def main():
    results = []

    with BulkheadAgent(tools=[read_file, shell]) as agent:
        # 1. Allowed tool call.
        r1 = agent.execute_tool("read_file", {"path": "/data/test.txt"})
        results.append({
            "tool": "read_file",
            "verdict": r1.verdict.value,
            "result": str(r1.result) if r1.result else None,
        })

        # 2. Report result for allowed call.
        if r1.action_id:
            agent.report_result(r1.action_id, success=True, result=r1.result)

        # 3. Tool call that may be denied by guardrails.
        r2 = agent.execute_tool("shell", {"cmd": "rm -rf /"})
        results.append({
            "tool": "shell",
            "verdict": r2.verdict.value,
            "result": str(r2.result) if r2.result else None,
        })

    # Output results as JSON for test verification.
    print(json.dumps({"status": "ok", "results": results}))
    return 0


if __name__ == "__main__":
    sys.exit(main())
