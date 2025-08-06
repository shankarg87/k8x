# Model Context Protocol (MCP) Integration

k8x supports the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) to extend its capabilities by integrating with external tools and services. This allows k8x to access additional tools beyond its built-in shell execution capabilities.

## Overview

MCP is a protocol that enables AI assistants to securely connect to external tools and data sources. k8x can act as both:

1. **MCP Client**: Connect to external MCP servers to access their tools
2. **MCP Server**: Expose k8x's shell execution tools to other MCP clients

The MCP integration transforms k8x from a standalone Kubernetes assistant into a composable tool that can:

1. **Extend capabilities** by connecting to specialized MCP servers (filesystem, databases, APIs)
2. **Share expertise** by exposing Kubernetes knowledge to other applications
3. **Enable workflows** that span multiple tools and systems
4. **Maintain security** while preserving its core strengths in Kubernetes operations and AI-driven assistance

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

### Configuration Structure

```yaml
mcp:
  enabled: bool                           # Enable/disable MCP integration
  servers:
    <server_name>:
      enabled: bool                       # Enable/disable this server
      transport: string                   # Transport type (stdio, sse, http, oauth-sse, oauth-http)

      # Stdio transport configuration
      command: string                     # Command to run the server
      args: []string                      # Command arguments
      env: map[string]string              # Environment variables

      # HTTP/SSE transport configuration
      base_url: string                    # Base URL for HTTP/SSE transports

      # OAuth configuration (for oauth-sse and oauth-http transports)
      oauth:                              # OAuth configuration
        client_id: string
        client_secret: string
        auth_url: string
        token_url: string

      description: string                 # Human-readable description
```

## Supported Transports

The MCP integration supports multiple transport types:

1. **Stdio Transport** (`stdio`) - Default
   - Process-based communication using stdin/stdout
   - Requires: `command`, optional `args` and `env`

2. **Server-Sent Events** (`sse`)
   - HTTP-based communication using SSE
   - Requires: `base_url`

3. **Streamable HTTP** (`http` or `streamable-http`)
   - HTTP streaming transport
   - Requires: `base_url`

4. **OAuth SSE** (`oauth-sse`)
   - OAuth-authenticated SSE transport
   - Requires: `base_url`, `oauth` configuration

5. **OAuth HTTP** (`oauth-http` or `oauth-streamable-http`)
   - OAuth-authenticated HTTP streaming transport
   - Requires: `base_url`, `oauth` configuration

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

## Usage Scenarios

### Scenario 1: Kubernetes + Filesystem Analysis

```bash
k8x run "List all pods, then check if there are any matching config files in /etc/kubernetes/"
```

This combines kubectl operations with filesystem access via MCP.

### Scenario 2: GitHub + Kubernetes Correlation

```bash
k8x run "Check current deployment status and compare with recent GitHub commits to the main branch"
```

This combines cluster inspection with GitHub repository analysis.

### Scenario 3: Cross-Cluster Operations

When running `k8x run "Check pod logs and save to project directory"`, the AI can:

1. Use kubectl to get pod logs (via shell tools)
2. Use filesystem tools to write logs to files (via MCP)

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

## Implementation Details

### Core Architecture

The MCP implementation in k8x is built on the official mcp-go library (<https://github.com/mark3labs/mcp-go>), ensuring full protocol compliance and compatibility with the standard MCP ecosystem.

#### Key Components

1. **MCP Client Wrapper** (`internal/mcp/manager.go`)
   - Wraps mcp-go's client for seamless integration
   - Manages multiple MCP client connections
   - Handles connection lifecycle and error recovery
   - Provides tool name prefixing to avoid conflicts

2. **Tool Manager Integration** (`internal/llm/mcp_tools.go`)
   - Extends base ToolManager with MCP capabilities
   - Seamlessly integrates MCP tools with shell tools
   - Automatic schema conversion between MCP and LLM formats
   - Connection management for MCP servers

3. **Transport Support**
   - Factory functions for different transport types
   - Automatic transport selection based on configuration
   - Support for authentication and OAuth flows

### Technical Achievements

✅ **Protocol Compliance**

- Full MCP 2024-11-05 protocol implementation
- JSON-RPC 2.0 compliant communication
- Proper capability negotiation

✅ **Bidirectional Integration**

- k8x can consume external MCP tools
- k8x can provide tools to external MCP clients
- Seamless tool interoperability

✅ **Configuration-Driven**

- YAML-based MCP server configuration
- Environment variable support
- Enable/disable controls

✅ **Developer Experience**

- Comprehensive CLI commands
- Clear error messages and status reporting
- Extensive documentation and examples

## Security Considerations

1. **Command Restrictions**: MCP tools inherit k8x's read-only safety restrictions
2. **Environment Isolation**: MCP servers run in controlled environments
3. **Input Validation**: All MCP tool arguments are validated before execution
4. **Error Handling**: Secure error reporting without exposing sensitive information
5. **File Access**: Filesystem MCP servers should be limited to specific directories
6. **Network Access**: Some MCP servers may require network access - review their security implications

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

## Future Enhancements

### Potential Improvements

1. **GUI Configuration**: Web interface for MCP server management
2. **Auto-Discovery**: Automatic detection of available MCP servers
3. **Caching**: Tool and result caching for better performance
4. **Monitoring**: Health checks and performance metrics for MCP servers
5. **Authentication**: Enhanced support for authenticated MCP connections

### Integration Opportunities

1. **CI/CD Pipelines**: Use k8x MCP server in automated workflows
2. **IDE Extensions**: Direct k8x integration in development environments
3. **Dashboard Integration**: Embed k8x capabilities in Kubernetes dashboards
4. **ChatOps**: Use k8x as backend for Slack/Teams Kubernetes bots

## Testing & Validation

### Manual Testing Completed

- ✅ MCP server initialization and handshake
- ✅ Tool discovery via `tools/list`
- ✅ Tool execution via `tools/call`
- ✅ Configuration commands functionality
- ✅ Integration with existing k8x workflows

### Test Coverage

- Unit tests exist for core components
- Integration testing framework in place
- E2E tests validate full workflows

## Conclusion

The MCP integration positions k8x as a key component in the emerging MCP ecosystem while preserving its core strengths in Kubernetes operations and AI-driven assistance. This implementation provides a robust, secure, and extensible foundation for connecting k8x with the broader ecosystem of AI tools and services.
