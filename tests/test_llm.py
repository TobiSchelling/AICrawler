"""Tests for the LLM provider module."""

from src.llm import parse_json_response


def test_parse_json_response_plain():
    """Test parsing plain JSON."""
    result = parse_json_response('{"key": "value", "num": 42}')
    assert result == {"key": "value", "num": 42}


def test_parse_json_response_with_code_fence():
    """Test parsing JSON wrapped in markdown code fences."""
    text = '```json\n{"key": "value"}\n```'
    result = parse_json_response(text)
    assert result == {"key": "value"}


def test_parse_json_response_with_plain_fence():
    """Test parsing JSON with plain code fences."""
    text = '```\n{"key": "value"}\n```'
    result = parse_json_response(text)
    assert result == {"key": "value"}


def test_parse_json_response_invalid():
    """Test that invalid JSON returns None."""
    result = parse_json_response("not json at all")
    assert result is None


def test_parse_json_response_empty():
    """Test that empty string returns None."""
    result = parse_json_response("")
    assert result is None


def test_parse_json_response_whitespace():
    """Test JSON with surrounding whitespace."""
    result = parse_json_response('  \n  {"key": "value"}  \n  ')
    assert result == {"key": "value"}
