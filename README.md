# Mini OpenCode

A terminal-based AI chat application written in Go that connects to any OpenAI-compatible API. It features built-in tool execution capabilities, allowing the AI to read/write files, list directories, and run shell commands within a sandboxed workspace.

## Features

- **Streaming chat** — Real-time responses via Server-Sent Events (SSE)
- **Tool execution loop** — The AI can chain up to 10 tool calls per request automatically
- **Sandboxed workspace** — All file operations are restricted to a `./workspace` directory
- **Command security** — Whitelisted commands, dangerous pattern blocking, and user confirmation prompts
- **Model switching** — Change the active model mid-session with `/model`
- **Structured logging** — JSON logs for all interactions in `./logs/chat.log`
- **Colored terminal UI** — Visual differentiation for user input, assistant responses, tool results, and errors

## Available Tools

| Tool | Description |
|------|-------------|
| `read_file` | Read file contents from the workspace |
| `write_file` | Create or overwrite files in the workspace |
| `list_directory` | List directory contents |
| `execute_command` | Run whitelisted shell commands (requires user confirmation) |

## Requirements

- Go 1.25+
- An OpenAI-compatible API endpoint (e.g., LM Studio, Ollama, vLLM, OpenAI)

## Installation

```bash
git clone git@github.com:oskarwest/mini_opencode.git
cd mini_opencode
go build -o mini_opencode ./cmd/chat
```

## Configuration

Edit `config.yaml` to point to your API:

```yaml
api_base_url: https://your-api-endpoint/v1
default_model: your-model-name
temperature: 0.7
max_tool_iterations: 10
command_timeout_seconds: 30
workspace_dir: ./workspace
log_file: ./logs/chat.log
```

## Usage

```bash
./mini_opencode
```

### Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/exit` | Exit the application |
| `/clear` | Clear conversation history |
| `/model` | List and switch to a different model |
| `/tools` | List available tools |

### Example

```
You: Create a Python script that prints the Fibonacci sequence and run it

[Tool 'write_file' executed successfully]
[Tool 'execute_command' executed successfully]
```

## Project Structure

```
mini_opencode/
├── config.yaml              # Application configuration
├── cmd/chat/
│   └── main.go              # Entry point
└── internal/
    ├── apiclient/           # OpenAI-compatible API client (streaming)
    ├── chat/                # Session management and tool execution loop
    ├── config/              # YAML configuration loader
    ├── logger/              # Structured JSON logger
    ├── sandbox/             # Workspace isolation and path validation
    ├── security/            # Command whitelisting and pattern blocking
    ├── tools/               # Tool interface, registry, and implementations
    └── ui/                  # Terminal I/O and ANSI colors
```

## Security

- All file operations are restricted to the `./workspace` directory with path traversal prevention
- Shell commands are validated against a whitelist of safe commands (e.g., `ls`, `cat`, `python`, `go`, `git`)
- Dangerous patterns are blocked: `rm -rf /`, `sudo`, fork bombs, pipe-to-shell, etc.
- Every command requires explicit user confirmation before execution
- Commands have a configurable timeout (default 30s)

## License

MIT