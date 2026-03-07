"""Tests for the @tool decorator and schema extraction."""

import pytest

from bulkhead.decorators import ToolDefinition, _extract_schema, tool


class TestToolDecorator:
    def test_attaches_tool_definition(self):
        @tool("my_tool", description="Does stuff")
        def my_func(x: str) -> str:
            return x

        assert hasattr(my_func, "_bulkhead_tool")
        defn = my_func._bulkhead_tool
        assert isinstance(defn, ToolDefinition)
        assert defn.name == "my_tool"
        assert defn.description == "Does stuff"
        assert defn.handler is my_func

    def test_preserves_original_behavior(self):
        @tool("echo")
        def echo(msg: str) -> str:
            return msg

        assert echo("hello") == "hello"

    def test_uses_docstring_as_description_fallback(self):
        @tool("documented")
        def documented(x: str) -> str:
            """This is the docstring."""
            return x

        assert documented._bulkhead_tool.description == "This is the docstring."

    def test_empty_description_and_no_docstring(self):
        @tool("bare")
        def bare(x: str) -> str:
            return x

        assert bare._bulkhead_tool.description == ""

    def test_schema_attached(self):
        @tool("with_schema", description="test")
        def func(path: str, count: int) -> dict:
            return {}

        schema = func._bulkhead_tool.input_schema
        assert schema is not None
        assert schema["type"] == "object"
        assert "path" in schema["properties"]
        assert "count" in schema["properties"]


class TestSchemaExtraction:
    def test_string_type(self):
        def f(name: str):
            pass

        schema = _extract_schema(f)
        assert schema["properties"]["name"] == {"type": "string"}

    def test_int_type(self):
        def f(count: int):
            pass

        schema = _extract_schema(f)
        assert schema["properties"]["count"] == {"type": "integer"}

    def test_float_type(self):
        def f(price: float):
            pass

        schema = _extract_schema(f)
        assert schema["properties"]["price"] == {"type": "number"}

    def test_bool_type(self):
        def f(active: bool):
            pass

        schema = _extract_schema(f)
        assert schema["properties"]["active"] == {"type": "boolean"}

    def test_list_type(self):
        def f(items: list):
            pass

        schema = _extract_schema(f)
        assert schema["properties"]["items"] == {"type": "array"}

    def test_dict_type(self):
        def f(metadata: dict):
            pass

        schema = _extract_schema(f)
        assert schema["properties"]["metadata"] == {"type": "object"}

    def test_unknown_type_defaults_to_string(self):
        class Custom:
            pass

        def f(x: Custom):
            pass

        schema = _extract_schema(f)
        assert schema["properties"]["x"] == {"type": "string"}

    def test_no_annotation_defaults_to_string(self):
        def f(x):
            pass

        schema = _extract_schema(f)
        assert schema["properties"]["x"] == {"type": "string"}

    def test_required_params(self):
        def f(required_arg: str, optional_arg: str = "default"):
            pass

        schema = _extract_schema(f)
        assert "required" in schema
        assert "required_arg" in schema["required"]
        assert "optional_arg" not in schema["required"]

    def test_no_required_when_all_optional(self):
        def f(x: str = "a", y: int = 0):
            pass

        schema = _extract_schema(f)
        assert "required" not in schema

    def test_multiple_params(self):
        def f(path: str, count: int, verbose: bool = False):
            pass

        schema = _extract_schema(f)
        assert len(schema["properties"]) == 3
        assert schema["required"] == ["path", "count"]
