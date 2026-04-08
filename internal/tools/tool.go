package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Tool represents an executable tool
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// ToolCall represents a parsed tool invocation from the model
type ToolCall struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Registry manages available tools
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ParseToolCall attempts to extract a tool call from model response
func ParseToolCall(response string) (*ToolCall, error) {
	var jsonStr string

	// Strategy 1: Look for JSON in markdown code blocks (with or without language specifier)
	// Patterns to try:
	// - ```json\n{...}\n```
	// - ```\n{...}\n```
	// - ```{...}```
	// - ```json{...}```
	codeBlockPatterns := []*regexp.Regexp{
		regexp.MustCompile("```json\\s*\\n\\s*(\\{[\\s\\S]*?\\})\\s*\\n\\s*```"),
		regexp.MustCompile("```\\s*\\n\\s*(\\{[\\s\\S]*?\\})\\s*\\n\\s*```"),
		regexp.MustCompile("```json\\s*(\\{[\\s\\S]*?\\})\\s*```"),
		regexp.MustCompile("```\\s*(\\{[\\s\\S]*?\\})\\s*```"),
	}

	for _, pattern := range codeBlockPatterns {
		matches := pattern.FindStringSubmatch(response)
		if len(matches) > 1 {
			jsonStr = matches[1]
			break
		}
	}

	// Strategy 2: If no code block found, try to extract raw JSON
	if jsonStr == "" {
		jsonStr = extractJSON(response)
	}

	if jsonStr == "" {
		return nil, fmt.Errorf("no tool call found in response")
	}

	// Clean up the JSON string
	jsonStr = strings.TrimSpace(jsonStr)

	// Try to parse as ToolCall
	var toolCall ToolCall
	if err := json.Unmarshal([]byte(jsonStr), &toolCall); err != nil {
		return nil, fmt.Errorf("failed to parse tool call JSON: %w", err)
	}

	// Validate required fields
	if toolCall.Tool == "" {
		return nil, fmt.Errorf("tool call missing 'tool' field")
	}

	return &toolCall, nil
}

// extractJSON finds and extracts a JSON object from text using brace matching
func extractJSON(text string) string {
	// Find the first opening brace
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}

	// Count braces to find the matching closing brace
	braceCount := 0
	inString := false
	escapeNext := false

	for i := start; i < len(text); i++ {
		char := text[i]

		// Handle escape sequences in strings
		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		// Handle string boundaries
		if char == '"' {
			inString = !inString
			continue
		}

		// Only count braces outside of strings
		if !inString {
			if char == '{' {
				braceCount++
			} else if char == '}' {
				braceCount--
				if braceCount == 0 {
					// Found matching closing brace
					return text[start : i+1]
				}
			}
		}
	}

	return ""
}

// ContainsToolCall checks if response contains a tool call pattern
func ContainsToolCall(response string) bool {
	// Quick check for tool call indicators
	return strings.Contains(response, `"tool"`) && strings.Contains(response, `"arguments"`)
}
