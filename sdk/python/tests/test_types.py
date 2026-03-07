"""Tests for core SDK types."""

from bulkhead.types import HumanResponse, ToolResult, ToolSchema, Verdict


class TestVerdict:
    def test_enum_values(self):
        assert Verdict.ALLOW.value == "allow"
        assert Verdict.DENY.value == "deny"
        assert Verdict.ESCALATE.value == "escalate"

    def test_all_variants_exist(self):
        assert len(Verdict) == 3


class TestToolResult:
    def test_defaults(self):
        r = ToolResult(verdict=Verdict.ALLOW)
        assert r.result is None
        assert r.denial_reason == ""
        assert r.escalation_id == ""
        assert r.action_id == ""

    def test_deny_result(self):
        r = ToolResult(
            verdict=Verdict.DENY,
            denial_reason="tool blocked by rule deny-shell",
        )
        assert r.verdict == Verdict.DENY
        assert r.denial_reason == "tool blocked by rule deny-shell"

    def test_escalate_result(self):
        r = ToolResult(
            verdict=Verdict.ESCALATE,
            escalation_id="esc-123",
            action_id="act-456",
        )
        assert r.verdict == Verdict.ESCALATE
        assert r.escalation_id == "esc-123"
        assert r.action_id == "act-456"

    def test_allow_with_result(self):
        r = ToolResult(
            verdict=Verdict.ALLOW,
            result={"invoices": [1, 2, 3]},
            action_id="act-789",
        )
        assert r.result == {"invoices": [1, 2, 3]}


class TestToolSchema:
    def test_creation(self):
        s = ToolSchema(
            name="read_file",
            description="Read a file",
            input_schema={"type": "object", "properties": {"path": {"type": "string"}}},
        )
        assert s.name == "read_file"
        assert s.description == "Read a file"
        assert "path" in s.input_schema["properties"]


class TestHumanResponse:
    def test_defaults(self):
        r = HumanResponse()
        assert r.request_id == ""
        assert r.status == "pending"
        assert r.response == ""
        assert r.responder_id == ""

    def test_responded(self):
        r = HumanResponse(
            request_id="req-001",
            status="responded",
            response="approved",
            responder_id="user-42",
        )
        assert r.status == "responded"
        assert r.response == "approved"
