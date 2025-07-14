# k8x Developer Documentation

This document provides comprehensive developer setup instructions, architecture details, and contribution guidelines for k8x.

## Development Environment Setup

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **kubectl** - Configured with cluster access for testing
- **LLM provider API key** - OpenAI, Anthropic, etc.
- **pre-commit** - For code linting and formatting: `brew install pre-commit`
- **Docker** - For E2E testing with kind clusters

### Developer Setup Workflow

```bash
# 1. Clone and navigate
git clone https://github.com/shankgan/k8x.git && cd k8x

# 2. Install dependencies
make deps

# 3. Set up pre-commit hooks
pre-commit install --install-hooks

# 4. Build the binary
make build

# 5. Configure k8x
./build/k8x configure

# 6. Test with a simple command
./build/k8x -c "are all pods running?"
```

### Development Commands

```bash
# Format, lint, and tidy (recommended before commits)
make dev

# Build binary
make build

# Run unit tests
make test

# Run E2E tests (requires API key)
export OPENAI_API_KEY=your-api-key
make test-e2e

# Run specific E2E test
make test-e2e-single TEST=TestCrashLoopBackoffDiagnosis

# Generate coverage report
make test-coverage
```

### Code Quality Tools

#### Pre-commit Hooks

The repository uses these pre-commit hooks (see `.pre-commit-config.yaml`):

- **golangci-lint** – Go code linting
- **go-fmt** – Code formatting
- **go-imports** – Import management
- **go-mod-tidy** – Module tidying
- **go-vet** – Go static analysis
- **trailing-whitespace** – Remove trailing whitespace
- **end-of-file-fixer** – Ensure files end with newlines
- **markdownlint** – Markdown formatting
- **yamlfmt** – YAML formatting

#### Formatting Setup

Ensure `goimports` is available in your PATH:

```bash
go install golang.org/x/tools/cmd/goimports@latest
export PATH="$PATH:$HOME/go/bin"  # Add to ~/.zshrc for persistence
```

## Architecture Overview

### Project Structure

```text
k8x/
├── build/                       # Compiled binary output
├── cmd/                         # CLI commands (Cobra framework)
│   ├── config.go               # Config management utilities
│   ├── configure.go            # `k8x configure` command
│   ├── history.go              # `k8x history` command
│   ├── root.go                 # Root cobra command and global flags
│   ├── run.go                  # `k8x -c` command (main functionality, also aliased as 'run' and 'command')
│   └── version.go              # `k8x version` command
├── internal/                    # Private packages
│   ├── config/                 # Configuration management
│   │   ├── config.go           # Configuration struct and file operations
│   │   ├── config_test.go      # Configuration tests
│   │   └── credentials.go      # LLM API key management
│   ├── history/                # Command history tracking
│   │   └── manager.go          # .k8x file creation and parsing
│   ├── llm/                    # LLM provider interfaces
│   │   ├── client.go           # LLM client interface definition
│   │   ├── client_test.go      # Client interface tests
│   │   ├── factory.go          # LLM client factory and provider selection
│   │   ├── tools.go            # LLM tool calling and prompt utilities
│   │   ├── tools_test.go       # LLM tools tests
│   │   └── providers/          # LLM provider implementations
│   │       ├── anthropic.go    # Claude/Anthropic client
│   │       ├── openai.go       # OpenAI GPT client
│   │       └── unified.go      # Common provider utilities
│   └── schemas/                # Data structure definitions
│       └── credentials.go      # Credential file format schemas
├── docs/                       # Documentation
│   ├── README.md               # This developer documentation
│   ├── shell-execution-tool.md # Security model and command filtering
│   └── testing.md              # Testing strategy and guidelines
├── examples/                   # Usage examples and sample configs
│   ├── basic-usage.md          # Common usage patterns
│   ├── config.yaml             # Sample configuration file
│   └── credentials             # Sample credentials file template
├── test/e2e/                   # End-to-end testing framework
│   ├── main_test.go            # E2E test setup and teardown
│   ├── run_e2e_test.go         # E2E test cases
│   ├── README.md               # E2E testing documentation
│   └── framework/              # E2E testing utilities
│       ├── cluster.go          # Kind cluster management
│       ├── ci_helpers.go       # CI-specific test utilities
│       └── scenarios/          # Kubernetes test scenarios
├── main.go                     # Application entry point
└── Makefile                    # Build and development tasks
```

### Key Components

#### CLI Framework (cmd/)

- Built with **Cobra** framework following POSIX standards
- Each command in separate file (`configure.go`, `run.go`, etc.)
- Global flags and configuration handled in `root.go`

#### Configuration Management (internal/config/)

- YAML-based configuration stored in `~/.k8x/`
- Secure credential storage separate from main config
- Environment-specific overrides supported

#### LLM Integration (internal/llm/)

- **Interface-based design** for multiple LLM providers
- **Factory pattern** for provider selection
- **Tool calling system** for safe kubectl execution
- Built-in **shell execution tool** with security filtering

#### Command History (internal/history/)

- Session tracking with `.k8x` files
- Undo/redo capability planning
- Timestamped command logs

## Testing Strategy

### Unit Tests

- **Location**: Co-located with source code (`*_test.go`)
- **Coverage**: All public functions and interfaces
- **Pattern**: Table-driven tests where applicable
- **Command**: `make test`

### End-to-End Tests

- **Location**: `test/e2e/`
- **Framework**: Custom testing framework with kind clusters
- **Scenarios**: Real Kubernetes problems (CrashLoopBackOff, ImagePullBackOff, etc.)
- **Requirements**: Docker, kind, kubectl, API key
- **Command**: `make test-e2e`

### Testing Commands

```bash
# Run all unit tests
make test

# Run unit tests with coverage
make test-coverage

# Run E2E tests (requires setup)
export OPENAI_API_KEY=your-api-key
make test-e2e

# Run specific E2E test with debugging
make test-e2e-single TEST=TestCrashLoopBackoffDiagnosis
```

See [testing.md](testing.md) for detailed testing documentation.

## Security Model

k8x implements a **secure-by-default** approach:

- **Read-only operations** - Only safe kubectl commands allowed
- **Command filtering** - Built-in allowlist of safe operations
- **Local credential storage** - API keys stored locally, never transmitted
- **No cluster modifications** - Write operations explicitly blocked

See [shell-execution-tool.md](shell-execution-tool.md) for detailed security documentation.

## Contributing Guidelines

### Code Style

- Follow **Go naming conventions**
- Use **kebab-case** for files, **snake_case** for packages, **camelCase** for variables
- Implement **explicit error handling** with wrapped errors
- Keep functions **small and single-purpose**
- Use **structured logging** (logrus/zap preferred)

### Contribution Workflow

1. **Fork** the repository
2. **Create feature branch** from main
3. **Write tests** for new functionality
4. **Run pre-commit checks**: `make dev`
5. **Submit pull request** with clear description

### Adding New Features

#### New CLI Commands

1. Create command file in `cmd/`
2. Follow Cobra patterns from existing commands
3. Add command to `root.go`
4. Include help text and examples

#### New LLM Providers

1. Implement `Client` interface in `internal/llm/providers/`
2. Add provider to factory in `factory.go`
3. Update credential schemas
4. Add provider-specific tests

#### New Test Scenarios

1. Add scenario to `test/e2e/framework/scenarios/`
2. Create test case in `test/e2e/run_e2e_test.go`
3. Update E2E documentation

### Release Process

- **Versioning**: Semantic versioning (semver)
- **Releases**: Automated via GitHub Actions
- **Distribution**: Homebrew, GitHub Releases, Docker

## Troubleshooting Development Issues

### Common Problems

#### Pre-commit Hook Failures

```bash
# Re-install hooks
pre-commit install --install-hooks

# Run specific hook
pre-commit run golangci-lint --all-files
```

#### E2E Test Failures

```bash
# Check requirements
kind version && kubectl version && docker version

# Clean up leftover clusters
kind get clusters | grep k8x | xargs -I {} kind delete cluster --name {}

# Run with preservation for debugging
make test-e2e-single TEST=TestName -preserve-on-failure
```

#### Build Issues

```bash
# Clean and rebuild
make clean && make deps && make build

# Check Go version
go version  # Should be 1.21+
```

### Getting Help

- **Issues**: [GitHub Issues](https://github.com/shankgan/k8x/issues)
- **Discussions**: [GitHub Discussions](https://github.com/shankgan/k8x/discussions)
- **Discord**: [Development Discord](https://discord.gg/k8x-dev)

---

## User Documentation

For end-user documentation, see the sections below or visit specific guides.

## Overview

k8x is an AI-powered CLI tool that acts as an intelligent layer on top of kubectl. It helps you manage Kubernetes resources through natural language commands and provides automated assistance for common operations.

## Installation

### Homebrew

```bash
brew install shankgan/tap/k8x
```

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/shankgan/k8x/releases).

### Docker

```bash
docker run ghcr.io/shankgan/k8x:latest
```

## Configuration

k8x stores its configuration in `~/.k8x/config.yaml`. You can initialize the configuration with:

```bash
k8x configure
```

### LLM Provider Setup

Configure your preferred LLM provider credentials in `~/.k8x/credentials`:

```yaml
openai:
  api_key: "your-openai-api-key"
anthropic:
  api_key: "your-anthropic-api-key"
```

## Usage

### Basic Commands

```bash
# Initialize configuration
k8x configure

# Get help for any command
k8x help

# List all pods with AI assistance
k8x -c "show me all pods in the default namespace"

# Troubleshoot a deployment
k8x -c "diagnose deployment my-app"

# Interactive mode (coming soon)
k8x interactive
```

### History and Undo

k8x automatically tracks command history and supports undo operations:

```bash
# View command history
k8x history list

# Undo the last operation (planned feature)
k8x undo

# Undo a specific operation (planned feature)
k8x undo <operation-id>
```

## Commands

- `k8x -c` - Execute goal-oriented AI sessions with kubectl
- `k8x -c --ask` or `k8x -c -a` - Execute with confirmation before each tool
- `k8x configure` - Manage configuration and credentials
- `k8x history` - View and manage command history
- `k8x version` - Show version information

For detailed command documentation, see the individual command help or use `k8x help <command>`.
