"""Basic example: an invoice-processing agent using the Bulkhead SDK.

This agent registers two tools — read_invoice and call_api — and
executes them through Bulkhead's evaluate → execute → report cycle.
The Runtime evaluates guardrails and budget before each tool call;
tools execute locally inside the container.

Run inside a Bulkhead sandbox (BULKHEAD_ENDPOINT and BULKHEAD_SANDBOX_ID
are injected automatically).
"""

import json

from bulkhead import BulkheadAgent, Verdict, tool


@tool("read_invoice", description="Read a JSON invoice from disk")
def read_invoice(path: str) -> dict:
    with open(path) as f:
        return json.load(f)


@tool("validate_total", description="Validate invoice line items sum to total")
def validate_total(invoice: dict) -> dict:
    line_total = sum(item["amount"] for item in invoice.get("line_items", []))
    expected = invoice.get("total", 0)
    return {
        "valid": abs(line_total - expected) < 0.01,
        "line_total": line_total,
        "expected_total": expected,
    }


def main():
    with BulkheadAgent(tools=[read_invoice, validate_total]) as agent:
        # Report progress
        agent.report_progress("Starting invoice processing", percent_complete=0)

        # Read the invoice — guardrails are evaluated before execution
        result = agent.execute_tool("read_invoice", {"path": "/workspace/inv-001.json"})

        if result.verdict == Verdict.DENY:
            print(f"Denied: {result.denial_reason}")
            return
        if result.verdict == Verdict.ESCALATE:
            print(f"Escalated to human review: {result.escalation_id}")
            return

        invoice = result.result
        print(f"Invoice {invoice.get('id')}: ${invoice.get('total')}")

        # Validate the total
        agent.report_progress("Validating totals", percent_complete=50)
        validation = agent.execute_tool("validate_total", {"invoice": invoice})
        if validation.verdict == Verdict.ALLOW:
            if validation.result["valid"]:
                print("Invoice validated successfully")
            else:
                print(f"Mismatch: line items sum to {validation.result['line_total']}")

        agent.report_progress("Done", percent_complete=100)


if __name__ == "__main__":
    main()
