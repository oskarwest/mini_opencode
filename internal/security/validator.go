package security

import (
	"fmt"
	"regexp"
	"strings"
)

// CommandValidator validates shell commands for security
type CommandValidator struct {
	allowedCommands map[string]bool
	dangerousPatterns []*regexp.Regexp
}

// NewCommandValidator creates a new command validator
func NewCommandValidator() *CommandValidator {
	// Whitelist of allowed commands
	allowedCommands := map[string]bool{
		"ls":     true,
		"cat":    true,
		"echo":   true,
		"pwd":    true,
		"grep":   true,
		"find":   true,
		"head":   true,
		"tail":   true,
		"wc":     true,
		"sort":   true,
		"uniq":   true,
		"cut":    true,
		"sed":    true,
		"awk":    true,
		"tr":     true,
		"mkdir":  true,
		"touch":  true,
		"cp":     true,
		"mv":     true,
		"chmod":  true,
		"diff":   true,
		"python": true,
		"python3": true,
		"node":   true,
		"npm":    true,
		"git":    true,
		"wget":   true,
		"curl":   true,
		"tar":    true,
		"gzip":   true,
		"gunzip": true,
		"zip":    true,
		"unzip":  true,
		"make":   true,
		"gcc":    true,
		"g++":    true,
		"go":     true,
		"cargo":  true,
		"rustc":  true,
		"java":   true,
		"javac":  true,
	}

	// Dangerous command patterns
	dangerousPatterns := []*regexp.Regexp{
		// Destructive file operations
		regexp.MustCompile(`rm\s+(-[rf]*\s+)*(/|~|\*)`),
		regexp.MustCompile(`rm\s+.*\s+/`),
		regexp.MustCompile(`mkfs`),
		regexp.MustCompile(`dd\s+.*\s+of=/dev/`),

		// System operations
		regexp.MustCompile(`shutdown`),
		regexp.MustCompile(`reboot`),
		regexp.MustCompile(`halt`),
		regexp.MustCompile(`poweroff`),
		regexp.MustCompile(`init\s+[06]`),

		// Fork bombs and resource exhaustion
		regexp.MustCompile(`:\(\)\{.*:\|:.*\};:`), // Classic fork bomb
		regexp.MustCompile(`while.*true.*do`),      // Potential infinite loop

		// Path traversal attempts
		regexp.MustCompile(`\.\./\.\./`),
		regexp.MustCompile(`/etc/passwd`),
		regexp.MustCompile(`/etc/shadow`),

		// Package management (could install malware)
		regexp.MustCompile(`apt-get|apt|yum|dnf|pacman.*-S`),

		// Privilege escalation
		regexp.MustCompile(`sudo`),
		regexp.MustCompile(`su\s`),

		// Network services
		regexp.MustCompile(`nc\s+.*\s+-l`), // Netcat listener
		regexp.MustCompile(`ncat\s+.*\s+-l`),

		// Kernel operations
		regexp.MustCompile(`modprobe`),
		regexp.MustCompile(`insmod`),
		regexp.MustCompile(`rmmod`),

		// Cron jobs
		regexp.MustCompile(`crontab`),

		// Shell spawning (potential breakout)
		regexp.MustCompile(`bash\s+-c`),
		regexp.MustCompile(`sh\s+-c`),
		regexp.MustCompile(`eval`),
	}

	return &CommandValidator{
		allowedCommands:   allowedCommands,
		dangerousPatterns: dangerousPatterns,
	}
}

// Validate checks if a command is safe to execute
func (v *CommandValidator) Validate(command string) error {
	// Trim whitespace
	command = strings.TrimSpace(command)

	if command == "" {
		return fmt.Errorf("empty command")
	}

	// Check for dangerous patterns first
	for _, pattern := range v.dangerousPatterns {
		if pattern.MatchString(command) {
			return fmt.Errorf("dangerous command pattern detected: blocked for security")
		}
	}

	// Extract the base command (first word)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("invalid command")
	}

	baseCommand := parts[0]

	// Remove path if present (e.g., /bin/ls -> ls)
	if strings.Contains(baseCommand, "/") {
		baseCommand = baseCommand[strings.LastIndex(baseCommand, "/")+1:]
	}

	// Check against whitelist
	if !v.allowedCommands[baseCommand] {
		return fmt.Errorf("command '%s' is not in the allowed list", baseCommand)
	}

	// Additional validation for specific commands
	if err := v.validateSpecificCommand(baseCommand, command); err != nil {
		return err
	}

	return nil
}

// validateSpecificCommand performs additional validation for specific commands
func (v *CommandValidator) validateSpecificCommand(baseCommand, fullCommand string) error {
	switch baseCommand {
	case "rm":
		// Extra caution with rm command
		if strings.Contains(fullCommand, "-rf") || strings.Contains(fullCommand, "-fr") {
			// Block rm -rf without specific file target
			if strings.Contains(fullCommand, "*") {
				return fmt.Errorf("rm -rf with wildcards is not allowed")
			}
		}
	case "chmod":
		// Block dangerous chmod operations
		if strings.Contains(fullCommand, "777") {
			return fmt.Errorf("chmod 777 is not recommended")
		}
	case "wget", "curl":
		// Block pipe to shell
		if strings.Contains(fullCommand, "|") && (strings.Contains(fullCommand, "sh") || strings.Contains(fullCommand, "bash")) {
			return fmt.Errorf("piping downloads to shell is not allowed")
		}
	}

	return nil
}

// IsCommandAllowed checks if a command is in the whitelist (without full validation)
func (v *CommandValidator) IsCommandAllowed(command string) bool {
	parts := strings.Fields(strings.TrimSpace(command))
	if len(parts) == 0 {
		return false
	}

	baseCommand := parts[0]
	if strings.Contains(baseCommand, "/") {
		baseCommand = baseCommand[strings.LastIndex(baseCommand, "/")+1:]
	}

	return v.allowedCommands[baseCommand]
}
