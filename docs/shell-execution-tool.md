# k8x Shell Execution Tool

> **Note:** The preferred way to invoke k8x is `k8x -c "<goal>"`. You can also use `k8x run "<goal>"` or `k8x command "<goal>"` as alternatives.

## Overview

k8x now includes a **shell execution tool** that allows the LLM agent to execute safe, read-only shell commands automatically. This feature bridges the gap between LLM planning and actual command execution, making k8x a truly autonomous Kubernetes workflow assistant.

## Features

### 🔧 Shell Execution Tool

- **Function Name**: `execute_shell_command`
- **Purpose**: Execute safe, read-only shell commands primarily for kubectl operations
- **Safety First**: Built-in command filtering and validation

### 🛡️ Security Features

#### Allowed Commands

The tool allows execution of these safe commands:

**Core Commands:**

- `kubectl` (read-only operations only)
- `helm` (read-only operations only)
- `kustomize` (read-only operations only)
- `docker` (read-only operations only)
- `git` (read-only operations only)
- `echo`, `cat`, `ls`, `pwd`, `whoami`, `date`, `uname`, `which`
- `curl`, `ping`, `nslookup`, `dig`, `wget`, `telnet`, `nc` (for connectivity checks)

**Text Processing:**

- `head`, `tail`, `grep`, `awk`, `sed`, `sort`, `uniq`, `wc`, `find`

**System Information:**

- `ps`, `netstat`, `ss`, `lsof`, `df`, `du`, `free`, `uptime`, `id`, `env`, `printenv`
- `hostname`, `mount`, `lsblk`, `ip`, `ifconfig`, `route`, `arp`, `traceroute`

**Data Processing:**

- `jq`, `yq`, `base64`, `xxd`, `file`, `stat`

**Shell Utilities:**

- `history`, `alias`

#### Kubectl Safety Checks

Additional safety for kubectl commands - these operations are blocked:

- Write operations: `create`, `apply`, `delete`, `patch`, `replace`, `edit`
- Scaling: `scale`
- Labeling: `annotate`, `label`
- Service exposure: `expose`
- Configuration: `set`
- Rollouts: `rollout`
- Node management: `drain`, `cordon`, `uncordon`, `taint`
- Interactive operations: `exec`, `port-forward`, `proxy`, `attach`, `cp`

#### Helm Safety Checks

Additional safety for helm commands - these operations are blocked:

- Installation/Deployment: `install`, `upgrade`, `uninstall`, `delete`, `create`, `rollback`
- Repository management: `repo` (add, remove, update)
- Plugin management: `plugin`
- Registry operations: `push`, `pull`, `registry`

Only read-only operations like `list`, `status`, `get`, `history`, `version` are allowed.

#### Docker Safety Checks

Additional safety for docker commands - only these read-only operations are allowed:

- `docker version` - Show Docker version information
- `docker info` - Display system-wide information
- `docker ps` - List running containers
- `docker images` - List images

All other Docker operations including `run`, `build`, `push`, `pull`, etc. are blocked.

#### Git Safety Checks

Additional safety for git commands - only these read-only operations are allowed:

- `git version` - Show Git version
- `git status` - Show working tree status
- `git log` - Show commit logs
- `git show` - Show object information
- `git diff` - Show changes between commits
- `git blame` - Show what revision and author last modified each line

All other Git operations including `add`, `commit`, `push`, `pull`, `merge`, etc. are blocked.

#### Kustomize Safety Checks

Additional safety for kustomize commands - only these read-only operations are allowed:

- `kustomize version` - Show Kustomize version
- `kustomize build` - Build and output the configured resources (read-only)

All other Kustomize operations including `create`, `edit`, `fix` are blocked.

#### Suggestions for additional security

##### Sandbox K8X on the Host

- **Dedicated Unix User[Preferred]**: Run K8X under an unprivileged user.

  ```bash
  sudo adduser --system --home /opt/k8x --group k8xsvc
  sudo install -d -o k8xsvc -g k8xsvc /opt/k8x
  sudo -u k8xsvc k8x ...
  ```

- **Sudoers Whitelist**: Restrict privilege elevation to only the K8X binary.

  ```bash
  # /etc/sudoers.d/k8x
  k8x ALL=(root) NOPASSWD: /usr/bin/k8x
  ```

- **Container Sandboxing**: Use Firejail/Bubblewrap or rootless containers:
  - Firejail example:

    ```bash
    firejail --private --net=none k8x ...
    ```

  - Rootless container example:

    ```bash
    docker run --rm \
      --user 10000:10000 --read-only \
      --cap-drop ALL \
      --security-opt seccomp=default \
      -v $HOME/.kube:/kube:ro \
      ghcr.io/your-org/k8x:latest
    ```

- **macOS Sandboxing**: Isolate execution using sandbox-exec.

  ```bash
  sandbox-exec -f k8x.sb k8x ...
  ```

##### Use a Least-Privilege Kubernetes Identity (Optional)

Grant only necessary permissions with a namespace-scoped Role:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8x-operator
rules:
# Core resources
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]
# Workload resources
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch"]
```

Then, bind the Role to a dedicated ServiceAccount:

```yaml
# ServiceAccount lives in whatever namespace you prefer (e.g., k8x-tools)
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8x-sa
  namespace: k8x-tools
---
# ClusterRoleBinding grants the ClusterRole to that SA
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8x-binding
subjects:
- kind: ServiceAccount
  name: k8x-sa
  namespace: k8x-tools
roleRef:
  kind: ClusterRole
  name: k8x-operator
  apiGroup: rbac.authorization.k8s.io
```

Next, generate a kubeconfig referencing this ServiceAccount (e.g., export KUBECONFIG=~/.kube/k8x.conf). Pass this kubeconfig file and context as configuration parameters to k8x.

#### Command Timeout

- All commands have a 30-second timeout to prevent hanging

## How It Works

### 1. LLM Integration

- **OpenAI**: Full tool support using the latest OpenAI Go SDK
- **Anthropic**: Simplified tool support (falls back to regular chat)
- **Tool Definition**: JSON schema defines the tool interface

### 2. Execution Flow

```text
User Goal
  ↓
LLM Planning
  ↓
Tool Call (`execute_shell_command`)
  ↓
Shell Execution (with safety checks)
  ↓
Result (output/error)
  ↓
LLM Next Step (repeat or finish)
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
