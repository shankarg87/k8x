package mcp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"k8x/internal/config"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ShellExecutor handles shell command execution for MCP
type ShellExecutor struct {
	allowedCommands []string
	workDir         string
	k8sConfig       *config.KubernetesConfig
}

// NewShellExecutor creates a new shell executor
func NewShellExecutor(workDir string, k8sConfig *config.KubernetesConfig) *ShellExecutor {
	allowedCommands := []string{
		"kubectl", "echo", "cat", "ls", "pwd", "whoami", "date", "uname", "which",
		"curl", "ping", "nslookup", "dig",
	}

	return &ShellExecutor{
		allowedCommands: allowedCommands,
		workDir:         workDir,
		k8sConfig:       k8sConfig,
	}
}

// Execute runs a shell command with safety checks
func (se *ShellExecutor) Execute(command string) (string, error) {
	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmdName := parts[0]
	args := parts[1:]

	// Check if command is allowed
	allowed := false
	for _, allowedCmd := range se.allowedCommands {
		if cmdName == allowedCmd {
			allowed = true
			break
		}
	}

	if !allowed {
		return "", fmt.Errorf("command '%s' is not allowed", cmdName)
	}

	// Execute command
	cmd := exec.Command(cmdName, args...)
	if se.workDir != "" {
		cmd.Dir = se.workDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// NewMCPServer creates a new MCP server with shell execution tools
func NewMCPServer(workDir string, k8sConfig *config.KubernetesConfig) *server.MCPServer {
	s := server.NewMCPServer(
		"k8x",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	executor := NewShellExecutor(workDir, k8sConfig)

	// Add shell execution tool
	shellTool := mcp.NewTool("shell_execute",
		mcp.WithDescription("Execute shell commands safely"),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("The shell command to execute"),
		),
	)

	s.AddTool(shellTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		command, err := req.RequireString("command")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid command parameter: %v", err)), nil
		}

		output, err := executor.Execute(command)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("command execution failed: %v", err)), nil
		}

		return mcp.NewToolResultText(output), nil
	})

	return s
}

// ServeStdio starts the MCP server using stdio transport
func ServeStdio(s *server.MCPServer) error {
	return server.ServeStdio(s)
}
