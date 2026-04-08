package tools

import (
	"testing"
)

func TestParseToolCall_CodeBlockWithJson(t *testing.T) {
	response := "```json\n{\n  \"tool\": \"write_file\",\n  \"arguments\": {\n    \"path\": \"test.txt\",\n    \"content\": \"Hello World\"\n  }\n}\n```"

	toolCall, err := ParseToolCall(response)
	if err != nil {
		t.Fatalf("ParseToolCall failed: %v", err)
	}

	if toolCall.Tool != "write_file" {
		t.Errorf("Expected tool 'write_file', got '%s'", toolCall.Tool)
	}

	if toolCall.Arguments["path"] != "test.txt" {
		t.Errorf("Expected path 'test.txt', got '%v'", toolCall.Arguments["path"])
	}

	if toolCall.Arguments["content"] != "Hello World" {
		t.Errorf("Expected content 'Hello World', got '%v'", toolCall.Arguments["content"])
	}
}

func TestParseToolCall_CodeBlockWithoutJson(t *testing.T) {
	response := "```\n{\n  \"tool\": \"read_file\",\n  \"arguments\": {\n    \"path\": \"example.txt\"\n  }\n}\n```"

	toolCall, err := ParseToolCall(response)
	if err != nil {
		t.Fatalf("ParseToolCall failed: %v", err)
	}

	if toolCall.Tool != "read_file" {
		t.Errorf("Expected tool 'read_file', got '%s'", toolCall.Tool)
	}
}

func TestParseToolCall_CodeBlockNoNewlines(t *testing.T) {
	response := "```json{\"tool\": \"list_directory\", \"arguments\": {\"path\": \".\"}}```"

	toolCall, err := ParseToolCall(response)
	if err != nil {
		t.Fatalf("ParseToolCall failed: %v", err)
	}

	if toolCall.Tool != "list_directory" {
		t.Errorf("Expected tool 'list_directory', got '%s'", toolCall.Tool)
	}
}

func TestParseToolCall_PlainJSON(t *testing.T) {
	response := `{
  "tool": "execute_command",
  "arguments": {
    "command": "ls -la"
  }
}`

	toolCall, err := ParseToolCall(response)
	if err != nil {
		t.Fatalf("ParseToolCall failed: %v", err)
	}

	if toolCall.Tool != "execute_command" {
		t.Errorf("Expected tool 'execute_command', got '%s'", toolCall.Tool)
	}
}

func TestParseToolCall_JSONInText(t *testing.T) {
	response := `I'll help you with that. Here's the tool call:
{
  "tool": "write_file",
  "arguments": {
    "path": "output.txt",
    "content": "Test content"
  }
}
This will create the file.`

	toolCall, err := ParseToolCall(response)
	if err != nil {
		t.Fatalf("ParseToolCall failed: %v", err)
	}

	if toolCall.Tool != "write_file" {
		t.Errorf("Expected tool 'write_file', got '%s'", toolCall.Tool)
	}
}

func TestParseToolCall_CompactJSON(t *testing.T) {
	response := `{"tool":"read_file","arguments":{"path":"test.txt"}}`

	toolCall, err := ParseToolCall(response)
	if err != nil {
		t.Fatalf("ParseToolCall failed: %v", err)
	}

	if toolCall.Tool != "read_file" {
		t.Errorf("Expected tool 'read_file', got '%s'", toolCall.Tool)
	}
}

func TestParseToolCall_NestedArguments(t *testing.T) {
	response := `{
  "tool": "complex_tool",
  "arguments": {
    "nested": {
      "level1": {
        "level2": "value"
      }
    },
    "array": [1, 2, 3]
  }
}`

	toolCall, err := ParseToolCall(response)
	if err != nil {
		t.Fatalf("ParseToolCall failed: %v", err)
	}

	if toolCall.Tool != "complex_tool" {
		t.Errorf("Expected tool 'complex_tool', got '%s'", toolCall.Tool)
	}

	// Verify nested structure exists
	if toolCall.Arguments["nested"] == nil {
		t.Error("Expected nested arguments")
	}
}

func TestParseToolCall_JSONWithEscapedQuotes(t *testing.T) {
	response := `{
  "tool": "write_file",
  "arguments": {
    "path": "quote.txt",
    "content": "She said \"Hello\""
  }
}`

	toolCall, err := ParseToolCall(response)
	if err != nil {
		t.Fatalf("ParseToolCall failed: %v", err)
	}

	if toolCall.Tool != "write_file" {
		t.Errorf("Expected tool 'write_file', got '%s'", toolCall.Tool)
	}

	content, ok := toolCall.Arguments["content"].(string)
	if !ok || content != `She said "Hello"` {
		t.Errorf("Expected escaped quotes in content, got '%v'", content)
	}
}

func TestParseToolCall_NoJSON(t *testing.T) {
	response := "This is just plain text without any JSON"

	_, err := ParseToolCall(response)
	if err == nil {
		t.Error("Expected error for response without JSON, got nil")
	}
}

func TestParseToolCall_InvalidJSON(t *testing.T) {
	response := `{
  "tool": "test",
  "arguments": {
    "broken":
  }
}`

	_, err := ParseToolCall(response)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseToolCall_MissingToolField(t *testing.T) {
	response := `{
  "arguments": {
    "path": "test.txt"
  }
}`

	_, err := ParseToolCall(response)
	if err == nil {
		t.Error("Expected error for missing 'tool' field, got nil")
	}
}

func TestContainsToolCall(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected bool
	}{
		{
			name:     "Valid tool call",
			response: `{"tool": "test", "arguments": {}}`,
			expected: true,
		},
		{
			name:     "Tool call in text",
			response: `Here is a tool call: {"tool": "test", "arguments": {}}`,
			expected: true,
		},
		{
			name:     "No tool call",
			response: "Just plain text",
			expected: false,
		},
		{
			name:     "Only tool field",
			response: `{"tool": "test"}`,
			expected: false,
		},
		{
			name:     "Only arguments field",
			response: `{"arguments": {}}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsToolCall(tt.response)
			if result != tt.expected {
				t.Errorf("ContainsToolCall(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestExtractJSON_BraceMatching(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple JSON",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "Nested JSON",
			input:    `{"outer": {"inner": "value"}}`,
			expected: `{"outer": {"inner": "value"}}`,
		},
		{
			name:     "JSON with text before",
			input:    `Here is some text {"key": "value"} and after`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with braces in string",
			input:    `{"message": "This has { and } inside"}`,
			expected: `{"message": "This has { and } inside"}`,
		},
		{
			name:     "No JSON",
			input:    "Just plain text",
			expected: "",
		},
		{
			name:     "Unmatched braces",
			input:    `{"unclosed": `,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("extractJSON() = %q, want %q", result, tt.expected)
			}
		})
	}
}
