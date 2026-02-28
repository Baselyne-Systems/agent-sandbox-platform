"""BulkheadAgent — main SDK client for the Bulkhead agent sandbox platform.

The client transparently handles the evaluate → execute → report cycle:

1. Calls ``ExecuteTool`` on the Runtime to evaluate guardrails and budget.
2. If the verdict is ALLOW, executes the tool locally via the registered handler.
3. Calls ``ReportActionResult`` to record the outcome for auditing.
"""

from __future__ import annotations

import os
from typing import Any, Callable, Sequence

import grpc

from .decorators import ToolDefinition
from .types import HumanResponse, ToolResult, Verdict


class BulkheadAgent:
    """Client for interacting with the Bulkhead Runtime from inside a sandbox.

    Automatically discovers the Runtime endpoint and sandbox ID from
    environment variables (``BULKHEAD_ENDPOINT``, ``BULKHEAD_SANDBOX_ID``).

    Args:
        tools: List of ``@tool``-decorated functions to register.
        endpoint: Override the Runtime gRPC endpoint.
        sandbox_id: Override the sandbox ID.
    """

    def __init__(
        self,
        tools: Sequence[Callable] = (),
        endpoint: str | None = None,
        sandbox_id: str | None = None,
    ):
        self._endpoint = endpoint or os.environ.get("BULKHEAD_ENDPOINT", "localhost:50052")
        self._sandbox_id = sandbox_id or os.environ.get("BULKHEAD_SANDBOX_ID", "")
        self._metadata = [("x-sandbox-id", self._sandbox_id)]
        self._channel: grpc.Channel | None = None
        self._stub: Any = None

        # Build tool registry from @tool-decorated functions
        self._tools: dict[str, ToolDefinition] = {}
        for func in tools:
            defn = getattr(func, "_bulkhead_tool", None)
            if defn is None:
                raise ValueError(
                    f"Function {func.__name__} is not decorated with @tool"
                )
            self._tools[defn.name] = defn

    def __enter__(self) -> BulkheadAgent:
        self.connect()
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()

    def connect(self) -> None:
        """Establish gRPC connection to the Runtime."""
        self._channel = grpc.insecure_channel(self._endpoint)

        # Import generated stubs — these are optional and the SDK degrades
        # gracefully if codegen hasn't been run yet.
        try:
            from ._generated.platform.runtime.v1 import runtime_pb2, runtime_pb2_grpc

            self._stub = runtime_pb2_grpc.AgentAPIServiceStub(self._channel)
            self._pb2 = runtime_pb2
        except ImportError:
            raise ImportError(
                "gRPC stubs not found. Run proto codegen or install bulkhead-sdk "
                "with generated stubs."
            )

    def close(self) -> None:
        """Close the gRPC channel."""
        if self._channel is not None:
            self._channel.close()
            self._channel = None

    def execute_tool(self, name: str, params: dict[str, Any] | None = None) -> ToolResult:
        """Execute a tool through the evaluate → execute → report cycle.

        1. Calls ExecuteTool RPC to evaluate guardrails/budget.
        2. If ALLOW, runs the locally registered handler.
        3. Calls ReportActionResult to record the outcome.

        Args:
            name: Name of the tool to execute.
            params: Parameters to pass to the tool handler.

        Returns:
            ToolResult with the verdict and (if allowed) the handler's output.
        """
        if self._stub is None:
            raise RuntimeError("Not connected — call connect() or use as context manager")

        params = params or {}

        # 1. Evaluate via gRPC
        parameters_struct = _dict_to_struct(params, self._pb2)
        request = self._pb2.ExecuteToolRequest(
            tool_name=name,
            parameters=parameters_struct,
            justification="",
        )
        resp = self._stub.ExecuteTool(request, metadata=self._metadata)

        # Map proto verdict enum
        verdict_map = {
            self._pb2.ACTION_VERDICT_ALLOW: Verdict.ALLOW,
            self._pb2.ACTION_VERDICT_DENY: Verdict.DENY,
            self._pb2.ACTION_VERDICT_ESCALATE: Verdict.ESCALATE,
        }
        verdict = verdict_map.get(resp.verdict, Verdict.DENY)

        if verdict == Verdict.DENY:
            return ToolResult(
                verdict=Verdict.DENY,
                denial_reason=resp.denial_reason,
                action_id=resp.action_id,
            )

        if verdict == Verdict.ESCALATE:
            return ToolResult(
                verdict=Verdict.ESCALATE,
                escalation_id=resp.escalation_id,
                action_id=resp.action_id,
            )

        # 2. Execute locally via registered handler
        handler_def = self._tools.get(name)
        if handler_def is None:
            # Report failure — no handler registered
            self._report_result(resp.action_id, success=False, error="no handler registered")
            raise ValueError(f"No handler registered for tool '{name}'")

        try:
            result = handler_def.handler(**params)
        except Exception as exc:
            # 3a. Report failure
            self._report_result(resp.action_id, success=False, error=str(exc))
            raise

        # 3b. Report success
        self._report_result(resp.action_id, success=True, result=result)

        return ToolResult(
            verdict=Verdict.ALLOW,
            result=result,
            action_id=resp.action_id,
        )

    def request_human_input(
        self,
        question: str,
        options: list[str] | None = None,
        context: str = "",
        timeout_seconds: int = 300,
    ) -> HumanResponse:
        """Submit a question to a human and return the request ID for polling.

        Args:
            question: The question to ask.
            options: Optional list of answer choices.
            context: Additional context for the human reviewer.
            timeout_seconds: How long to wait for a response.
        """
        if self._stub is None:
            raise RuntimeError("Not connected")

        request = self._pb2.RequestHumanInputRequest(
            question=question,
            options=options or [],
            context=context,
            timeout_seconds=timeout_seconds,
        )
        resp = self._stub.RequestHumanInput(request, metadata=self._metadata)
        return HumanResponse(
            request_id=resp.request_id,
            status="pending",
            response=resp.response,
            responder_id=resp.responder_id,
        )

    def check_human_request(self, request_id: str) -> HumanResponse:
        """Poll the status of a human interaction request.

        Args:
            request_id: The ID returned by :meth:`request_human_input`.
        """
        if self._stub is None:
            raise RuntimeError("Not connected")

        request = self._pb2.CheckHumanRequestRequest(request_id=request_id)
        resp = self._stub.CheckHumanRequest(request, metadata=self._metadata)
        return HumanResponse(
            request_id=request_id,
            status=resp.status,
            response=resp.response,
            responder_id=resp.responder_id,
        )

    def report_progress(
        self, message: str, percent_complete: float = 0.0, metadata: dict[str, str] | None = None
    ) -> None:
        """Report task progress to the platform.

        Args:
            message: Progress description.
            percent_complete: Estimated completion percentage (0-100).
            metadata: Optional key-value metadata.
        """
        if self._stub is None:
            raise RuntimeError("Not connected")

        request = self._pb2.ReportProgressRequest(
            message=message,
            percent_complete=percent_complete,
            metadata=metadata or {},
        )
        self._stub.ReportProgress(request, metadata=self._metadata)

    def _report_result(
        self,
        action_id: str,
        success: bool,
        result: Any = None,
        error: str = "",
    ) -> None:
        """Fire-and-forget: report action result to the Runtime for auditing."""
        try:
            result_struct = None
            if result is not None:
                if isinstance(result, dict):
                    result_struct = _dict_to_struct(result, self._pb2)
                else:
                    result_struct = _dict_to_struct({"value": result}, self._pb2)

            request = self._pb2.ReportActionResultRequest(
                action_id=action_id,
                success=success,
                result=result_struct,
                error_message=error,
            )
            self._stub.ReportActionResult(request, metadata=self._metadata)
        except Exception:
            pass  # fire-and-forget


def _dict_to_struct(d: dict[str, Any], pb2_module: Any) -> Any:
    """Convert a Python dict to a google.protobuf.Struct."""
    from google.protobuf import struct_pb2

    s = struct_pb2.Struct()
    s.update(d)
    return s
