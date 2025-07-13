package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"k8x/internal/config"
)

// Tool represents a function that can be called by the LLM
type Tool struct {
	Type     string                            `json:"type"`
	Function ToolFunction                      `json:"function"`
	Handler  func(args string) (string, error) `json:"-"`
}

// ToolFunction describes a tool function
type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

// ToolParameters describes the parameters for a tool function
type ToolParameters struct {
	Type       string                       `json:"type"`
	Properties map[string]ToolParameterSpec `json:"properties"`
	Required   []string                     `json:"required"`
}

// ToolParameterSpec describes a single parameter
type ToolParameterSpec struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolCall represents a tool call made by the LLM
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

// ShellExecutor handles shell command execution
type ShellExecutor struct {
	allowedCommands []string
	workDir         string
	k8sConfig       *config.KubernetesConfig
}

// NewShellExecutor creates a new shell executor with safety restrictions
func NewShellExecutor(workDir string) *ShellExecutor {
	// Define allowed kubectl and other safe read-only commands
	allowedCommands := []string{
		"kubectl",
		"helm",      // for Helm releases and chart information
		"kustomize", // for Kustomize version and operations
		"echo",
		"cat",
		"ls",
		"pwd",
		"whoami",
		"date",
		"uname",
		"which",
		"curl", // for health checks
		"ping", // for connectivity checks
		"nslookup",
		"dig",
	}

	return &ShellExecutor{
		allowedCommands: allowedCommands,
		workDir:         workDir,
	}
}

// SetKubernetesConfig sets the Kubernetes configuration for this executor
func (se *ShellExecutor) SetKubernetesConfig(k8sConfig *config.KubernetesConfig) {
	se.k8sConfig = k8sConfig
}

// Execute runs a shell command with safety checks
func (se *ShellExecutor) Execute(command string) (string, error) {
	// Parse command to check if it's allowed
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	baseCmd := parts[0]

	// Check if command is in allowed list
	allowed := false
	for _, allowedCmd := range se.allowedCommands {
		if baseCmd == allowedCmd {
			allowed = true
			break
		}
	}

	if !allowed {
		return "", fmt.Errorf("command '%s' is not allowed for security reasons. Allowed commands: %v", baseCmd, se.allowedCommands)
	}

	// Additional safety checks for kubectl
	if baseCmd == "kubectl" {
		if se.containsWriteOperations(command) {
			return "", fmt.Errorf("kubectl write operations are not allowed in read-only mode. Command: %s", command)
		}

		// Apply Kubernetes configuration if available
		if se.k8sConfig != nil {
			// Add context flag if specified
			if se.k8sConfig.Context != "" && !strings.Contains(command, "--context") {
				command = fmt.Sprintf("%s --context=%s", command, se.k8sConfig.Context)
			}

			// Add namespace flag if specified and not already in command
			if se.k8sConfig.Namespace != "" && !strings.Contains(command, "--namespace") && !strings.Contains(command, "-n ") {
				command = fmt.Sprintf("%s --namespace=%s", command, se.k8sConfig.Namespace)
			}
		}
	}

	// Additional safety checks for helm
	if baseCmd == "helm" {
		if se.containsHelmWriteOperations(command) {
			return "", fmt.Errorf("helm write operations are not allowed in read-only mode. Command: %s", command)
		}
	}

	// Additional safety checks for kustomize
	if baseCmd == "kustomize" {
		if se.containsKustomizeWriteOperations(command) {
			return "", fmt.Errorf("kustomize write operations are not allowed in read-only mode. Command: %s", command)
		}
	}

	// Set up environment for kubectl if kubeconfig path is specified
	env := os.Environ()
	if se.k8sConfig != nil && se.k8sConfig.KubeConfigPath != "" && (baseCmd == "kubectl" || baseCmd == "helm" || baseCmd == "kustomize") {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", se.k8sConfig.KubeConfigPath))
	}

	// Execute the command
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if se.workDir != "" {
		cmd.Dir = se.workDir
	}

	// Set the environment variables
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// containsWriteOperations checks if a kubectl command contains write operations
func (se *ShellExecutor) containsWriteOperations(command string) bool {
	writeOps := []string{
		"create", "apply", "delete", "patch", "replace", "edit",
		"scale", "annotate", "label", "expose", "set", "rollout",
		"drain", "cordon", "uncordon", "taint", "exec", "port-forward",
		"proxy", "attach", "cp",
	}

	lowerCmd := strings.ToLower(command)
	for _, op := range writeOps {
		if strings.Contains(lowerCmd, " "+op+" ") || strings.Contains(lowerCmd, " "+op) {
			return true
		}
	}
	return false
}

// containsHelmWriteOperations checks if a helm command contains write operations
func (se *ShellExecutor) containsHelmWriteOperations(command string) bool {
	writeOps := []string{
		"install", "upgrade", "uninstall", "delete", "create", "rollback",
		"plugin", "push", "pull", "registry",
	}
	repoWriteOps := []string{"add", "remove"}

	lowerCmd := strings.ToLower(command)

	// Handle "helm repo" commands separately
	if strings.HasPrefix(lowerCmd, "helm repo") {
		for _, op := range repoWriteOps {
			if strings.Contains(lowerCmd, " repo "+op+" ") || strings.HasSuffix(lowerCmd, " repo "+op) {
				return true
			}
		}
		return false
	}

	// Check for other write operations
	for _, op := range writeOps {
		if strings.Contains(lowerCmd, " "+op+" ") || strings.Contains(lowerCmd, " "+op) {
			return true
		}
	}
	return false
}

// containsKustomizeWriteOperations checks if a kustomize command contains write operations
func (se *ShellExecutor) containsKustomizeWriteOperations(command string) bool {
	writeOps := []string{
		"create", "edit", "fix", "localize", "cfg",
	}

	lowerCmd := strings.ToLower(command)
	// Allow only version and build commands (build is read-only, just outputs YAML)
	if strings.Contains(lowerCmd, " version") || strings.Contains(lowerCmd, " build") ||
		strings.Contains(lowerCmd, " --version") {
		return false
	}

	for _, op := range writeOps {
		if strings.Contains(lowerCmd, " "+op+" ") || strings.Contains(lowerCmd, " "+op) {
			return true
		}
	}
	return false
}

// GetShellExecutionTool returns the shell execution tool definition
func GetShellExecutionTool(executor *ShellExecutor) Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "execute_shell_command",
			Description: "Execute a safe, read-only shell command. Primarily used for kubectl get, describe, logs commands and other diagnostic operations.",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolParameterSpec{
					"command": {
						Type:        "string",
						Description: "The shell command to execute. Must be a safe, read-only command like 'kubectl get pods' or 'kubectl describe service myservice'",
					},
				},
				Required: []string{"command"},
			},
		},
		Handler: func(args string) (string, error) {
			var params struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", fmt.Errorf("failed to parse arguments: %w", err)
			}

			if params.Command == "" {
				return "", fmt.Errorf("command parameter is required")
			}

			return executor.Execute(params.Command)
		},
	}
}

// ToolManager manages available tools
type ToolManager struct {
	tools            map[string]Tool
	executor         *ShellExecutor
	confirmationMode bool
}

// SetKubernetesConfig sets the Kubernetes configuration for the shell executor
func (tm *ToolManager) SetKubernetesConfig(k8sConfig *config.KubernetesConfig) {
	tm.executor.SetKubernetesConfig(k8sConfig)
}

// SetConfirmationMode enables or disables user confirmation before tool execution
func (tm *ToolManager) SetConfirmationMode(confirm bool) {
	tm.confirmationMode = confirm
}

// NewToolManager creates a new tool manager
func NewToolManager(workDir string) *ToolManager {
	executor := NewShellExecutor(workDir)
	tm := &ToolManager{
		tools:            make(map[string]Tool),
		executor:         executor,
		confirmationMode: false,
	}

	// Register shell execution tool
	shellTool := GetShellExecutionTool(executor)
	tm.tools[shellTool.Function.Name] = shellTool

	return tm
}

// GetTools returns all available tools
func (tm *ToolManager) GetTools() []Tool {
	tools := make([]Tool, 0, len(tm.tools))
	for _, tool := range tm.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ExecuteTool executes a tool by name with given arguments
func (tm *ToolManager) ExecuteTool(name, arguments string) (string, error) {
	tool, exists := tm.tools[name]
	if !exists {
		return "", fmt.Errorf("tool '%s' not found", name)
	}

	// If confirmation mode is enabled, ask for user permission
	if tm.confirmationMode {
		// Extract command from arguments for display
		var displayCmd string
		if name == "execute_shell_command" {
			var params struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal([]byte(arguments), &params); err == nil {
				displayCmd = params.Command
			} else {
				displayCmd = arguments
			}
		} else {
			displayCmd = fmt.Sprintf("%s with args: %s", name, arguments)
		}

		if !UserConfirmation(displayCmd) {
			return "", fmt.Errorf("tool execution cancelled by user")
		}
	}

	return tool.Handler(arguments)
}

// UserConfirmation prompts the user for confirmation before executing a command
func UserConfirmation(command string) bool {
	fmt.Printf("\n🔍 About to execute command: %s\n", command)
	fmt.Print("Do you want to proceed? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
