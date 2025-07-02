# k8x

[![CI](https://github.com/shankgan/k8x/workflows/CI/badge.svg)](https://github.com/shankgan/k8x/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/shankgan/k8x)](https://goreportcard.com/report/github.com/shankgan/k8x)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Agentic kubectl** - AI-powered Kubernetes operations

k8x is an intelligent CLI tool that acts as an AI-powered layer on top of kubectl. It helps you manage Kubernetes resources through natural language commands, provides automated assistance for common operations, and offers intelligent troubleshooting capabilities.

## Features

- 🤖 **Natural Language Interface**: Ask questions about your cluster in plain English
- � **Autonomous Execution**: AI agent can execute safe kubectl commands automatically
- �🔍 **Intelligent Diagnostics**: AI-powered troubleshooting and resource analysis  
- 📚 **Command History**: Automatic tracking of all operations with undo support
- 🔌 **Multi-LLM Support**: Works with OpenAI, Anthropic Claude, and other LLM providers
- 🛡️ **Secure by Default**: Read-only mode, command filtering, and local credential storage
- 🎯 **Context-Aware**: Understands your cluster state and provides relevant suggestions
- ⚡ **Tool Integration**: Built-in shell execution tool for seamless kubectl operations

## Quick Start

### Installation

```bash
# Homebrew (recommended)
brew install shankgan/tap/k8x

# Download binary from releases
curl -L https://github.com/shankgan/k8x/releases/latest/download/k8x_Linux_x86_64.tar.gz | tar xz
sudo mv k8x /usr/local/bin/

# Docker
docker run ghcr.io/shankgan/k8x:latest
```


### Setup

1. Initialize configuration:
```bash
k8x config init
```

2. Configure your LLM provider:
```bash
# Edit ~/.k8x/credentials and add your API keys
vim ~/.k8x/credentials
```

3. Start using k8x:
```bash
# Goal-oriented autonomous execution
k8x run "Diagnose why my nginx pod is failing"
k8x run "List all pods and show their resource usage"
k8x run "Check if any services are not receiving traffic"

# Interactive mode (coming soon)
k8x interactive
```

## Usage Examples

```bash
# Goal-oriented autonomous execution
k8x run "Find all pods that are not ready and explain why"
k8x run "Check resource usage across all namespaces"
k8x run "Diagnose why my service endpoints are empty"
k8x run "List all failed deployments and their error messages"

# View session history
k8x history

# Future commands (in development)
k8x ask "Which pods are using the most memory?"
k8x diagnose deployment my-app
k8x interactive
```

## Documentation

- [Full Documentation](./docs/README.md)
- [Shell Execution Tool](./docs/shell-execution-tool.md)
- [Usage Examples](./examples/basic-usage.md)
- [Configuration Guide](./docs/configuration.md)
- [Contributing](./CONTRIBUTING.md)

## Development

### Prerequisites

- Go 1.21+
- kubectl configured with cluster access
- LLM provider API key (OpenAI, Anthropic, etc.)
- [pre-commit](https://pre-commit.com/) for code linting and formatting

### Developer Setup & Example Workflow

```bash
# Clone the repository
git clone https://github.com/shankgan/k8x.git
cd k8x

# Install dependencies
make deps

# (Optional) Install pre-commit hooks
pre-commit install --install-hooks

# Run all pre-commit hooks manually
pre-commit run --all-files

# Build the binary
make build

# Run initial configuration (interactive)
./build/k8x configure

# Run a goal-oriented session
./build/k8x run "are all arcade ai pods running?"

# List session history
./build/k8x history list
```

#### Pre-commit Hooks Used

The repository uses the following pre-commit hooks (see `.pre-commit-config.yaml`):
- `golangci-lint` – Go code linting
- `go-fmt` – Code formatting
- `go-mod-tidy` – Ensure `go.mod`/`go.sum` are tidy
- `trailing-whitespace` – Remove trailing whitespace
- `end-of-file-fixer` – Ensure files end with a newline

To add or update hooks, edit `.pre-commit-config.yaml` and re-run `pre-commit install`.

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

### Development workflow

```bash
make dev
```

### Project Structure

```
k8x/
├── cmd/                 # CLI commands (Cobra)
├── internal/
│   ├── config/         # Configuration management  
│   ├── llm/            # LLM provider interfaces
│   └── history/        # Command history tracking
├── docs/               # Documentation
├── examples/           # Usage examples
└── .github/            # CI/CD workflows
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
