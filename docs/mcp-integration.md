# Model Context Protocol (MCP) Integration

k8x supports the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) to extend its capabilities by integrating with external tools and services. This allows k8x to access additional tools beyond its built-in shell execution capabilities.

## Overview

MCP is a protocol that enables AI assistants to securely connect to external tools and data sources. k8x can act as both:

1. **MCP Client**: Connect to external MCP servers to access their tools
2. **MCP Server**: Expose k8x's shell execution tools to other MCP clients

## Configuration

MCP servers are configured in your `~/.k8x/config.yaml` file:

```yaml
mcp:
  enabled: true
  servers:
    filesystem:
      enabled: true
      command: "npx"
      args: ["@modelcontextprotocol/server-filesystem", "/path/to/directory"]
      description: "Filesystem operations"
      env:
        DEBUG: "mcp:*"
```

## Using k8x as an MCP Client

When MCP is enabled, k8x will automatically connect to configured MCP servers and make their tools available during AI sessions.

### Available MCP Servers

Popular MCP servers you can integrate with k8x:

- **Filesystem**: `@modelcontextprotocol/server-filesystem` - File operations
- **GitHub**: `@modelcontextprotocol/server-github` - GitHub repository access
- **PostgreSQL**: `@modelcontextprotocol/server-postgres` - Database operations
- **Brave Search**: `@modelcontextprotocol/server-brave-search` - Web search
- **Slack**: `@modelcontextprotocol/server-slack` - Slack integration

### Example Configuration

```yaml
mcp:
  enabled: true
  servers:
    filesystem:
      enabled: true
      command: "npx"
      args: ["@modelcontextprotocol/server-filesystem", "/tmp"]
      description: "Temporary file operations"

    github:
      enabled: true
      command: "npx"
      args: ["@modelcontextprotocol/server-github"]
      description: "GitHub repository access"
      env:
        GITHUB_PERSONAL_ACCESS_TOKEN: "ghp_your_token_here"
```

### Running with MCP Tools

When you run k8x with MCP enabled, it will show connected servers:

```bash
k8x run "Check if there are any config files in /tmp and analyze the pod issues"
```

Output:

```bash
🔌 Connecting to MCP servers...
✓ Connected to MCP server: filesystem
✓ Connected to MCP server: github
🔌 Connected to 2 MCP server(s)
🔧 Available tools: 3 (including 2 MCP tools)
```

## Using k8x as an MCP Server

k8x can expose its shell execution capabilities as an MCP server for other applications to use.

### Starting the MCP Server

```bash
k8x mcp serve
```

This will start k8x in MCP server mode, exposing its tools via stdio transport.

### Integrating with Other Applications

You can configure other MCP clients to use k8x as a tool provider:

```json
{
  "mcpServers": {
    "k8x": {
      "command": "k8x",
      "args": ["mcp", "serve"]
    }
  }
}
```

### Available Tools

When running as an MCP server, k8x exposes:

- `execute_shell_command`: Safe, read-only shell command execution
- All configured kubectl operations with your cluster context

## Management Commands

### List MCP Servers

```bash
k8x config mcp list
```

### Enable/Disable MCP

```bash
k8x config mcp enable
k8x config mcp disable
```

### Add MCP Server

```bash
k8x config mcp add filesystem npx @modelcontextprotocol/server-filesystem /tmp
```

### Remove MCP Server

```bash
k8x config mcp remove filesystem
```

## Security Considerations

1. **Command Restrictions**: k8x maintains its read-only safety restrictions even when using MCP tools
2. **Environment Variables**: Be careful with sensitive environment variables in MCP server configurations
3. **Network Access**: Some MCP servers may require network access - review their security implications
4. **File Access**: Filesystem MCP servers should be limited to specific directories

## Troubleshooting

### Connection Issues

If MCP servers fail to connect:

1. Verify the command and arguments are correct
2. Check that required dependencies are installed
3. Review environment variables
4. Check server logs for errors

### Tool Execution Failures

If MCP tools fail during execution:

1. Verify the MCP server is still connected
2. Check tool arguments match the expected schema
3. Review server-specific documentation
4. Check k8x logs for detailed error messages

## Examples

### Kubernetes + Filesystem Analysis

```bash
k8x run "List all pods, then check if there are any matching config files in /etc/kubernetes/"
```

This combines kubectl operations with filesystem access via MCP.

### GitHub + Kubernetes Correlation

```bash
k8x run "Check current deployment status and compare with recent GitHub commits to the main branch"
```

This combines cluster inspection with GitHub repository analysis.

## Advanced Configuration

### Custom MCP Servers

You can integrate any MCP-compatible server by specifying its command:

```yaml
mcp:
  servers:
    custom_server:
      enabled: true
      command: "/path/to/custom/mcp/server"
      args: ["--config", "/path/to/config.json"]
      description: "Custom business logic server"
      env:
        API_KEY: "your-api-key"
        LOG_LEVEL: "debug"
```

### Multiple k8x Instances

Run multiple k8x instances as MCP servers for different clusters:

```yaml
mcp:
  servers:
    k8x_prod:
      enabled: true
      command: "k8x"
      args: ["mcp", "serve"]
      description: "Production cluster access"
      env:
        KUBECONFIG: "/home/user/.kube/config-prod"

    k8x_staging:
      enabled: true
      command: "k8x"
      args: ["mcp", "serve"]
      description: "Staging cluster access"
      env:
        KUBECONFIG: "/home/user/.kube/config-staging"
```

This allows a single k8x session to operate across multiple Kubernetes clusters through MCP.
