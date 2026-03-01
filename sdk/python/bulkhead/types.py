"""Core types for the Bulkhead SDK."""

from dataclasses import dataclass, field
from enum import Enum
from typing import Any


class Verdict(Enum):
    """Result of a guardrail evaluation."""

    ALLOW = "allow"
    DENY = "deny"
    ESCALATE = "escalate"


@dataclass
class ToolResult:
    """Result of a tool execution through the Bulkhead evaluate-execute-report cycle."""

    verdict: Verdict
    result: Any = None
    denial_reason: str = ""
    escalation_id: str = ""
    action_id: str = ""


@dataclass
class ToolSchema:
    """Schema for a tool, compatible with MCP tool definitions."""

    name: str
    description: str
    input_schema: dict[str, Any]


@dataclass
class HumanResponse:
    """Response from a human interaction request."""

    request_id: str = ""
    status: str = "pending"
    response: str = ""
    responder_id: str = ""
