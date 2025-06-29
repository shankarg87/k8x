# k8x Documentation

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
k8x config init
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
k8x config init

# Get help for any command
k8x help

# List all pods with AI assistance
k8x ask "show me all pods in the default namespace"

# Troubleshoot a deployment
k8x diagnose deployment my-app

# Interactive mode
k8x interactive
```

### History and Undo

k8x automatically tracks command history and supports undo operations:

```bash
# View command history
k8x history list

# Undo the last operation
k8x undo

# Undo a specific operation
k8x undo <operation-id>
```

## Commands

- `k8x ask` - Ask questions about your Kubernetes cluster
- `k8x diagnose` - Diagnose issues with resources
- `k8x config` - Manage configuration
- `k8x history` - View and manage command history
- `k8x undo` - Undo previous operations
- `k8x interactive` - Start interactive mode

For detailed command documentation, see the individual command help or use `k8x help <command>`.
