package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/oscar/my_opencode/internal/apiclient"
	"github.com/oscar/my_opencode/internal/chat"
	"github.com/oscar/my_opencode/internal/config"
	"github.com/oscar/my_opencode/internal/logger"
	"github.com/oscar/my_opencode/internal/sandbox"
	"github.com/oscar/my_opencode/internal/tools"
	"github.com/oscar/my_opencode/internal/ui"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		ui.DisplayError(err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.LogFile)
	if err != nil {
		ui.DisplayError(err)
		os.Exit(1)
	}
	defer log.Close()

	// Log session start
	log.SessionStart(cfg.DefaultModel)

	// Display welcome message
	ui.DisplayWelcome()

	// Create API client
	client := apiclient.NewClient(cfg.APIBaseURL)

	// Initialize sandbox
	sb, err := sandbox.New(cfg.WorkspaceDir)
	if err != nil {
		ui.DisplayError(err)
		log.Error("sandbox_init_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
	ui.DisplayMessage("Workspace directory: " + sb.WorkspaceDir())

	// Initialize tool registry
	toolRegistry := tools.NewRegistry()
	toolRegistry.Register(tools.NewReadFileTool(sb))
	toolRegistry.Register(tools.NewWriteFileTool(sb))
	toolRegistry.Register(tools.NewListDirectoryTool(sb))
	toolRegistry.Register(tools.NewExecuteCommandTool(sb, ui.ConfirmCommand))

	// Use default model from config
	ui.DisplayMessage("Using model: " + cfg.DefaultModel)
	ui.DisplayMessage("Type /help for available commands.")

	// Create session and handler
	session := chat.NewSession(cfg.DefaultModel)

	// Initialize session with system prompt for tool usage
	systemPrompt := tools.GenerateSystemPrompt(toolRegistry)
	session.InitializeWithSystemPrompt(systemPrompt)

	handler := chat.NewHandler(client, session, toolRegistry, log, cfg)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Main chat loop
	for {
		// Check for interrupt signal
		select {
		case <-sigChan:
			ui.DisplayMessage("\n\nGoodbye!")
			log.SessionEnd()
			return
		default:
		}

		// Read user input
		userInput, err := ui.ReadUserInput()
		if err != nil {
			ui.DisplayMessage("\n\nGoodbye!")
			log.SessionEnd()
			return
		}

		// Skip empty input
		if userInput == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(userInput, "/") {
			switch strings.ToLower(userInput) {
			case "/exit":
				ui.DisplayMessage("\nGoodbye!")
				log.SessionEnd()
				return
			case "/help":
				ui.DisplayWelcome()
				continue
			case "/clear":
				session.Clear()
				log.ConversationCleared()
				ui.DisplayMessage("Conversation history cleared")
				continue
			case "/model":
				// Change model mid-session
				ui.DisplayMessage("Fetching available models...")

				// Fetch models from API
				modelCtx, modelCancel := context.WithTimeout(context.Background(), 30*time.Second)
				models, err := client.ListModels(modelCtx)
				modelCancel()

				if err != nil {
					ui.DisplayError(err)
					continue
				}

				if len(models) == 0 {
					ui.DisplayMessage("No models available")
					continue
				}

				// Prompt user to select a model
				newModel, err := ui.PromptModelSelection(models)
				if err != nil {
					ui.DisplayError(err)
					continue
				}

				oldModel := session.GetModel()
				session.SetModel(newModel)
				log.ModelChanged(oldModel, newModel)
				ui.DisplayMessage("Model changed to: " + newModel)
				continue
			case "/tools":
				// List available tools
				ui.DisplayToolsList(toolRegistry)
				continue
			default:
				ui.DisplayMessage("Unknown command. Type /help for available commands.")
				continue
			}
		}

		// Create context for this request (longer timeout for multi-step operations)
		reqCtx, reqCancel := context.WithTimeout(context.Background(), 180*time.Second)

		// Handle user input and stream response with tool execution loop
		err = handler.HandleUserInput(
			reqCtx,
			userInput,
			// Display callback - displays response content (only for non-tool responses)
			ui.DisplayAssistantResponse,
			// Response callback - called after receiving response, before display
			func(iteration int, isToolCall bool) {
				// Print "Assistant: " prefix
				ui.StartAssistantResponse()

				// If it's a tool call, we won't be displaying the JSON
				// The tool execution result will be shown instead
			},
			// Tool callback - called after each tool execution
			func(toolName, result string, err error) {
				// Display tool execution result
				ui.DisplayToolExecution(toolName, result, err)
			},
		)
		reqCancel()

		if err != nil {
			ui.DisplayError(err)
			continue
		}

		// End final assistant response
		ui.EndAssistantResponse()
	}
}
