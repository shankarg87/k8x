package mcp

import (
	"context"
	"fmt"
	"os"

	"k8x/internal/config"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPClient wraps the mcp-go client to provide our interface
type MCPClient struct {
	client     *client.Client
	serverInfo mcp.Implementation
	connected  bool
}

// Tool represents an MCP tool definition (alias to mcp-go type)
type Tool = mcp.Tool

// ToolCall represents a tool call to an MCP server
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of an MCP tool call
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError"`
}

// ContentBlock represents a content block in MCP responses
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
}

// Client represents an MCP client interface
type Client interface {
	// Connect establishes connection to the MCP server
	Connect(ctx context.Context) error

	// Disconnect closes the connection to the MCP server
	Disconnect() error

	// ListTools returns available tools from the MCP server
	ListTools(ctx context.Context) ([]Tool, error)

	// CallTool executes a tool on the MCP server
	CallTool(ctx context.Context, call ToolCall) (*ToolResult, error)

	// IsConnected returns true if client is connected to server
	IsConnected() bool

	// GetServerInfo returns information about the connected server
	GetServerInfo() ServerInfo
}

// ServerInfo contains information about an MCP server
type ServerInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// NewMCPStdioClient creates a new stdio-based MCP client using mcp-go
func NewMCPStdioClient(command string, env []string, args ...string) (*MCPClient, error) {
	c, err := client.NewStdioMCPClient(command, env, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdio client: %w", err)
	}

	return &MCPClient{
		client:    c,
		connected: false,
	}, nil
}

// NewMCPSSEClient creates a new Server-Sent Events based MCP client
func NewMCPSSEClient(baseURL string, options ...transport.ClientOption) (*MCPClient, error) {
	c, err := client.NewSSEMCPClient(baseURL, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSE client: %w", err)
	}

	return &MCPClient{
		client:    c,
		connected: false,
	}, nil
}

// NewMCPStreamableHTTPClient creates a new streamable HTTP-based MCP client
func NewMCPStreamableHTTPClient(baseURL string, options ...transport.StreamableHTTPCOption) (*MCPClient, error) {
	c, err := client.NewStreamableHttpClient(baseURL, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create streamable HTTP client: %w", err)
	}

	return &MCPClient{
		client:    c,
		connected: false,
	}, nil
}

// NewMCPOAuthSSEClient creates a new OAuth-authenticated SSE-based MCP client
func NewMCPOAuthSSEClient(baseURL string, oauthConfig client.OAuthConfig, options ...transport.ClientOption) (*MCPClient, error) {
	c, err := client.NewOAuthSSEClient(baseURL, oauthConfig, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth SSE client: %w", err)
	}

	return &MCPClient{
		client:    c,
		connected: false,
	}, nil
}

// NewMCPOAuthStreamableHTTPClient creates a new OAuth-authenticated streamable HTTP-based MCP client
func NewMCPOAuthStreamableHTTPClient(baseURL string, oauthConfig client.OAuthConfig, options ...transport.StreamableHTTPCOption) (*MCPClient, error) {
	c, err := client.NewOAuthStreamableHttpClient(baseURL, oauthConfig, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth streamable HTTP client: %w", err)
	}

	return &MCPClient{
		client:    c,
		connected: false,
	}, nil
}

// Connect establishes connection to the MCP server
func (c *MCPClient) Connect(ctx context.Context) error {
	if c.connected {
		return nil
	}

	// Initialize the connection
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "k8x",
		Version: "1.0.0",
	}
	initReq.Params.Capabilities = mcp.ClientCapabilities{}

	result, err := c.client.Initialize(ctx, initReq)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	c.serverInfo = result.ServerInfo
	c.connected = true
	return nil
}

// Disconnect closes the connection to the MCP server
func (c *MCPClient) Disconnect() error {
	if !c.connected {
		return nil
	}

	err := c.client.Close()
	c.connected = false
	return err
}

// IsConnected returns true if client is connected to server
func (c *MCPClient) IsConnected() bool {
	return c.connected
}

// GetServerInfo returns information about the connected server
func (c *MCPClient) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        c.serverInfo.Name,
		Version:     c.serverInfo.Version,
		Description: "", // mcp-go doesn't have description in Implementation
	}
}

// ListTools returns available tools from the MCP server
func (c *MCPClient) ListTools(ctx context.Context) ([]Tool, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to MCP server")
	}

	req := mcp.ListToolsRequest{}
	result, err := c.client.ListTools(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return result.Tools, nil
}

// CallTool executes a tool on the MCP server
func (c *MCPClient) CallTool(ctx context.Context, call ToolCall) (*ToolResult, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to MCP server")
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = call.Name
	req.Params.Arguments = call.Arguments

	result, err := c.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}

	// Convert mcp-go result to our format
	var contentBlocks []ContentBlock
	for _, content := range result.Content {
		// Handle different content types
		switch c := content.(type) {
		case mcp.TextContent:
			contentBlocks = append(contentBlocks, ContentBlock{
				Type: c.Type,
				Text: c.Text,
			})
		case mcp.ImageContent:
			contentBlocks = append(contentBlocks, ContentBlock{
				Type: c.Type,
				Data: c.Data,
			})
		case mcp.AudioContent:
			contentBlocks = append(contentBlocks, ContentBlock{
				Type: c.Type,
				Data: c.Data,
			})
		default:
			// Fallback for unknown content types
			contentBlocks = append(contentBlocks, ContentBlock{
				Type: "text",
				Text: fmt.Sprintf("%v", content),
			})
		}
	}

	return &ToolResult{
		Content: contentBlocks,
		IsError: result.IsError,
	}, nil
}

// Manager manages multiple MCP clients
type Manager struct {
	clients map[string]Client
}

// NewManager creates a new MCP manager
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]Client),
	}
}

// RegisterClient registers an MCP client with a given name
func (m *Manager) RegisterClient(name string, client Client) {
	m.clients[name] = client
}

// GetClient returns an MCP client by name
func (m *Manager) GetClient(name string) (Client, error) {
	client, exists := m.clients[name]
	if !exists {
		return nil, fmt.Errorf("MCP client '%s' not found", name)
	}
	return client, nil
}

// ListClients returns names of all registered clients
func (m *Manager) ListClients() []string {
	var names []string
	for name := range m.clients {
		names = append(names, name)
	}
	return names
}

// ConnectAll connects to all registered MCP servers
func (m *Manager) ConnectAll(ctx context.Context) error {
	for name, client := range m.clients {
		if err := client.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to MCP server '%s': %w", name, err)
		}
	}
	return nil
}

// DisconnectAll disconnects from all MCP servers
func (m *Manager) DisconnectAll() error {
	var lastErr error
	for _, client := range m.clients {
		if err := client.Disconnect(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// GetAllTools returns all tools from all connected MCP servers
func (m *Manager) GetAllTools(ctx context.Context) (map[string][]Tool, error) {
	allTools := make(map[string][]Tool)

	for name, client := range m.clients {
		if !client.IsConnected() {
			continue
		}

		tools, err := client.ListTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tools from MCP server '%s': %w", name, err)
		}

		allTools[name] = tools
	}

	return allTools, nil
}

// CallTool calls a tool on the specified MCP server
func (m *Manager) CallTool(ctx context.Context, serverName string, call ToolCall) (*ToolResult, error) {
	client, err := m.GetClient(serverName)
	if err != nil {
		return nil, err
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("MCP server '%s' is not connected", serverName)
	}

	return client.CallTool(ctx, call)
}

// CreateFromConfig creates an MCP manager from configuration
func CreateManagerFromConfig(cfg *config.Config) (*Manager, error) {
	manager := NewManager()

	if !cfg.MCP.Enabled {
		return manager, nil
	}

	for name, serverConfig := range cfg.MCP.Servers {
		if !serverConfig.Enabled {
			continue
		}

		client, err := createClientFromConfig(serverConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create MCP client for server '%s': %w", name, err)
		}

		manager.RegisterClient(name, client)
	}

	return manager, nil
}

// createClientFromConfig creates an MCP client from server config
func createClientFromConfig(serverConfig config.MCPServerConfig) (Client, error) {
	// Default to stdio transport if not specified
	transportType := serverConfig.Transport
	if transportType == "" {
		transportType = "stdio"
	}

	switch transportType {
	case "stdio":
		if serverConfig.Command == "" {
			return nil, fmt.Errorf("MCP server command is required for stdio transport")
		}
		return NewMCPStdioClient(serverConfig.Command, mapToEnvSlice(serverConfig.Env), serverConfig.Args...)

	case "sse":
		if serverConfig.BaseURL == "" {
			return nil, fmt.Errorf("MCP server base_url is required for SSE transport")
		}
		return NewMCPSSEClient(serverConfig.BaseURL)

	case "http", "streamable-http":
		if serverConfig.BaseURL == "" {
			return nil, fmt.Errorf("MCP server base_url is required for HTTP transport")
		}
		return NewMCPStreamableHTTPClient(serverConfig.BaseURL)

	case "oauth-sse":
		if serverConfig.BaseURL == "" {
			return nil, fmt.Errorf("MCP server base_url is required for OAuth SSE transport")
		}
		if serverConfig.OAuth == nil {
			return nil, fmt.Errorf("OAuth configuration is required for OAuth SSE transport")
		}
		oauthConfig := client.OAuthConfig{
			ClientID:              serverConfig.OAuth.ClientID,
			ClientSecret:          serverConfig.OAuth.ClientSecret,
			RedirectURI:           serverConfig.OAuth.RedirectURI,
			Scopes:                serverConfig.OAuth.Scopes,
			AuthServerMetadataURL: serverConfig.OAuth.AuthServerMetadataURL,
			PKCEEnabled:           serverConfig.OAuth.PKCEEnabled,
			TokenStore:            transport.NewMemoryTokenStore(),
		}
		return NewMCPOAuthSSEClient(serverConfig.BaseURL, oauthConfig)

	case "oauth-http", "oauth-streamable-http":
		if serverConfig.BaseURL == "" {
			return nil, fmt.Errorf("MCP server base_url is required for OAuth HTTP transport")
		}
		if serverConfig.OAuth == nil {
			return nil, fmt.Errorf("OAuth configuration is required for OAuth HTTP transport")
		}
		oauthConfig := client.OAuthConfig{
			ClientID:              serverConfig.OAuth.ClientID,
			ClientSecret:          serverConfig.OAuth.ClientSecret,
			RedirectURI:           serverConfig.OAuth.RedirectURI,
			Scopes:                serverConfig.OAuth.Scopes,
			AuthServerMetadataURL: serverConfig.OAuth.AuthServerMetadataURL,
			PKCEEnabled:           serverConfig.OAuth.PKCEEnabled,
			TokenStore:            transport.NewMemoryTokenStore(),
		}
		return NewMCPOAuthStreamableHTTPClient(serverConfig.BaseURL, oauthConfig)

	default:
		return nil, fmt.Errorf("unsupported MCP transport type: %s", transportType)
	}
}

// mapToEnvSlice converts a map of environment variables to a slice format
func mapToEnvSlice(envMap map[string]string) []string {
	if envMap == nil {
		return nil
	}

	env := make([]string, 0, len(envMap))
	for key, value := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}

// ConnectAllWithConfig connects all MCP clients in the manager with configuration
func (m *Manager) ConnectAllWithConfig(ctx context.Context, serverConfigs map[string]config.MCPServerConfig) error {
	for _, serverName := range m.ListClients() {
		client, err := m.GetClient(serverName)
		if err != nil {
			continue
		}

		// Set environment variables if specified
		if serverConfig, exists := serverConfigs[serverName]; exists {
			if err := setEnvironmentVariables(serverConfig.Env); err != nil {
				return fmt.Errorf("failed to set environment variables for MCP server '%s': %w", serverName, err)
			}
		}

		if err := client.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to MCP server '%s': %w", serverName, err)
		}
	}

	return nil
}

// setEnvironmentVariables sets environment variables for MCP server execution
func setEnvironmentVariables(env map[string]string) error {
	for key, value := range env {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}
	return nil
}
