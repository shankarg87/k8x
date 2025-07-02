# k8x Shell Execution Tool

## Overview

k8x now includes a **shell execution tool** that allows the LLM agent to execute safe, read-only shell commands automatically. This feature bridges the gap between LLM planning and actual command execution, making k8x a truly autonomous Kubernetes workflow assistant.

## Features

### 🔧 Shell Execution Tool
- **Function Name**: `execute_shell_command`
- **Purpose**: Execute safe, read-only shell commands primarily for kubectl operations
- **Safety First**: Built-in command filtering and validation

### 🛡️ Security Features

#### Allowed Commands
The tool only allows execution of these safe commands:
- `kubectl` (read-only operations only)
- `echo`, `cat`, `ls`, `pwd`, `whoami`, `date`, `uname`, `which`
- `curl`, `ping`, `nslookup`, `dig` (for connectivity checks)

#### kubectl Safety Checks
Additional safety for kubectl commands - these operations are blocked:
- Write operations: `create`, `apply`, `delete`, `patch`, `replace`, `edit`
- Scaling: `scale`
- Labeling: `annotate`, `label`
- Service exposure: `expose`
- Configuration: `set`
- Rollouts: `rollout`
- Node management: `drain`, `cordon`, `uncordon`, `taint`
- Interactive operations: `exec`, `port-forward`, `proxy`, `attach`, `cp`

#### Command Timeout
- All commands have a 30-second timeout to prevent hanging

## How It Works

### 1. LLM Integration
- **OpenAI**: Full tool support using the latest OpenAI Go SDK
- **Anthropic**: Simplified tool support (falls back to regular chat)
- **Tool Definition**: JSON schema defines the tool interface

### 2. Execution Flow
```
User Goal → LLM Planning → Tool Call → Shell Execution → Result → Next Step
```

### 3. Example Workflow
```bash
k8x run "Check if any pods are failing in the default namespace"
```

The LLM will:
1. Plan the approach
2. Call `execute_shell_command` with `kubectl get pods`
3. Analyze the results
4. Potentially call more commands like `kubectl describe pod <failing-pod>`
5. Provide a comprehensive diagnosis

## Technical Implementation

### Tool Manager (`internal/llm/tools.go`)
- `NewToolManager(workDir)`: Creates a tool manager with shell executor
- `ExecuteTool(name, arguments)`: Executes a tool by name
- `GetTools()`: Returns available tool definitions

### Shell Executor
- `NewShellExecutor(workDir)`: Creates executor with safety restrictions
- `Execute(command)`: Runs command with security checks
- Command parsing and validation
- Output capture and error handling

### Provider Integration
- **OpenAI**: `ChatWithTools()` method using native tool support
- **Anthropic**: `ChatWithTools()` method with fallback to regular chat
- **Unified Provider**: Automatic provider detection and tool routing

## Usage Examples

### Basic Diagnostic
```bash
k8x run "List all pods and show their status"
```

### Advanced Troubleshooting
```bash
k8x run "Find why my nginx service is not receiving traffic"
```

### Resource Investigation
```bash
k8x run "Check resource usage and identify any resource-constrained pods"
```

## Safety Guarantees

1. **Read-Only Mode**: Only safe, non-destructive commands are allowed
2. **Command Filtering**: Whitelist-based approach for maximum security
3. **kubectl Protection**: Additional layer specifically for kubectl write operations
4. **Timeout Protection**: All commands timeout after 30 seconds
5. **Error Handling**: Graceful failure with informative error messages

## Future Enhancements

- **User Confirmation**: Optional prompt before executing commands
- **Command History**: Track and replay command sequences
- **Custom Tool Support**: Allow users to define additional safe tools
- **Advanced kubectl Safety**: More granular permission controls
- **Multi-cluster Support**: Execute commands across different clusters

## Integration with k8x Workflow

The shell execution tool integrates seamlessly with k8x's existing workflow:
- **History Tracking**: All commands and outputs are saved in `.k8x` files
- **Step-by-Step Execution**: Each tool call becomes a step in the session
- **LLM Context**: Tool results are fed back to the LLM for continued planning
- **Goal Completion**: LLM can determine when the goal is achieved

This makes k8x a powerful, autonomous Kubernetes assistant that can actually perform diagnostic and investigative tasks without manual intervention.
