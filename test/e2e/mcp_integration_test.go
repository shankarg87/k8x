package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"k8x/internal/config"
	"k8x/internal/mcp"
)

func TestDuckDuckGoMCPIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if Python and uv are available
	if err := checkPythonRequirements(); err != nil {
		t.Skipf("Skipping DuckDuckGo MCP integration test: %v", err)
	}

	// Install the DuckDuckGo MCP server
	if err := installDuckDuckGoServer(); err != nil {
		t.Fatalf("Failed to install DuckDuckGo MCP server: %v", err)
	}

	// Get the command and arguments for running the server
	command, args := getDuckDuckGoCommandAndArgs()

	// Create MCP client configuration
	serverConfig := config.MCPServerConfig{
		Transport: "stdio",
		Command:   command,
		Args:      args,
		Enabled:   true,
	}

	// Test the client functionality
	t.Run("ClientCreation", func(t *testing.T) {
		testClientCreation(t, serverConfig)
	})

	t.Run("ToolDiscovery", func(t *testing.T) {
		testToolDiscovery(t, serverConfig)
	})

	t.Run("SearchFunctionality", func(t *testing.T) {
		testSearchFunctionality(t, serverConfig)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, serverConfig)
	})
}

func checkPythonRequirements() error {
	// Check if Python is available
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err := exec.LookPath("python"); err != nil {
			return fmt.Errorf("Python is not available")
		}
	}

	// Check if uv is available, install if not
	if _, err := exec.LookPath("uv"); err != nil {
		fmt.Println("Installing uv...")
		cmd := exec.Command("pip", "install", "uv")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install uv: %w", err)
		}
	}

	return nil
}

func installDuckDuckGoServer() error {
	fmt.Println("Installing DuckDuckGo MCP server...")

	// Create a virtual environment
	venvCmd := exec.Command("uv", "venv", ".venv")
	venvCmd.Stdout = os.Stdout
	venvCmd.Stderr = os.Stderr

	if err := venvCmd.Run(); err != nil {
		return fmt.Errorf("failed to create virtual environment: %w", err)
	}

	// Install the package in the virtual environment
	cmd := exec.Command("uv", "pip", "install", "duckduckgo-mcp-server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install duckduckgo-mcp-server: %w", err)
	}

	fmt.Println("DuckDuckGo MCP server installed successfully")
	return nil
}

func getDuckDuckGoCommandAndArgs() (string, []string) {
	// Try different possible commands for running the server
	commands := []struct {
		cmd  string
		args []string
	}{
		{"uv", []string{"run", "duckduckgo-mcp-server"}},     // UV with run command
		{"uvx", []string{"duckduckgo-mcp-server"}},           // UVX preferred method
		{"duckduckgo-mcp-server", []string{}},                // Direct binary if installed globally
		{"python3", []string{"-m", "duckduckgo_mcp_server"}}, // Python module execution
		{"python", []string{"-m", "duckduckgo_mcp_server"}},  // Python fallback
	}

	for _, cmdInfo := range commands {
		if _, err := exec.LookPath(cmdInfo.cmd); err == nil {
			return cmdInfo.cmd, cmdInfo.args
		}
	}

	// Fallback - this will likely fail but provides a clear error
	return "python3", []string{"-m", "duckduckgo_mcp_server"}
}

func testClientCreation(t *testing.T, serverConfig config.MCPServerConfig) {
	client, err := mcp.NewMCPStdioClient(serverConfig.Command, nil, serverConfig.Args...)
	if err != nil {
		t.Fatalf("Failed to create MCP client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	if client.IsConnected() {
		t.Error("Client should not be connected initially")
	}
}

func testToolDiscovery(t *testing.T, serverConfig config.MCPServerConfig) {
	client, err := mcp.NewMCPStdioClient(serverConfig.Command, nil, serverConfig.Args...)
	if err != nil {
		t.Fatalf("Failed to create MCP client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to the server
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to MCP server: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			t.Logf("failed to disconnect MCP client: %v", err)
		}
	}()

	// List available tools
	tools, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	if len(tools) == 0 {
		t.Fatal("Expected at least one tool to be available")
	}

	// Check for expected tools
	expectedTools := []string{"search", "fetch_content"}
	foundTools := make(map[string]bool)

	for _, tool := range tools {
		foundTools[tool.Name] = true
		t.Logf("Found tool: %s - %s", tool.Name, tool.Description)
	}

	for _, expectedTool := range expectedTools {
		if !foundTools[expectedTool] {
			t.Errorf("Expected tool '%s' not found", expectedTool)
		}
	}
}

func testSearchFunctionality(t *testing.T, serverConfig config.MCPServerConfig) {
	client, err := mcp.NewMCPStdioClient(serverConfig.Command, nil, serverConfig.Args...)
	if err != nil {
		t.Fatalf("Failed to create MCP client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Connect to the server
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to MCP server: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			t.Logf("failed to disconnect MCP client: %v", err)
		}
	}()

	// Test search functionality
	searchCall := mcp.ToolCall{
		Name: "search",
		Arguments: map[string]interface{}{
			"query":       "golang programming",
			"max_results": 3,
		},
	}

	result, err := client.CallTool(ctx, searchCall)
	if err != nil {
		t.Fatalf("Failed to execute search tool: %v", err)
	}

	if result == nil {
		t.Fatal("Search result should not be nil")
	}

	if result != nil && result.IsError {
		t.Fatalf("Search returned error: %+v", result)
	}

	if len(result.Content) == 0 {
		t.Fatal("Search result should have content")
	}

	// Check that we got meaningful search results
	searchContent := ""
	for _, content := range result.Content {
		if content.Type == "text" {
			searchContent += content.Text
		}
	}

	if !strings.Contains(strings.ToLower(searchContent), "golang") &&
		!strings.Contains(strings.ToLower(searchContent), "go") {
		t.Errorf("Search results should contain relevant content about golang, got: %s", searchContent[:min(200, len(searchContent))])
	}

	t.Logf("Search test completed successfully, found %d content blocks", len(result.Content))
}

func testErrorHandling(t *testing.T, serverConfig config.MCPServerConfig) {
	client, err := mcp.NewMCPStdioClient(serverConfig.Command, nil, serverConfig.Args...)
	if err != nil {
		t.Fatalf("Failed to create MCP client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to the server
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to MCP server: %v", err)
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			t.Logf("failed to disconnect MCP client: %v", err)
		}
	}()

	// Test with invalid tool name
	invalidCall := mcp.ToolCall{
		Name: "invalid_tool",
		Arguments: map[string]interface{}{
			"query": "test",
		},
	}

	_, err = client.CallTool(ctx, invalidCall)
	if err == nil {
		t.Log("Note: Invalid tool call did not return an error - server may handle invalid tools differently")
	} else {
		t.Logf("Expected error received for invalid tool: %v", err)
	}

	// Test search with missing required parameter
	missingParamCall := mcp.ToolCall{
		Name:      "search",
		Arguments: map[string]interface{}{},
	}

	result, err := client.CallTool(ctx, missingParamCall)
	// This might succeed but return an error result, or fail - either is acceptable
	if err == nil && result != nil && !result.IsError {
		t.Log("Note: Search without query parameter did not return an error - server may provide default behavior")
	} else if err != nil {
		t.Logf("Expected error received for missing parameter: %v", err)
	} else if result != nil && result.IsError {
		t.Logf("Error result received for missing parameter as expected")
	}

	t.Logf("Error handling test completed successfully")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
