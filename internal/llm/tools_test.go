package llm

import (
	"k8x/internal/config"
	"testing"
)

func TestToolManager(t *testing.T) {
	toolManager := NewToolManager(".")

	t.Run("basic echo command", func(t *testing.T) {
		result, err := toolManager.ExecuteTool("execute_shell_command", `{"command": "echo 'Hello from k8x shell execution!'"}`)
		if err != nil {
			t.Errorf("Echo command failed: %v", err)
			return
		}

		if result == "" {
			t.Error("Echo command returned empty result")
		}
		t.Logf("Echo test result: %s", result)
	})

	t.Run("kubectl version", func(t *testing.T) {
		result, err := toolManager.ExecuteTool("execute_shell_command", `{"command": "kubectl version --client"}`)
		if err != nil {
			t.Logf("kubectl test error (expected if kubectl not available): %v", err)
		} else {
			t.Logf("kubectl version result: %s", result)
		}
	})

	t.Run("unsafe command should be blocked", func(t *testing.T) {
		_, err := toolManager.ExecuteTool("execute_shell_command", `{"command": "rm -rf /"}`)
		if err == nil {
			t.Error("Unsafe command was not blocked!")
		} else {
			t.Logf("Unsafe command correctly blocked: %v", err)
		}
	})
}

func TestShellExecutorKubernetesConfigIntegration(t *testing.T) {
	// This is just a simple test to check that SetKubernetesConfig works
	executor := NewShellExecutor(".")

	// Set kubernetes config
	k8sConfig := &config.KubernetesConfig{
		Context:        "test-context",
		Namespace:      "test-namespace",
		KubeConfigPath: "/path/to/kubeconfig",
	}
	executor.SetKubernetesConfig(k8sConfig)

	// Verify that the config was set correctly
	if executor.k8sConfig == nil {
		t.Error("Kubernetes config not set")
	}

	if executor.k8sConfig.Context != "test-context" {
		t.Errorf("Context = %q, want %q", executor.k8sConfig.Context, "test-context")
	}

	if executor.k8sConfig.Namespace != "test-namespace" {
		t.Errorf("Namespace = %q, want %q", executor.k8sConfig.Namespace, "test-namespace")
	}

	if executor.k8sConfig.KubeConfigPath != "/path/to/kubeconfig" {
		t.Errorf("KubeConfigPath = %q, want %q", executor.k8sConfig.KubeConfigPath, "/path/to/kubeconfig")
	}
}
