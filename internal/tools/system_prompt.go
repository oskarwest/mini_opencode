package tools

import "fmt"

// GenerateSystemPrompt creates a system prompt that instructs the model to use tools
func GenerateSystemPrompt(registry *Registry) string {
	tools := registry.List()

	prompt := `You are an AI assistant with access to tools for file operations and command execution. When the user asks you to perform an action that requires a tool, you MUST respond with a JSON tool call instead of explaining how to do it manually.

IMPORTANT: When a tool is needed, your response should contain ONLY the JSON tool call, with no additional explanation or text before or after it. Do not say "I'll use the tool" or "Here's the command" - just output the JSON directly.

Available Tools:
`

	// Add each tool's description
	for _, tool := range tools {
		prompt += fmt.Sprintf("\n%s\n", tool.Name())
		prompt += fmt.Sprintf("  Description: %s\n", tool.Description())

		// Add usage examples
		switch tool.Name() {
		case "read_file":
			prompt += `  Arguments: {"path": "file_path"}
  Example use case: When user asks to read, show, display, or check contents of a file
  Example JSON:
  {
    "tool": "read_file",
    "arguments": {
      "path": "example.txt"
    }
  }
`
		case "write_file":
			prompt += `  Arguments: {"path": "file_path", "content": "file_content"}
  Example use case: When user asks to create, write, save, or update a file
  Example JSON:
  {
    "tool": "write_file",
    "arguments": {
      "path": "example.txt",
      "content": "Hello World"
    }
  }
`
		case "list_directory":
			prompt += `  Arguments: {"path": "directory_path"} (optional, defaults to workspace root)
  Example use case: When user asks to list, show, or see files in a directory
  Example JSON:
  {
    "tool": "list_directory",
    "arguments": {
      "path": "."
    }
  }
`
		case "execute_command":
			prompt += `  Arguments: {"command": "shell_command"}
  Example use case: When user asks to run, execute a command or perform system operations
  Example JSON:
  {
    "tool": "execute_command",
    "arguments": {
      "command": "ls -la"
    }
  }
  Note: This tool requires user confirmation and has security restrictions
`
		}
	}

	prompt += `
Tool Invocation Rules:
1. When the user asks to read a file, respond with ONLY the read_file JSON tool call
2. When the user asks to write/create a file, respond with ONLY the write_file JSON tool call
3. When the user asks to list files/directories, respond with ONLY the list_directory JSON tool call
4. When the user asks to execute a command, respond with ONLY the execute_command JSON tool call
5. The JSON must be valid and follow the exact format shown above
6. You can optionally wrap the JSON in markdown code blocks for better formatting
7. After the tool executes, you will receive the result and can then provide an explanation

Example Interaction:
User: "Create a file called test.txt with the content 'Hello World'"
Your Response:
{
  "tool": "write_file",
  "arguments": {
    "path": "test.txt",
    "content": "Hello World"
  }
}

User: "What's in test.txt?"
Your Response:
{
  "tool": "read_file",
  "arguments": {
    "path": "test.txt"
  }
}

After the tool executes, you'll see the result and can provide context or explanation if needed.

Multi-Step Operations:
When the user requests an action that requires multiple steps (e.g., "create a Python script and run it"), you should:
1. Execute the first tool (e.g., write_file to create the script)
2. Wait for the result in the conversation
3. Execute the next tool (e.g., execute_command to run the script)
4. Continue until all steps are complete
5. Provide a final summary when done

You can execute up to 10 tool calls in a single conversation turn. Each tool result will be added to the conversation automatically.

Example Multi-Step:
User: "Create a Python script that prints 'Hello' and run it"

Step 1 - Your Response:
{
  "tool": "write_file",
  "arguments": {
    "path": "hello.py",
    "content": "print('Hello')"
  }
}

[System adds: "Tool 'write_file' result: Successfully wrote 15 bytes to hello.py"]

Step 2 - Your Response:
{
  "tool": "execute_command",
  "arguments": {
    "command": "python hello.py"
  }
}

[System adds: "Tool 'execute_command' result: Hello"]

Step 3 - Your Response:
The script has been created and executed successfully. It printed "Hello" as expected.

CRITICAL: Do not explain what you're going to do - just output the JSON tool call immediately when a tool action is requested.
`

	return prompt
}
