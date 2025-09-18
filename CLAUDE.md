# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, Test, and Development Commands

### Essential Commands

```bash
# Build the binary
make build

# Run unit tests
make test

# Run a single unit test
go test -v -run TestName ./path/to/package

# Run E2E tests (requires Docker, kind, and API key)
export OPENAI_API_KEY=your-api-key  # Or ANTHROPIC_API_KEY for Claude
make test-e2e-all

# Run a single E2E test with debugging
make test-e2e-single TEST=TestCrashLoopBackoffDiagnosis

# Development workflow - format, imports, vet, lint, mod-tidy
make dev

# Lint only
make lint

# Format code
make fmt

# Manage imports
make go-imports

# Run go vet
make vet

# Tidy modules
make mod-tidy

# Test with coverage
make test-coverage
```

### Installation and Setup

```bash
# Install dependencies
make deps

# Set up pre-commit hooks
pre-commit install --install-hooks

# Build and test locally
make build
./build/k8x configure
./build/k8x  # Launches interactive console by default
./build/k8x -c "are all pods running?"  # One-shot command mode
```

## Architecture Overview

k8x is an AI-powered CLI tool for Kubernetes operations built in Go. It acts as an intelligent layer on top of kubectl, allowing natural language interaction with Kubernetes clusters.

### Key Design Principles

- **Interface-based LLM abstraction**: Multiple providers (OpenAI, Anthropic, Google) implement a common `Provider` interface in `internal/llm/`
- **Secure-by-default**: Only read-only kubectl operations allowed via built-in command filtering in `ShellExecutor`
- **Tool-calling architecture**: LLMs execute kubectl commands through a structured tool system rather than generating raw shell commands
- **Session tracking**: Commands are logged to `.k8x` files for history and potential undo operations
- **Interactive console**: Default mode launches an interactive REPL with slash commands similar to Claude Code

### Core Components

#### CLI Framework (`cmd/`)

- Built with Cobra framework
- Main commands:
  - `console.go`: Interactive console (default when running `k8x`)
  - `run.go`: Execute single AI sessions (`k8x -c` or `k8x run`)
  - `configure.go`: Setup wizard for LLM providers
  - `history.go`: Command history management
  - `mcp.go`: Model Context Protocol server/client management
- Console slash commands: `/help`, `/configure`, `/history`, `/version`, `/confirm`, `/mcp`, `/clear`, `/exit`

#### LLM Integration (`internal/llm/`)

- `client.go`: Defines the `Provider` interface and `Client` for managing multiple providers
- `tools.go`: Implements `ShellExecutor` for safe kubectl operations and `ToolManager` for tool orchestration
- `executor.go`: Handles LLM execution flow with tool calls
- `providers/unified.go`: Unified provider wrapper supporting OpenAI, Anthropic, Google
- Tool system: Function calling with JSON schema for `execute_shell_command`

#### Configuration (`internal/config/`)

- YAML-based config in `~/.k8x/config.yaml`
- Credentials stored separately in `~/.k8x/credentials`
- Provider selection and API key management via `Credentials` struct
- Kubernetes context configuration support

#### Testing Framework (`test/e2e/`)

- Custom E2E testing with kind clusters
- Real Kubernetes failure scenarios (CrashLoopBackOff, ImagePullBackOff, etc.)
- Framework utilities in `test/e2e/framework/` for cluster management
- Test patterns: Table-driven tests for unit tests, scenario-based for E2E

### Module Structure

- Module name: `k8x`
- Go version: 1.23+ (toolchain 1.24.4)
- Main entry point: `main.go`
- Binary output: `build/k8x`
- Key dependencies:
  - `github.com/spf13/cobra`: CLI framework
  - `github.com/anthropics/anthropic-sdk-go`: Anthropic provider
  - `github.com/openai/openai-go`: OpenAI provider
  - `google.golang.org/genai`: Google Gemini provider
  - `github.com/mark3labs/mcp-go`: MCP protocol support
