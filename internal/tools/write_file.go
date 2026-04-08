package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oscar/my_opencode/internal/sandbox"
)

// WriteFileTool creates or overwrites a file
type WriteFileTool struct {
	sandbox *sandbox.Sandbox
}

// NewWriteFileTool creates a new write_file tool
func NewWriteFileTool(sb *sandbox.Sandbox) *WriteFileTool {
	return &WriteFileTool{sandbox: sb}
}

// Name returns the tool name
func (t *WriteFileTool) Name() string {
	return "write_file"
}

// Description returns the tool description
func (t *WriteFileTool) Description() string {
	return "Creates or overwrites a file with the given content"
}

// Execute writes content to a file
func (t *WriteFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Extract path argument
	pathArg, ok := args["path"]
	if !ok {
		return "", fmt.Errorf("missing required argument: path")
	}

	path, ok := pathArg.(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}

	// Extract content argument
	contentArg, ok := args["content"]
	if !ok {
		return "", fmt.Errorf("missing required argument: content")
	}

	content, ok := contentArg.(string)
	if !ok {
		return "", fmt.Errorf("content must be a string")
	}

	// Validate and resolve path
	absPath, err := t.sandbox.ResolvePath(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Write file
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}
