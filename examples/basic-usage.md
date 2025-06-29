# Basic Usage Examples

## Getting Started

First, initialize k8x configuration:

```bash
k8x config init
```

Set up your LLM provider credentials:

```bash
# Edit the credentials file
vim ~/.k8x/credentials
```

## Common Use Cases

### 1. Asking Questions About Your Cluster

```bash
# Get an overview of your cluster
k8x ask "What's the status of my cluster?"

# Find problematic pods
k8x ask "Which pods are not running?"

# Get resource usage information
k8x ask "Show me pods using the most CPU"
```

### 2. Troubleshooting

```bash
# Diagnose a failing deployment
k8x diagnose deployment my-app

# Check why a pod is pending
k8x diagnose pod my-pod-123

# Get logs with context
k8x ask "Show me logs for the failed pods in namespace production"
```

### 3. Resource Management

```bash
# Scale a deployment with natural language
k8x ask "Scale my-app deployment to 5 replicas"

# Update resource limits
k8x ask "Increase memory limit for my-app to 2Gi"

# Create resources from description
k8x ask "Create a nginx deployment with 3 replicas"
```

### 4. Monitoring and Alerts

```bash
# Check for issues
k8x ask "Are there any unhealthy nodes?"

# Monitor specific resources
k8x ask "Watch the status of all deployments"

# Get security insights
k8x ask "Show me pods running as root"
```

## Interactive Mode

Start an interactive session for continuous assistance:

```bash
k8x interactive
```

In interactive mode, you can have a conversation about your cluster:

```
> What pods are running in the kube-system namespace?
> Can you restart the coredns deployment?
> Show me the events for the last 10 minutes
```

## Configuration Examples

### Setting Default Context and Namespace

```bash
k8x config set kubernetes.context my-cluster
k8x config set kubernetes.namespace production
```

### Configuring LLM Provider

```bash
k8x config set llm.default_provider openai
k8x config set llm.providers.openai.model gpt-4
```
