# k8x

[![CI](https://github.com/shankgan/k8x/workflows/CI/badge.svg)](https://github.com/shankgan/k8x/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/shankgan/k8x)](https://goreportcard.com/report/github.com/shankgan/k8x)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Agentic kubectl** - AI-powered Kubernetes operations

k8x is an intelligent CLI tool that acts as an AI-powered layer on top of kubectl. It helps you manage Kubernetes resources through natural language commands and provides intelligent troubleshooting capabilities.

## Features

- 🤖 **Natural Language Interface**: Ask questions about your cluster in plain English
- 🔄 **Autonomous Execution**: AI agent executes safe kubectl commands automatically
- 🔍 **Intelligent Diagnostics**: AI-powered troubleshooting and resource analysis
- 📚 **Command History**: Automatic tracking with undo support
- 🔌 **Multi-LLM Support**: OpenAI, Anthropic Claude, and other providers
- 🛡️ **Secure by Default**: Read-only mode with command filtering
- 🎯 **Context-Aware**: Understands cluster state and provides relevant suggestions

## Quick Start

### Installation

```bash
# Homebrew (recommended)
brew install shankgan/tap/k8x

# Download binary
curl -L https://github.com/shankgan/k8x/releases/latest/download/k8x_Linux_x86_64.tar.gz | tar xz
sudo mv k8x /usr/local/bin/

# Docker
docker run ghcr.io/shankgan/k8x:latest
```

### Setup

1. **Initialize configuration:**

   ```bash
   k8x config init
   ```

2. **Add your LLM API key to ~/.k8x/credentials:**

   ```bash
   vim ~/.k8x/credentials
   ```

3. **Start using k8x:**

   ```bash
   k8x -c "are all pods running?"
   # Or
   k8x run "are all pods running?"
   # Or
   k8x command "are all pods running?"
   ```

## Usage Examples

```bash
# Diagnose pod issues
k8x -c "Find all pods that are not ready and explain why"

# Resource analysis
k8x -c "Check resource usage across all namespaces"

# Service troubleshooting
k8x -c "Diagnose why my service endpoints are empty"

# View command history
k8x history list
```

## Documentation

- 📖 **[Complete Documentation](./docs/README.md)** - Installation, configuration, and usage guide
- 🔧 **[Shell Execution Tool](./docs/shell-execution-tool.md)** - Security model and command filtering
- 📚 **[Usage Examples](./examples/basic-usage.md)** - Common use cases and patterns
- 🧪 **[Testing Guide](./docs/testing.md)** - Unit and E2E testing documentation

## Development

> **For Developers**: See [Developer Documentation](./docs/README.md#development) for complete setup instructions.

### Quick Developer Setup

```bash
# Clone and setup
git clone https://github.com/shankgan/k8x.git && cd k8x
make deps && make build

# Configure and test
./build/k8x configure
./build/k8x -c "are all pods running?"
```

### Project Structure

```text
k8x/
├── cmd/                 # CLI commands (Cobra, 'run', 'command', '-c' all supported)
├── internal/
│   ├── config/         # Configuration management
│   ├── llm/            # LLM provider interfaces
│   └── history/        # Command history tracking
├── docs/               # Documentation
├── examples/           # Usage examples
└── test/e2e/           # End-to-end tests
```

### Testing

```bash
# Unit tests
make test

# E2E tests (requires kind, kubectl, Docker)
export OPENAI_API_KEY=your-api-key
make test-e2e
```

## Contributing

We welcome contributions! Please see [Developer Documentation](./docs/README.md#development) for guidelines.

## License

Apache License - see [LICENSE](LICENSE) file for details.
