package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/oscar/my_opencode/internal/apiclient"
	"github.com/oscar/my_opencode/internal/config"
	"github.com/oscar/my_opencode/internal/logger"
	"github.com/oscar/my_opencode/internal/tools"
)

// Handler orchestrates chat interactions
type Handler struct {
	client       *apiclient.Client
	session      *Session
	toolRegistry *tools.Registry
	logger       *logger.Logger
	config       *config.Config
}

// NewHandler creates a new chat handler
func NewHandler(client *apiclient.Client, session *Session, toolRegistry *tools.Registry, log *logger.Logger, cfg *config.Config) *Handler {
	return &Handler{
		client:       client,
		session:      session,
		toolRegistry: toolRegistry,
		logger:       log,
		config:       cfg,
	}
}

// GetToolRegistry returns the tool registry
func (h *Handler) GetToolRegistry() *tools.Registry {
	return h.toolRegistry
}

// ToolExecutionCallback is called when a tool is executed
// It receives the tool name, result (empty if error), and error (nil if success)
type ToolExecutionCallback func(toolName string, result string, err error)

// ResponseCallback is called before streaming each assistant response
type ResponseCallback func(iteration int, isToolCall bool)

// HandleUserInput processes user input and streams the response
// It implements a tool execution loop that allows the model to execute multiple tools in sequence
func (h *Handler) HandleUserInput(
	ctx context.Context,
	userInput string,
	displayCallback func(string),
	responseCallback ResponseCallback,
	toolCallback ToolExecutionCallback,
) error {
	// Add user message to session
	h.session.AddMessage("user", userInput)

	// Log user message
	h.logger.UserMessage(userInput)

	// Tool execution loop
	maxIterations := h.config.MaxToolIterations
	for iteration := 0; iteration < maxIterations; iteration++ {
		// Build chat request
		req := apiclient.ChatRequest{
			Model:       h.session.GetModel(),
			Messages:    h.session.GetMessages(),
			Temperature: h.config.Temperature,
			Stream:      true,
		}

		// Accumulate assistant response (don't display during streaming)
		var assistantResponse strings.Builder

		// Stream the response (accumulate only, don't display yet)
		err := h.client.StreamChatCompletion(ctx, req, func(content string) {
			assistantResponse.WriteString(content)
			// Don't call displayCallback here - we'll display after checking if it's a tool call
		})

		if err != nil {
			// Remove the user message if the request failed (only on first iteration)
			if iteration == 0 {
				messages := h.session.GetMessages()
				if len(messages) > 0 {
					h.session.messages = messages[:len(messages)-1]
				}
			}
			return fmt.Errorf("failed to get response: %w", err)
		}

		// Get the full response
		fullResponse := assistantResponse.String()

		// Check if response contains a tool call
		isToolCall := tools.ContainsToolCall(fullResponse)

		// Notify about the response (with tool call status)
		responseCallback(iteration, isToolCall)

		// Only display if it's NOT a tool call
		if !isToolCall {
			displayCallback(fullResponse)
		}

		// Add assistant response to session
		h.session.AddMessage("assistant", fullResponse)

		// Log assistant message
		h.logger.AssistantMessage(fullResponse, isToolCall)

		// If no tool call, conversation complete
		if !isToolCall {
			return nil
		}

		// Parse and execute tool call
		toolCall, err := tools.ParseToolCall(fullResponse)
		if err != nil {
			// Failed to parse, but response is already added to history
			// Just return the error and let the user decide
			return fmt.Errorf("failed to parse tool call: %w", err)
		}

		// Execute the tool
		tool, exists := h.toolRegistry.Get(toolCall.Tool)
		if !exists {
			// Unknown tool, notify and stop
			toolCallback(toolCall.Tool, "", fmt.Errorf("unknown tool: %s", toolCall.Tool))
			return fmt.Errorf("unknown tool: %s", toolCall.Tool)
		}

		result, err := tool.Execute(ctx, toolCall.Arguments)

		// Log tool execution
		h.logger.ToolExecuted(toolCall.Tool, toolCall.Arguments, result, err)

		// Notify about tool execution
		toolCallback(toolCall.Tool, result, err)

		if err != nil {
			// Tool execution failed, add error to conversation and stop
			errorMsg := fmt.Sprintf("Tool '%s' failed: %v", toolCall.Tool, err)
			h.session.AddMessage("user", errorMsg)
			h.logger.Error("tool_execution_failed", map[string]interface{}{
				"tool":  toolCall.Tool,
				"error": err.Error(),
			})
			return fmt.Errorf("tool execution failed: %w", err)
		}

		// Add tool result to conversation for the model to see
		toolResultMsg := fmt.Sprintf("Tool '%s' result: %s", toolCall.Tool, result)
		h.session.AddMessage("user", toolResultMsg)

		// Continue loop to let model process the result and potentially call more tools
	}

	// Max iterations reached
	return fmt.Errorf("maximum tool execution iterations (%d) reached", maxIterations)
}

