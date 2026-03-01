"""Tool registration decorator for the Bulkhead SDK."""

import inspect
from dataclasses import dataclass, field
from typing import Any, Callable, get_type_hints


@dataclass
class ToolDefinition:
    """Metadata for a registered tool."""

    name: str
    description: str
    handler: Callable[..., Any]
    input_schema: dict[str, Any] | None = None


def _extract_schema(func: Callable) -> dict[str, Any]:
    """Extract JSON Schema from function signature type hints."""
    sig = inspect.signature(func)
    hints = get_type_hints(func)

    properties: dict[str, Any] = {}
    required: list[str] = []

    type_map = {
        str: {"type": "string"},
        int: {"type": "integer"},
        float: {"type": "number"},
        bool: {"type": "boolean"},
        list: {"type": "array"},
        dict: {"type": "object"},
    }

    for param_name, param in sig.parameters.items():
        hint = hints.get(param_name)
        if hint and hint in type_map:
            properties[param_name] = type_map[hint].copy()
        else:
            properties[param_name] = {"type": "string"}  # fallback

        if param.default is inspect.Parameter.empty:
            required.append(param_name)

    schema: dict[str, Any] = {
        "type": "object",
        "properties": properties,
    }
    if required:
        schema["required"] = required
    return schema


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
        schema = _extract_schema(func)
        func._bulkhead_tool = ToolDefinition(  # type: ignore[attr-defined]
            name=name,
            description=description or func.__doc__ or "",
            handler=func,
            input_schema=schema,
        )
        return func

    return decorator
