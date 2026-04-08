package tools

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/oscar/my_opencode/internal/sandbox"
	"github.com/oscar/my_opencode/internal/security"
)

// ExecuteCommandTool executes shell commands in a sandboxed environment
type ExecuteCommandTool struct {
	sandbox   *sandbox.Sandbox
	validator *security.CommandValidator
	confirmer func(string) (bool, error)
	timeout   time.Duration
}

// NewExecuteCommandTool creates a new execute_command tool
func NewExecuteCommandTool(sb *sandbox.Sandbox, confirmer func(string) (bool, error)) *ExecuteCommandTool {
	return &ExecuteCommandTool{
		sandbox:   sb,
		validator: security.NewCommandValidator(),
		confirmer: confirmer,
		timeout:   30 * time.Second, // Default 30s timeout
	}
}

// Name returns the tool name
func (t *ExecuteCommandTool) Name() string {
	return "execute_command"
}

// Description returns the tool description
func (t *ExecuteCommandTool) Description() string {
	return "Executes a shell command in the sandboxed workspace directory (requires user confirmation)"
}

// Execute runs a shell command with security validation
func (t *ExecuteCommandTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Extract command argument
	cmdArg, ok := args["command"]
	if !ok {
		return "", fmt.Errorf("missing required argument: command")
	}

	command, ok := cmdArg.(string)
	if !ok {
		return "", fmt.Errorf("command must be a string")
	}

	// Validate command security
	if err := t.validator.Validate(command); err != nil {
		return "", fmt.Errorf("command validation failed: %w", err)
	}

	// Request user confirmation
	confirmed, err := t.confirmer(command)
	if err != nil {
		return "", fmt.Errorf("confirmation failed: %w", err)
	}

	if !confirmed {
		return "", fmt.Errorf("command execution denied by user")
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	// Execute command in workspace directory
	cmd := exec.CommandContext(execCtx, "sh", "-c", command)
	cmd.Dir = t.sandbox.WorkspaceDir()

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it was a timeout
		if execCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", t.timeout)
		}

		// Include stderr in error message
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
