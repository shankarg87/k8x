# k8x - GenAI-powered Kubernetes operations

k8x is an intelligent CLI tool that acts as an AI-powered layer on top of `kubectl`. It helps you manage Kubernetes resources through natural language commands and provides intelligent troubleshooting capabilities.

[![CI](https://github.com/shankarg87/k8x/workflows/CI/badge.svg)](https://github.com/shankarg87/k8x/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/shankarg87/k8x)](https://goreportcard.com/report/github.com/shankarg87/k8x)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Release v0.1.1

`v0.1.1` is the first release for `k8x`. It performs read only operations on your kubernetes cluster
using `kubectl`, and is super effective at running multi-step checks and diagnostics over
different deployments/services/ingress/etc in your kubernetes cluster.

## Features

- 🤖 **Natural Language Interface**: Ask questions about your cluster in plain English
- 🔄 **Autonomous Multi-step Execution**: AI agent executes safe kubectl commands automatically
- 🔌 **Multi-LLM Support**: OpenAI, Anthropic Claude, and other providers
- 🔍 **Intelligent Diagnostics**: AI-powered troubleshooting and resource analysis
- 🛡️ **Secure by Default**: Read-only mode with command filtering
- 📚 **Command History**: Automatic tracking with undo support
- 🎯 **Context-Aware**: Understands cluster state and provides relevant suggestions

## Quick Start

### Installation

```bash
# Homebrew (recommended)
brew tap aihero/k8x
brew install k8x

k8x --version # v0.1.1
```

### Setup

1. **Initialize configuration:**

   ```bash
   k8x configure
   ```

1. **When prompted, choose your LLM provider and add your LLM API key**

   ```text
   Select your preferred LLM provider:
   1. OpenAI
   2. Anthropic
   3. Google (Gemini API)
   Enter choice [1-3]: 2
   Enter your Anthropic API key:
   ```

   > NOTE: The configuration and key is saved in `~/.k8x/credentials`.

1. **Start using k8x:**

   ```bash
   k8x -c "are all pods running?"
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

### Upgrades

   ```bash
   brew update          # fetches the latest tap and core metadata
   brew upgrade k8x     # upgrades only k8x (leaving other formulae untouched)
   ```

## Developer Documentation

> **For Developers**: See [Developer Documentation](./docs/README.md#development) for complete setup instructions.

## Contributing

We welcome contributions! Please see [Developer Documentation](./docs/README.md#development) for guidelines.

## License

Apache License - see [LICENSE](LICENSE) file for details.
