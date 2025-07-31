package mcp

import (
	"context"
	"strings"
	"testing"
	"time"

	"k8x/internal/config"
)

func TestMCPClientCreation(t *testing.T) {
	// Test that we can create an MCP client (without connecting)
	client, err := NewMCPStdioClient("echo", nil, "test")
	if err != nil {
		t.Fatalf("NewMCPStdioClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewMCPStdioClient returned nil client")
	}

	// Test initial state
	if client.IsConnected() {
		t.Error("Client should not be connected initially")
	}
}

func TestManagerCreation(t *testing.T) {
	// Test that we can create an MCP manager
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	// Test initial state
	clients := manager.ListClients()
	if len(clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(clients))
	}
}

func TestManagerClientRegistration(t *testing.T) {
	manager := NewManager()

	// Create a test client
	client, err := NewMCPStdioClient("echo", nil, "test")
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Register the client
	manager.RegisterClient("test", client)

	// Verify registration
	clients := manager.ListClients()
	if len(clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(clients))
	}

	if clients[0] != "test" {
		t.Errorf("Expected client name 'test', got '%s'", clients[0])
	}

	// Retrieve the client
	retrievedClient, err := manager.GetClient("test")
	if err != nil {
		t.Errorf("Failed to retrieve client: %v", err)
	}

	if retrievedClient != client {
		t.Error("Retrieved client does not match registered client")
	}
}

func TestCreateManagerFromConfig(t *testing.T) {
	// Test creating manager from configuration
	cfg := &config.Config{
		MCP: config.MCPConfig{
			Enabled: false,
			Servers: map[string]config.MCPServerConfig{},
		},
	}

	manager, err := CreateManagerFromConfig(cfg)
	if err != nil {
		t.Fatalf("CreateManagerFromConfig failed: %v", err)
	}

	if manager == nil {
		t.Fatal("CreateManagerFromConfig returned nil manager")
	}

	// Should have no clients when MCP is disabled
	clients := manager.ListClients()
	if len(clients) != 0 {
		t.Errorf("Expected 0 clients when MCP disabled, got %d", len(clients))
	}
}

func TestCreateManagerFromConfigWithServers(t *testing.T) {
	// Test creating manager with configured servers
	cfg := &config.Config{
		MCP: config.MCPConfig{
			Enabled: true,
			Servers: map[string]config.MCPServerConfig{
				"test-server": {
					Enabled: true,
					Command: "echo",
					Args:    []string{"hello"},
				},
				"disabled-server": {
					Enabled: false,
					Command: "echo",
					Args:    []string{"world"},
				},
			},
		},
	}

	manager, err := CreateManagerFromConfig(cfg)
	if err != nil {
		t.Fatalf("CreateManagerFromConfig failed: %v", err)
	}

	// Should have one client (only enabled server)
	clients := manager.ListClients()
	if len(clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(clients))
	}

	if clients[0] != "test-server" {
		t.Errorf("Expected client name 'test-server', got '%s'", clients[0])
	}
}

func TestMCPClientConnectionLifecycle(t *testing.T) {
	// This test verifies the connection lifecycle without actually connecting
	// to avoid requiring a real MCP server
	client, err := NewMCPStdioClient("echo", nil, "test")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Should start disconnected
	if client.IsConnected() {
		t.Error("Client should start disconnected")
	}

	// Test timeout context for connection attempts
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This will fail because "echo test" is not a valid MCP server,
	// but we're testing that the connection attempt is properly handled
	err = client.Connect(ctx)
	if err == nil {
		t.Error("Connect should fail with invalid server command")
	}

	// Should still be disconnected after failed connection
	if client.IsConnected() {
		t.Error("Client should remain disconnected after failed connection")
	}

	// Test disconnect on unconnected client
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect should not fail on unconnected client: %v", err)
	}
}

func TestMultipleTransportTypes(t *testing.T) {
	tests := []struct {
		name          string
		serverConfig  config.MCPServerConfig
		expectedError bool
		errorContains string
	}{
		{
			name: "stdio transport",
			serverConfig: config.MCPServerConfig{
				Transport: "stdio",
				Command:   "echo",
				Args:      []string{"hello"},
				Enabled:   true,
			},
			expectedError: false,
		},
		{
			name: "sse transport",
			serverConfig: config.MCPServerConfig{
				Transport: "sse",
				BaseURL:   "http://example.com/mcp",
				Enabled:   true,
			},
			expectedError: false,
		},
		{
			name: "http transport",
			serverConfig: config.MCPServerConfig{
				Transport: "http",
				BaseURL:   "http://example.com/mcp",
				Enabled:   true,
			},
			expectedError: false,
		},
		{
			name: "missing command for stdio",
			serverConfig: config.MCPServerConfig{
				Transport: "stdio",
				Enabled:   true,
			},
			expectedError: true,
			errorContains: "command is required",
		},
		{
			name: "missing base_url for sse",
			serverConfig: config.MCPServerConfig{
				Transport: "sse",
				Enabled:   true,
			},
			expectedError: true,
			errorContains: "base_url is required",
		},
		{
			name: "unsupported transport",
			serverConfig: config.MCPServerConfig{
				Transport: "websocket",
				Enabled:   true,
			},
			expectedError: true,
			errorContains: "unsupported MCP transport type",
		},
		{
			name: "default to stdio when transport not specified",
			serverConfig: config.MCPServerConfig{
				Command: "echo",
				Args:    []string{"hello"},
				Enabled: true,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := createClientFromConfig(tt.serverConfig)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if client == nil {
					t.Error("Expected client to be created but got nil")
				}
			}
		})
	}
}

func TestCreateManagerFromConfigWithMultipleTransports(t *testing.T) {
	cfg := &config.Config{
		MCP: config.MCPConfig{
			Enabled: true,
			Servers: map[string]config.MCPServerConfig{
				"stdio-server": {
					Transport: "stdio",
					Command:   "echo",
					Args:      []string{"hello"},
					Enabled:   true,
				},
				"sse-server": {
					Transport: "sse",
					BaseURL:   "http://example.com/sse",
					Enabled:   true,
				},
				"http-server": {
					Transport: "http",
					BaseURL:   "http://example.com/http",
					Enabled:   true,
				},
			},
		},
	}

	manager, err := CreateManagerFromConfig(cfg)
	if err != nil {
		t.Fatalf("CreateManagerFromConfig failed: %v", err)
	}

	clients := manager.ListClients()
	if len(clients) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(clients))
	}

	expectedClients := []string{"stdio-server", "sse-server", "http-server"}
	for _, expected := range expectedClients {
		found := false
		for _, actual := range clients {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find client '%s' but didn't", expected)
		}
	}
}
