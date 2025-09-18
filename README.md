# k8x - GenAI-powered Kubernetes operations

k8x is an intelligent CLI tool that acts as an AI-powered layer on top of `kubectl`. It helps you manage Kubernetes resources through natural language commands and provides intelligent troubleshooting capabilities.

[![CI](https://github.com/shankarg87/k8x/workflows/CI/badge.svg)](https://github.com/shankarg87/k8x/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/shankarg87/k8x)](https://goreportcard.com/report/github.com/shankarg87/k8x)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Release v0.2.0

`v0.2.0` is the first release for `k8x`. It performs read only operations on your kubernetes cluster
using `kubectl`, and is super effective at running multi-step checks and diagnostics over
different deployments/services/ingress/etc in your kubernetes cluster.

## Features

- 💬 **Interactive Console**: REPL-style interface with slash commands for continuous interaction
- 🤖 **Natural Language Interface**: Ask questions about your cluster in plain English
- 🔄 **Autonomous Multi-step Execution**: AI agent executes safe kubectl commands automatically
- 🔌 **Multi-LLM Support**: OpenAI, Anthropic Claude, and Google Gemini providers
- 🔍 **Intelligent Diagnostics**: AI-powered troubleshooting and resource analysis
- 🛡️ **Secure by Default**: Read-only mode with command filtering
- 📚 **Command History**: Automatic tracking with `.k8x` session files
- 🎯 **Context-Aware**: Understands cluster state and provides relevant suggestions
- 🔌 **MCP Integration**: Connect to external Model Context Protocol servers for extended capabilities
- 🔧 **MCP Server**: Can expose k8x's capabilities as an MCP server for other applications

## Quick Start

### Installation

```bash
# Homebrew (recommended)
brew tap aihero/k8x
brew install k8x

k8x --version # v0.2.0
```

### Getting Started

```bash
# 1. Launch k8x (will auto-configure on first run)
k8x

# 2. Follow the setup prompts to choose your LLM provider
# 3. Start asking questions about your cluster!
```

### Interactive Console

k8x features an interactive console similar to Claude Code. Simply run:

```bash
k8x
```

This will launch the interactive console where you can:

- Type natural language questions about your Kubernetes cluster
- Use slash commands for special operations
- Get immediate feedback and results

#### Console Commands

```text
/help, /h       - Show available commands
/configure, /f  - Configure k8x settings
/history, /x    - Show command history
/version, /v    - Show version information
/confirm        - Toggle confirmation mode
/mcp            - Show MCP server status
/clear, /cls    - Clear the screen
/exit, /q       - Exit the console
```

Example console session:

```text
╔════════════════════════════════════════════╗
║         Welcome to k8x Console! 🚀         ║
╚════════════════════════════════════════════╝
🤖 Using LLM provider: Anthropic

Type your Kubernetes questions in natural language.
Type /help for available commands or /exit to quit.

> Why is my nginx pod failing?

📋 Step 1:
💭 I'll help you diagnose why your nginx pod is failing...
```

### Setup

The first time you run k8x, it will automatically prompt you to configure your LLM provider:

1. **Run k8x:**

   ```bash
   k8x
   ```

2. **Follow the configuration prompts:**

   ```text
   🔧 k8x is not configured. Let's set it up now.
   Select your preferred LLM provider:
   1. OpenAI
   2. Anthropic
   3. Google (Gemini API)
   Enter choice [1-3]: 2
   Enter your Anthropic API key:
   ```

   > NOTE: The configuration and key is saved in `~/.k8x/credentials`.

#### Usage Examples

In the k8x console, you can type natural language commands:

```text
> Find all pods that are not ready and explain why

> Check resource usage across all namespaces

> Why is my ingress not working?

> Show me all services in the default namespace
```

Or use slash commands:

```text
> /history

> /configure

> /version
```

### One-Shot Command Mode

You can also run k8x in non-interactive mode for single commands:

```bash
# Execute a single command
k8x -c "are all pods running?"

# With confirmation mode
k8x -c -a "diagnose my failing deployment"
```

### Upgrade

```bash
brew update          # fetches the latest tap and core metadata
brew upgrade k8x     # upgrades only k8x (leaving other formulae untouched)
```

## Experimental MCP (Model Context Protocol) Integration

k8x supports the Model Context Protocol for extending capabilities:

```bash
# Configure MCP servers
k8x config mcp list                    # List configured MCP servers
k8x config mcp enable                  # Enable MCP integration
k8x config mcp add <server-name>       # Add a new MCP server

# Check MCP status in console
k8x
> /mcp                                 # Show MCP server connection status
```

## Developer Documentation

> **For Developers**: See [Developer Documentation](./docs/README.md#development) for complete setup instructions.

## Contributing

We welcome contributions! Please see [Developer Documentation](./docs/README.md#development) for guidelines.

## License

Apache License - see [LICENSE](LICENSE) file for details.
