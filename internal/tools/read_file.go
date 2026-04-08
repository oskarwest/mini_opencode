package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/oscar/my_opencode/internal/sandbox"
)

// ReadFileTool reads the contents of a file
type ReadFileTool struct {
	sandbox *sandbox.Sandbox
}

// NewReadFileTool creates a new read_file tool
func NewReadFileTool(sb *sandbox.Sandbox) *ReadFileTool {
	return &ReadFileTool{sandbox: sb}
}

// Name returns the tool name
func (t *ReadFileTool) Name() string {
	return "read_file"
}

// Description returns the tool description
func (t *ReadFileTool) Description() string {
	return "Reads and returns the content of a file"
}

// Execute reads a file and returns its contents
func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Extract path argument
	pathArg, ok := args["path"]
	if !ok {
		return "", fmt.Errorf("missing required argument: path")
	}

	path, ok := pathArg.(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}

	// Validate and resolve path
	absPath, err := t.sandbox.ResolvePath(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check if file exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file does not exist: %s", path)
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Read file contents
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}
