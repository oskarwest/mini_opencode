package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents log level
type Level string

const (
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     Level                  `json:"level"`
	Event     string                 `json:"event"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Logger handles structured logging to a file
type Logger struct {
	file *os.File
	mu   sync.Mutex
}

// New creates a new logger that writes to the specified file
func New(logPath string) (*Logger, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file in append mode
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{file: file}, nil
}

// Log writes a log entry
func (l *Logger) Log(level Level, event string, details map[string]interface{}) {
	if l == nil || l.file == nil {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Event:     event,
		Details:   details,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.file.Write(data)
	l.file.Write([]byte("\n"))
}

// Info logs an info-level event
func (l *Logger) Info(event string, details map[string]interface{}) {
	l.Log(LevelInfo, event, details)
}

// Warn logs a warning-level event
func (l *Logger) Warn(event string, details map[string]interface{}) {
	l.Log(LevelWarn, event, details)
}

// Error logs an error-level event
func (l *Logger) Error(event string, details map[string]interface{}) {
	l.Log(LevelError, event, details)
}

// SessionStart logs session start
func (l *Logger) SessionStart(model string) {
	l.Info("session_start", map[string]interface{}{
		"model": model,
	})
}

// SessionEnd logs session end
func (l *Logger) SessionEnd() {
	l.Info("session_end", nil)
}

// UserMessage logs a user message
func (l *Logger) UserMessage(content string) {
	l.Info("user_message", map[string]interface{}{
		"content": content,
	})
}

// AssistantMessage logs an assistant message
func (l *Logger) AssistantMessage(content string, isToolCall bool) {
	l.Info("assistant_message", map[string]interface{}{
		"content":      content,
		"is_tool_call": isToolCall,
	})
}

// ToolExecuted logs a tool execution
func (l *Logger) ToolExecuted(toolName string, args map[string]interface{}, result string, err error) {
	details := map[string]interface{}{
		"tool":      toolName,
		"arguments": args,
	}

	if err != nil {
		details["result"] = "failure"
		details["error"] = err.Error()
		l.Error("tool_executed", details)
	} else {
		details["result"] = "success"
		details["output"] = result
		l.Info("tool_executed", details)
	}
}

// ModelChanged logs a model change
func (l *Logger) ModelChanged(oldModel, newModel string) {
	l.Info("model_changed", map[string]interface{}{
		"old_model": oldModel,
		"new_model": newModel,
	})
}

// ConversationCleared logs when conversation is cleared
func (l *Logger) ConversationCleared() {
	l.Info("conversation_cleared", nil)
}

// Close closes the log file
func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	return l.file.Close()
}
