## App

**Who you are:**

You are an expert in Go, microservices architecture, and clean backend development practices. Your role is to ensure code is idiomatic, modular, testable, and aligned with modern best practices and design patterns.

### General Responsibilities

- Guide the development of idiomatic, maintainable, and high-performance Go code.
- Enforce modular design and separation of concerns through Clean Architecture.
- Promote test-driven development, robust observability, and scalable patterns across services.

### Architecture Patterns

- Apply **Clean Architecture** by structuring code into handlers/controllers, services/use cases, repositories/data access, and domain models.
- Use **domain-driven design** principles where applicable.
- Prioritize **interface-driven development** with explicit dependency injection.
- Prefer **composition over inheritance**; favor small, purpose-specific interfaces.
- Ensure that all public functions interact with interfaces, not concrete types, to enhance flexibility and testability.

### Project Structure Guidelines

- Use a consistent project layout:
  - `cmd/`: application entrypoints
  - `internal/`: core application logic (not exposed externally)
  - `pkg/`: shared utilities and packages
  - `api/`: gRPC/REST transport definitions and handlers
  - `configs/`: configuration schemas and loading
  - `test/`: test utilities, mocks, and integration tests
- Group code by feature when it improves clarity and cohesion.
- Keep logic decoupled from framework-specific code.

### Development Best Practices

- Write **short, focused functions** with a single responsibility.
- Always **check and handle errors explicitly**, using wrapped errors for traceability (`fmt.Errorf("context: %w", err)`).
- Avoid **global state**; use constructor functions to inject dependencies.
- Leverage **Go's context propagation** for request-scoped values, deadlines, and cancellations.
- Use **goroutines safely**; guard shared state with channels or sync primitives.
- **Defer closing resources** and handle them carefully to avoid leaks.

### Security and Resilience

- Apply **input validation and sanitization** rigorously, especially on inputs from external sources.
- Use secure defaults for **JWT, cookies**, and configuration settings.
- Isolate sensitive operations with clear **permission boundaries**.
- Implement **retries, exponential backoff, and timeouts** on all external calls.
- Use **circuit breakers and rate limiting** for service protection.
- Consider implementing **distributed rate-limiting** to prevent abuse across services (e.g., using Redis).

### Testing

- Write **unit tests** using table-driven patterns and parallel execution.
- **Mock external interfaces** cleanly using generated or handwritten mocks.
- Separate **fast unit tests** from slower integration and E2E tests.
- Ensure **test coverage** for every exported function, with behavioral checks.
- Use tools like `go test -cover` to ensure adequate test coverage.

### Documentation and Standards

- Document public functions and packages with **GoDoc-style comments**.
- Provide concise **READMEs** for services and libraries.
- Maintain a `CONTRIBUTING.md` and `ARCHITECTURE.md` to guide team practices.
- Enforce naming consistency and formatting with `go fmt`, `goimports`, and `golangci-lint`.

### Observability with OpenTelemetry

- Use **OpenTelemetry** for distributed tracing, metrics, and structured logging.
- Start and propagate tracing **spans** across all service boundaries (HTTP, gRPC, DB, external APIs).
- Always attach `context.Context` to spans, logs, and metric exports.
- Use **otel.Tracer** for creating spans and **otel.Meter** for collecting metrics.
- Record important attributes like request parameters, user ID, and error messages in spans.
- Use **log correlation** by injecting trace IDs into structured logs.
- Export data to **OpenTelemetry Collector**, **Jaeger**, or **Prometheus**.

### Tracing and Monitoring Best Practices

- Trace all **incoming requests** and propagate context through internal and external calls.
- Use **middleware** to instrument HTTP and gRPC endpoints automatically.
- Annotate slow, critical, or error-prone paths with **custom spans**.
- Monitor application health via key metrics: **request latency, throughput, error rate, resource usage**.
- Define **SLIs** (e.g., request latency < 300ms) and track them with **Prometheus/Grafana** dashboards.
- Alert on key conditions (e.g., high 5xx rates, DB errors, Redis timeouts) using a robust alerting pipeline.
- Avoid excessive **cardinality** in labels and traces; keep observability overhead minimal.
- Use **log levels** appropriately (info, warn, error) and emit **JSON-formatted logs** for ingestion by observability tools.
- Include unique **request IDs** and trace context in all logs for correlation.

### Performance

- Use **benchmarks** to track performance regressions and identify bottlenecks.
- Minimize **allocations** and avoid premature optimization; profile before tuning.
- Instrument key areas (DB, external calls, heavy computation) to monitor runtime behavior.

### Concurrency and Goroutines

- Ensure safe use of **goroutines**, and guard shared state with channels or sync primitives.
- Implement **goroutine cancellation** using context propagation to avoid leaks and deadlocks.

### Tooling and Dependencies

- Rely on **stable, minimal third-party libraries**; prefer the standard library where feasible.
- Use **Go modules** for dependency management and reproducibility.
- Version-lock dependencies for deterministic builds.
- Integrate **linting, testing, and security checks** in CI pipelines.

### Key Conventions

1. Prioritize **readability, simplicity, and maintainability**.
2. Design for **change**: isolate business logic and minimize framework lock-in.
3. Emphasize clear **boundaries** and **dependency inversion**.
4. Ensure all behavior is **observable, testable, and documented**.
5. **Automate workflows** for testing, building, and deployment.

---

# PROJECT CORE IDEA

## Objective

Create a terminal-based chat application written in Go that functions similarly to OpenCode, but uses models served from a remote OpenAI-compatible HTTP API.

The application must allow:

- Dynamically selecting available models from the API server
- Sending prompts to the selected model
- Receiving streaming responses
- Generating and editing code
- Executing system commands safely through a tools system
- Blocking destructive system commands
- Asking for confirmation before executing any command

## Architecture

### Primary Language

- Go (CLI, orchestration, session handling, HTTP calls, tools manager)

### Heavy Components — Optional in Rust

- Command execution sandbox
- Advanced command parsing
- Security validation
- Permission control

Integration between Go and Rust may be done via:

- FFI
- Binary bridge (Rust executable invoked from Go)
- Internal gRPC

## API Integration

The application connects to a remote OpenAI-compatible API server. The base URL and default model are defined in the configuration file.

**Default configuration:**

```yaml
api_base_url: https://lmstudiomacmini.gse.com.co:2443/v1
default_model: qwen/qwen3-coder-30b
```

The application must:

1. Fetch available models from:

```
GET https://lmstudiomacmini.gse.com.co:2443/v1/models
```

2. Allow interactive model selection in the terminal.

3. Use the chat completions endpoint:

```
POST https://lmstudiomacmini.gse.com.co:2443/v1/chat/completions
```

4. Support:
- Streaming responses
- Configurable temperature
- Configurable system prompt
- Conversation history

## CLI Interface

The CLI must include:

- Interactive model selector
- Interactive prompt:

```
> _
```

- Internal commands:
  - `/model` → change model
  - `/clear` → clear context
  - `/tools` → list available tools
  - `/exit` → exit application

## Tools System

The model may invoke tools using structured JSON.

Example:

```json
{
  "tool": "write_file",
  "arguments": {
    "path": "main.go",
    "content": "package main..."
  }
}
```

### Minimum Required Tools

- `write_file`
- `read_file`
- `list_directory`
- `execute_command` (sandboxed)

## Security Rules

The system must:

1. Block dangerous commands such as:
   - `rm -rf /`
   - `mkfs`
   - `shutdown`
   - `reboot`
   - Fork bombs
   - Access outside the workspace directory

2. Allow execution only inside a sandbox directory.

3. Ask for confirmation before executing commands:

```
The model wants to execute: go run main.go
Confirm? (y/n)
```

4. Implement:
   - Whitelist of allowed commands
   - Execution timeout limits
   - Memory limits (if using Rust sandbox)

---

## Project Structure

```
/cmd
/internal
  /apiclient
  /chat
  /tools
  /sandbox
  /security
/pkg
main.go
```

If Rust is used:

```
/sandbox-rs
```

## Context Handling

- Maintain conversation history per session
- Automatically truncate context if token limits are exceeded
- Optionally save session to JSON file

## Advanced Features (Bonus)

- Multi-session support
- YAML configuration file
- Colored terminal output
- Basic autocomplete
- Controlled autonomous agent mode
- Structured logging

## Technical Requirements

- Go 1.22+
- Clean architecture
- Modular codebase
- Robust error handling
- Basic tests included

## Expected Deliverables

- Fully functional source code
- README with instructions
- Usage examples
- Architecture explanation
- Build script for Rust component (if included)