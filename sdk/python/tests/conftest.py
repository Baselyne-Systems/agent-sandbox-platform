"""Shared fixtures for SDK tests."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any
from unittest.mock import MagicMock

import pytest

from bulkhead.decorators import tool
from bulkhead.client import BulkheadAgent


# --- Mock proto module ---

@dataclass
class _FakeExecuteToolResponse:
    verdict: int = 0
    denial_reason: str = ""
    escalation_id: str = ""
    action_id: str = "act-test-001"
    result: Any = None


@dataclass
class _FakeRequestHumanInputResponse:
    request_id: str = "req-test-001"
    response: str = ""
    responder_id: str = ""
    timed_out: bool = False


@dataclass
class _FakeCheckHumanRequestResponse:
    status: str = "pending"
    response: str = ""
    responder_id: str = ""


class FakeProtoModule:
    """Simulates the generated runtime_pb2 module."""

    ACTION_VERDICT_ALLOW = 0
    ACTION_VERDICT_DENY = 1
    ACTION_VERDICT_ESCALATE = 2

    @staticmethod
    def ExecuteToolRequest(**kwargs):
        return kwargs

    @staticmethod
    def ReportActionResultRequest(**kwargs):
        return kwargs

    @staticmethod
    def RequestHumanInputRequest(**kwargs):
        return kwargs

    @staticmethod
    def CheckHumanRequestRequest(**kwargs):
        return kwargs

    @staticmethod
    def ReportProgressRequest(**kwargs):
        return kwargs

    @staticmethod
    def ListToolsRequest(**kwargs):
        return kwargs


# --- Sample tools ---

@tool("read_file", description="Read a file from disk")
def sample_read_file(path: str) -> dict:
    return {"content": f"data from {path}"}


@tool("failing_tool", description="Always raises")
def sample_failing_tool(x: str) -> str:
    raise RuntimeError("tool exploded")


@tool("add_numbers", description="Add two numbers")
def sample_add(a: int, b: int) -> int:
    return a + b


# --- Fixtures ---

@pytest.fixture
def fake_pb2():
    return FakeProtoModule()


@pytest.fixture
def mock_stub():
    return MagicMock()


@pytest.fixture
def connected_agent(mock_stub, fake_pb2):
    """A BulkheadAgent with mocked gRPC stub, ready to use."""
    agent = BulkheadAgent(
        tools=[sample_read_file, sample_failing_tool, sample_add],
        endpoint="localhost:50052",
        sandbox_id="sbx-test-001",
    )
    agent._channel = MagicMock()
    agent._stub = mock_stub
    agent._pb2 = fake_pb2
    return agent


def make_allow_response(action_id: str = "act-001") -> _FakeExecuteToolResponse:
    return _FakeExecuteToolResponse(
        verdict=FakeProtoModule.ACTION_VERDICT_ALLOW,
        action_id=action_id,
    )


def make_deny_response(
    reason: str = "blocked by guardrail",
    action_id: str = "act-002",
) -> _FakeExecuteToolResponse:
    return _FakeExecuteToolResponse(
        verdict=FakeProtoModule.ACTION_VERDICT_DENY,
        denial_reason=reason,
        action_id=action_id,
    )


def make_escalate_response(
    escalation_id: str = "esc-001",
    action_id: str = "act-003",
) -> _FakeExecuteToolResponse:
    return _FakeExecuteToolResponse(
        verdict=FakeProtoModule.ACTION_VERDICT_ESCALATE,
        escalation_id=escalation_id,
        action_id=action_id,
    )
