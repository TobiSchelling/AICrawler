package llm

import (
	"testing"
)

func TestParseJSONResponsePlain(t *testing.T) {
	result := ParseJSONResponse(`{"key": "value", "num": 42}`)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["key"] != "value" {
		t.Errorf("expected key='value', got %v", result["key"])
	}
	if result["num"] != float64(42) {
		t.Errorf("expected num=42, got %v", result["num"])
	}
}

func TestParseJSONResponseWithCodeFence(t *testing.T) {
	text := "```json\n{\"key\": \"value\"}\n```"
	result := ParseJSONResponse(text)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["key"] != "value" {
		t.Errorf("expected key='value', got %v", result["key"])
	}
}

func TestParseJSONResponseWithPlainFence(t *testing.T) {
	text := "```\n{\"key\": \"value\"}\n```"
	result := ParseJSONResponse(text)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["key"] != "value" {
		t.Errorf("expected key='value', got %v", result["key"])
	}
}

func TestParseJSONResponseInvalid(t *testing.T) {
	result := ParseJSONResponse("not json at all")
	if result != nil {
		t.Error("expected nil for invalid JSON")
	}
}

func TestParseJSONResponseEmpty(t *testing.T) {
	result := ParseJSONResponse("")
	if result != nil {
		t.Error("expected nil for empty string")
	}
}

func TestParseJSONResponseWhitespace(t *testing.T) {
	result := ParseJSONResponse("  \n  {\"key\": \"value\"}  \n  ")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["key"] != "value" {
		t.Errorf("expected key='value', got %v", result["key"])
	}
}
