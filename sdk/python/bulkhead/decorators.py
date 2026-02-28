"""Tool registration decorator for the Bulkhead SDK."""

from dataclasses import dataclass
from typing import Any, Callable


@dataclass
class ToolDefinition:
    """Metadata for a registered tool."""

    name: str
    description: str
    handler: Callable[..., Any]


def tool(name: str, description: str = ""):
    """Decorator to register a function as a Bulkhead tool.

    Usage::

        @tool("read_invoice", description="Read an invoice file")
        def read_invoice(path: str) -> dict:
            with open(path) as f:
                return json.load(f)

    The decorated function retains its original behavior but gains a
    ``_bulkhead_tool`` attribute containing a :class:`ToolDefinition`.
    """

    def decorator(func: Callable) -> Callable:
        func._bulkhead_tool = ToolDefinition(  # type: ignore[attr-defined]
            name=name,
            description=description or func.__doc__ or "",
            handler=func,
        )
        return func

    return decorator
