package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Sandbox manages a restricted workspace directory
type Sandbox struct {
	workspaceDir string
}

// New creates a new sandbox with the given workspace directory
func New(workspaceDir string) (*Sandbox, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	// Create workspace directory if it doesn't exist
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	return &Sandbox{
		workspaceDir: absPath,
	}, nil
}

// WorkspaceDir returns the absolute path to the workspace
func (s *Sandbox) WorkspaceDir() string {
	return s.workspaceDir
}

// ResolvePath resolves a path relative to the workspace and validates it's within bounds
func (s *Sandbox) ResolvePath(path string) (string, error) {
	// If path is already absolute, check if it's within workspace
	var absPath string
	var err error

	if filepath.IsAbs(path) {
		absPath = filepath.Clean(path)
	} else {
		// Resolve relative to workspace
		absPath, err = filepath.Abs(filepath.Join(s.workspaceDir, path))
		if err != nil {
			return "", fmt.Errorf("failed to resolve path: %w", err)
		}
	}

	// Ensure the path is within workspace
	if !s.IsWithinWorkspace(absPath) {
		return "", fmt.Errorf("path '%s' is outside workspace directory", path)
	}

	return absPath, nil
}

// IsWithinWorkspace checks if a path is within the workspace directory
func (s *Sandbox) IsWithinWorkspace(path string) bool {
	// Clean the path to resolve any .. or . elements
	cleanPath := filepath.Clean(path)

	// Check if path starts with workspace directory
	relPath, err := filepath.Rel(s.workspaceDir, cleanPath)
	if err != nil {
		return false
	}

	// If relative path starts with "..", it's outside workspace
	return !strings.HasPrefix(relPath, "..") && !filepath.IsAbs(relPath)
}

// ValidatePath validates that a path is safe to use
func (s *Sandbox) ValidatePath(path string) error {
	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}

	// Resolve and validate
	_, err := s.ResolvePath(path)
	return err
}
