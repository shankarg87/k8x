package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"k8x/internal/config"
	"k8x/internal/mcp"

	mcpTypes "github.com/mark3labs/mcp-go/mcp"
)

// MCPToolManager extends ToolManager with MCP server integration
type MCPToolManager struct {
	*ToolManager
	mcpManager *mcp.Manager
	config     *config.Config
}

// NewMCPToolManager creates a new MCP-aware tool manager
func NewMCPToolManager(workDir string, cfg *config.Config) (*MCPToolManager, error) {
	baseManager := NewToolManager(workDir)

	// Create MCP manager and configure clients from config
	mcpManager, err := mcp.CreateManagerFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP manager from config: %w", err)
	}

	return &MCPToolManager{
		ToolManager: baseManager,
		mcpManager:  mcpManager,
		config:      cfg,
	}, nil
}

// ConnectMCPServers connects to all configured MCP servers
func (mtm *MCPToolManager) ConnectMCPServers(ctx context.Context) error {
	if !mtm.config.MCP.Enabled {
		return nil
	}

	if err := mtm.mcpManager.ConnectAll(ctx); err != nil {
		log.Printf("[DEBUG] Error connecting to MCP servers: %v", err)
		return err
	}
	return nil
}

// DisconnectMCPServers disconnects from all MCP servers
func (mtm *MCPToolManager) DisconnectMCPServers() error {
	return mtm.mcpManager.DisconnectAll()
}

// GetAllTools returns all available tools including MCP tools
func (mtm *MCPToolManager) GetAllTools(ctx context.Context) ([]Tool, error) {
	// Start with base shell tools
	tools := mtm.GetTools()

	// Add MCP tools if enabled
	if mtm.config.MCP.Enabled {
		mcpTools, err := mtm.getMCPTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get MCP tools: %w", err)
		}
		for _, tool := range mcpTools {
			log.Printf("[DEBUG] MCP tool: %s - %s", tool.Function.Name, tool.Function.Description)
		}
		tools = append(tools, mcpTools...)
	}

	return tools, nil
}

// getMCPTools converts MCP tools to LLM tools
func (mtm *MCPToolManager) getMCPTools(ctx context.Context) ([]Tool, error) {
	allMCPTools, err := mtm.mcpManager.GetAllTools(ctx)
	if err != nil {
		return nil, err
	}

	var tools []Tool
	for serverName, serverTools := range allMCPTools {
		for _, mcpTool := range serverTools {
			tool := mtm.convertMCPTool(serverName, mcpTool)
			tools = append(tools, tool)
		}
	}

	return tools, nil
}

// convertMCPTool converts an MCP tool to an LLM tool
func (mtm *MCPToolManager) convertMCPTool(serverName string, mcpTool mcp.Tool) Tool {
	// Create a unique tool name by prefixing with server name
	toolName := fmt.Sprintf("mcp_%s_%s", serverName, mcpTool.Name)

	// Convert MCP input schema to LLM tool parameters
	parameters := mtm.convertInputSchema(mcpTool.InputSchema)

	log.Printf("Loaded MCP tool: %s (server: %s)", toolName, serverName) // Debug log

	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        toolName,
			Description: fmt.Sprintf("[MCP:%s] %s", serverName, mcpTool.Description),
			Parameters:  parameters,
		},
		Handler: func(args string) (string, error) {
			return mtm.executeMCPTool(context.Background(), serverName, mcpTool.Name, args)
		},
	}
}

// convertInputSchema converts MCP input schema to LLM tool parameters
func (mtm *MCPToolManager) convertInputSchema(inputSchema mcpTypes.ToolInputSchema) ToolParameters {
	params := ToolParameters{
		Type:       inputSchema.Type,
		Properties: make(map[string]ToolParameterSpec),
		Required:   inputSchema.Required,
	}

	// Convert properties from map[string]any to our format
	for propName, propDef := range inputSchema.Properties {
		if propDefMap, ok := propDef.(map[string]interface{}); ok {
			spec := ToolParameterSpec{}

			if propType, ok := propDefMap["type"].(string); ok {
				spec.Type = propType
			}

			if description, ok := propDefMap["description"].(string); ok {
				spec.Description = description
			}

			if enumVal, ok := propDefMap["enum"].([]interface{}); ok {
				for _, e := range enumVal {
					if eStr, ok := e.(string); ok {
						spec.Enum = append(spec.Enum, eStr)
					}
				}
			}

			params.Properties[propName] = spec
		}
	}

	return params
}

// executeMCPTool executes an MCP tool
func (mtm *MCPToolManager) executeMCPTool(ctx context.Context, serverName, toolName, arguments string) (string, error) {
	// Parse arguments JSON
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %w", err)
	}

	// Create MCP tool call
	toolCall := mcp.ToolCall{
		Name:      toolName,
		Arguments: args,
	}

	// Execute the tool
	result, err := mtm.mcpManager.CallTool(ctx, serverName, toolCall)
	if err != nil {
		return "", fmt.Errorf("MCP tool execution failed: %w", err)
	}

	if result.IsError {
		return "", fmt.Errorf("MCP tool returned error: %s", mtm.formatMCPResult(result))
	}

	return mtm.formatMCPResult(result), nil
}

// formatMCPResult formats MCP tool result for display
func (mtm *MCPToolManager) formatMCPResult(result *mcp.ToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}

	var parts []string
	for _, block := range result.Content {
		switch block.Type {
		case "text":
			parts = append(parts, block.Text)
		case "data":
			parts = append(parts, block.Data)
		default:
			parts = append(parts, fmt.Sprintf("[%s content]", block.Type))
		}
	}

	return strings.Join(parts, "\n")
}

// ExecuteTool executes a tool by name, handling both shell and MCP tools
func (mtm *MCPToolManager) ExecuteTool(name, arguments string) (string, error) {
	// Check if it's an MCP tool
	if strings.HasPrefix(name, "mcp_") {
		// Extract server name and tool name
		parts := strings.SplitN(name, "_", 3)
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid MCP tool name format: %s", name)
		}

		serverName := parts[1]
		toolName := parts[2]

		return mtm.executeMCPTool(context.Background(), serverName, toolName, arguments)
	}

	// Fall back to shell tools using the embedded ToolManager
	return mtm.ToolManager.ExecuteTool(name, arguments)
}

// GetMCPServerStatus returns status of all MCP servers
func (mtm *MCPToolManager) GetMCPServerStatus() map[string]bool {
	status := make(map[string]bool)

	for _, serverName := range mtm.mcpManager.ListClients() {
		client, err := mtm.mcpManager.GetClient(serverName)
		if err != nil {
			status[serverName] = false
			continue
		}
		status[serverName] = client.IsConnected()
	}

	return status
}
