package llm

import (
	"k8x/internal/config"
	"strings"
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

		expected := "Hello from k8x shell execution!"
		if !strings.Contains(result, expected) {
			t.Errorf("Echo result does not contain expected text. Got: %q", result)
		}
	})

	t.Run("kubectl version", func(t *testing.T) {
		result, err := toolManager.ExecuteTool("execute_shell_command", `{"command": "kubectl version --client"}`)
		if err != nil {
			t.Logf("kubectl test skipped (kubectl not available): %v", err)
		} else {
			// Just check that we got some output, don't log the full result
			if result == "" {
				t.Error("kubectl version returned empty result")
			}
			if !strings.Contains(result, "Client Version") {
				t.Error("kubectl version output doesn't contain expected 'Client Version' text")
			}
		}
	})

	t.Run("unsafe command should be blocked", func(t *testing.T) {
		_, err := toolManager.ExecuteTool("execute_shell_command", `{"command": "rm -rf /"}`)
		if err == nil {
			t.Error("Unsafe command was not blocked!")
		} else {
			// Just verify it was blocked, don't log the full error with all allowed commands
			if !strings.Contains(err.Error(), "not allowed for security reasons") {
				t.Errorf("Unexpected error message: %v", err)
			}
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

func TestToolManagerConfirmationMode(t *testing.T) {
	toolManager := NewToolManager(".")

	t.Run("confirmation mode disabled by default", func(t *testing.T) {
		if toolManager.confirmationMode {
			t.Error("ToolManager should have confirmation mode disabled by default")
		}
	})

	t.Run("set confirmation mode", func(t *testing.T) {
		toolManager.SetConfirmationMode(true)
		if !toolManager.confirmationMode {
			t.Error("SetConfirmationMode(true) should enable confirmation mode")
		}

		toolManager.SetConfirmationMode(false)
		if toolManager.confirmationMode {
			t.Error("SetConfirmationMode(false) should disable confirmation mode")
		}
	})

	t.Run("confirmation mode execution without user input", func(t *testing.T) {
		// Since we can't simulate user input in unit tests easily,
		// we'll just test that the tool manager can handle the confirmation logic
		// The actual user interaction would need to be tested manually or with integration tests
		toolManager.SetConfirmationMode(true)
		
		// In confirmation mode, tools should fail if no user input is provided
		// This is expected behavior since UserConfirmation would be waiting for input
		// We'll just check that the confirmation mode is properly set
		if !toolManager.confirmationMode {
			t.Error("Confirmation mode should be enabled")
		}
	})
}
