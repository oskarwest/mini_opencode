package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/oscar/my_opencode/internal/sandbox"
)

// ListDirectoryTool lists files and folders in a directory
type ListDirectoryTool struct {
	sandbox *sandbox.Sandbox
}

// NewListDirectoryTool creates a new list_directory tool
func NewListDirectoryTool(sb *sandbox.Sandbox) *ListDirectoryTool {
	return &ListDirectoryTool{sandbox: sb}
}

// Name returns the tool name
func (t *ListDirectoryTool) Name() string {
	return "list_directory"
}

// Description returns the tool description
func (t *ListDirectoryTool) Description() string {
	return "Lists files and folders in a given directory path"
}

// Execute lists directory contents
func (t *ListDirectoryTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Extract path argument (optional, defaults to workspace root)
	path := "."
	if pathArg, ok := args["path"]; ok {
		if pathStr, ok := pathArg.(string); ok {
			path = pathStr
		}
	}

	// Validate and resolve path
	absPath, err := t.sandbox.ResolvePath(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", path)
		}
		return "", fmt.Errorf("failed to stat path: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", path)
	}

	// Read directory contents
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	// Format output
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Contents of %s:\n\n", path))

	if len(entries) == 0 {
		result.WriteString("(empty directory)\n")
		return result.String(), nil
	}

	// Separate directories and files
	var dirs []string
	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name()+"/")
		} else {
			// Get file size
			info, err := entry.Info()
			if err == nil {
				files = append(files, fmt.Sprintf("%-40s %10d bytes", entry.Name(), info.Size()))
			} else {
				files = append(files, entry.Name())
			}
		}
	}

	// Write directories first
	if len(dirs) > 0 {
		result.WriteString("Directories:\n")
		for _, dir := range dirs {
			result.WriteString(fmt.Sprintf("  %s\n", dir))
		}
		result.WriteString("\n")
	}

	// Write files
	if len(files) > 0 {
		result.WriteString("Files:\n")
		for _, file := range files {
			result.WriteString(fmt.Sprintf("  %s\n", file))
		}
	}

	return result.String(), nil
}
