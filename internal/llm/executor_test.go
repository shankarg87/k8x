package llm

import (
	"k8x/internal/config"
	"os"
	"strings"
	"testing"
)

// executeCommandForTest is a helper function that creates a command but doesn't execute it
// Instead, it returns the command that would have been executed and any environment variables
func executeCommandForTest(command string, k8sConfig *config.KubernetesConfig) (string, []string) {
	parts := strings.Fields(command)
	baseCmd := parts[0]

	// Apply Kubernetes configuration if available for kubectl
	if baseCmd == "kubectl" && k8sConfig != nil {
		// Add context flag if specified
		if k8sConfig.Context != "" && !strings.Contains(command, "--context") {
			command = command + " --context=" + k8sConfig.Context
		}

		// Add namespace flag if specified and not already in command
		if k8sConfig.Namespace != "" && !strings.Contains(command, "--namespace") && !strings.Contains(command, "-n ") {
			command = command + " --namespace=" + k8sConfig.Namespace
		}
	}

	// Create environment slice
	env := os.Environ()
	if k8sConfig != nil && k8sConfig.KubeConfigPath != "" && baseCmd == "kubectl" {
		env = append(env, "KUBECONFIG="+k8sConfig.KubeConfigPath)
	}

	return command, env
}

func TestKubernetesConfigCommandModification(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		k8sConfig      *config.KubernetesConfig
		expectedCmd    string
		expectedEnvVar string
	}{
		{
			name:        "no config",
			command:     "kubectl get pods",
			k8sConfig:   nil,
			expectedCmd: "kubectl get pods",
		},
		{
			name:        "with context",
			command:     "kubectl get pods",
			k8sConfig:   &config.KubernetesConfig{Context: "test-context"},
			expectedCmd: "kubectl get pods --context=test-context",
		},
		{
			name:        "with namespace",
			command:     "kubectl get pods",
			k8sConfig:   &config.KubernetesConfig{Namespace: "test-namespace"},
			expectedCmd: "kubectl get pods --namespace=test-namespace",
		},
		{
			name:        "with context and namespace",
			command:     "kubectl get pods",
			k8sConfig:   &config.KubernetesConfig{Context: "test-context", Namespace: "test-namespace"},
			expectedCmd: "kubectl get pods --context=test-context --namespace=test-namespace",
		},
		{
			name:           "with kubeconfig",
			command:        "kubectl get pods",
			k8sConfig:      &config.KubernetesConfig{KubeConfigPath: "/path/to/kubeconfig"},
			expectedCmd:    "kubectl get pods",
			expectedEnvVar: "KUBECONFIG=/path/to/kubeconfig",
		},
		{
			name:        "non-kubectl command",
			command:     "echo hello",
			k8sConfig:   &config.KubernetesConfig{Context: "test-context", Namespace: "test-namespace"},
			expectedCmd: "echo hello",
		},
		{
			name:        "command with existing namespace",
			command:     "kubectl get pods -n other-namespace",
			k8sConfig:   &config.KubernetesConfig{Namespace: "test-namespace"},
			expectedCmd: "kubectl get pods -n other-namespace",
		},
		{
			name:        "command with existing context",
			command:     "kubectl get pods --context=other-context",
			k8sConfig:   &config.KubernetesConfig{Context: "test-context"},
			expectedCmd: "kubectl get pods --context=other-context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, env := executeCommandForTest(tt.command, tt.k8sConfig)

			if cmd != tt.expectedCmd {
				t.Errorf("Command = %q, want %q", cmd, tt.expectedCmd)
			}

			if tt.expectedEnvVar != "" {
				found := false
				for _, e := range env {
					if e == tt.expectedEnvVar {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Environment variable %q not found in %v", tt.expectedEnvVar, env)
				}
			}
		})
	}
}

func TestToolManagerKubernetesConfig(t *testing.T) {
	// This is an integration test that verifies the ToolManager correctly passes the k8s config to executor

	// Create a ToolManager
	tm := NewToolManager(".")

	// Set Kubernetes config
	tm.SetKubernetesConfig(&config.KubernetesConfig{
		Context:   "test-context",
		Namespace: "test-namespace",
	})

	// Verify that the executor has the config set (indirect test since we can't access private fields)
	// We'll need to inspect the actual command execution in a real environment
	if tm.executor.k8sConfig == nil {
		t.Error("Kubernetes config not passed to executor")
	}

	if tm.executor.k8sConfig.Context != "test-context" {
		t.Errorf("Context = %q, want %q", tm.executor.k8sConfig.Context, "test-context")
	}

	if tm.executor.k8sConfig.Namespace != "test-namespace" {
		t.Errorf("Namespace = %q, want %q", tm.executor.k8sConfig.Namespace, "test-namespace")
	}
}
