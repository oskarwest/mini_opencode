package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/oscar/my_opencode/internal/apiclient"
	"github.com/oscar/my_opencode/internal/tools"
)

// PromptModelSelection displays available models and prompts the user to select one
func PromptModelSelection(models []apiclient.Model) (string, error) {
	fmt.Println("\nAvailable models:")
	for i, model := range models {
		fmt.Printf("%d. %s\n", i+1, model.ID)
	}

	fmt.Print("\nSelect a model (enter number): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	// Echo the input back so it's visible in terminal history
	fmt.Printf("%s\n", input)

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(models) {
		return "", fmt.Errorf("invalid selection")
	}

	return models[selection-1].ID, nil
}

// ReadUserInput reads a line of input from the user
func ReadUserInput() (string, error) {
	fmt.Print("\n" + Colorize("You: ", ColorCyan))

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		// EOF reached (Ctrl+D)
		return "", fmt.Errorf("EOF")
	}

	input := strings.TrimSpace(scanner.Text())
	return input, nil
}

// DisplayAssistantResponse displays streaming content from the assistant
func DisplayAssistantResponse(content string) {
	fmt.Print(content)
}

// StartAssistantResponse prints the assistant prefix
func StartAssistantResponse() {
	fmt.Print("\n" + Colorize("Assistant: ", ColorGreen))
}

// EndAssistantResponse prints a newline after the response
func EndAssistantResponse() {
	fmt.Println()
}

// DisplayError displays an error message
func DisplayError(err error) {
	fmt.Printf("\n"+Colorize("Error: ", ColorRed)+"%v\n", err)
}

// DisplayMessage displays a general message
func DisplayMessage(msg string) {
	fmt.Println(Colorize(msg, ColorGray))
}

// DisplayWelcome displays the welcome message
func DisplayWelcome() {
	fmt.Println("=== Terminal Chat Application ===")
	fmt.Println("Commands:")
	fmt.Println("  /help   - Show this help message")
	fmt.Println("  /exit   - Exit the application")
	fmt.Println("  /clear  - Clear conversation history")
	fmt.Println("  /model  - Change the active model")
	fmt.Println("  /tools  - List available tools")
	fmt.Println()
}

// ConfirmCommand prompts the user to confirm command execution
func ConfirmCommand(command string) (bool, error) {
	fmt.Print("\n" + Colorize("The model wants to execute: ", ColorYellow) + command + "\n")
	fmt.Print(Colorize("Confirm? (y/n): ", ColorYellow))

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	// Echo the confirmation response so it's visible in terminal history
	fmt.Printf("%s\n", input)

	return input == "y" || input == "yes", nil
}

// DisplayToolsList displays all available tools
func DisplayToolsList(registry *tools.Registry) {
	fmt.Println("\nAvailable Tools:")
	fmt.Println()

	toolsList := registry.List()
	if len(toolsList) == 0 {
		fmt.Println("  (no tools registered)")
		return
	}

	for _, tool := range toolsList {
		fmt.Printf("  • %s\n", tool.Name())
		fmt.Printf("    %s\n\n", tool.Description())
	}
}

// DisplayToolExecution displays tool execution information
func DisplayToolExecution(toolName string, result string, err error) {
	if err != nil {
		fmt.Printf("\n"+Colorize("[Tool '%s' failed: %v]", ColorRed)+"\n", toolName, err)
	} else {
		fmt.Printf("\n"+Colorize("[Tool '%s' executed successfully]", ColorGreen)+"\n", toolName)
		if result != "" {
			fmt.Print(Colorize("Result:\n", ColorYellow))
			fmt.Printf("%s\n", result)
		}
	}
}
