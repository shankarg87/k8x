package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"k8x/internal/config"
)

func TestRunCommand_LoadsKubernetesConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Errorf("Failed to restore HOME environment variable: %v", err)
		}
	}()
	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}

	// Create necessary directory structure
	configDir := filepath.Join(tempDir, config.DefaultConfigDir)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create a minimal config file with Kubernetes settings
	configPath := filepath.Join(configDir, config.DefaultConfigFileName)
	configContent := `
llm:
  default_provider: "openai"
  providers:
    openai:
      model: "gpt-4"
kubernetes:
  context: "test-context"
  namespace: "test-namespace"
  kubeconfig_path: "/path/to/kubeconfig"
settings:
  history_enabled: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create credentials file to avoid errors
	credsPath := filepath.Join(configDir, config.CredentialsFile)
	credsContent := `
selectedProvider: openai
openai:
  api_key: "test-api-key"
`
	if err := os.WriteFile(credsPath, []byte(credsContent), 0644); err != nil {
		t.Fatalf("Failed to write test credentials file: %v", err)
	}

	// Create history directory
	historyDir := filepath.Join(configDir, config.DefaultHistoryDir)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		t.Fatalf("Failed to create history directory: %v", err)
	}

	// Now test that LoadConfig works and loads the correct Kubernetes settings
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Verify kubernetes settings
	if cfg.Kubernetes.Context != "test-context" {
		t.Errorf("Kubernetes.Context = %q, want %q", cfg.Kubernetes.Context, "test-context")
	}

	if cfg.Kubernetes.Namespace != "test-namespace" {
		t.Errorf("Kubernetes.Namespace = %q, want %q", cfg.Kubernetes.Namespace, "test-namespace")
	}

	if cfg.Kubernetes.KubeConfigPath != "/path/to/kubeconfig" {
		t.Errorf("Kubernetes.KubeConfigPath = %q, want %q", cfg.Kubernetes.KubeConfigPath, "/path/to/kubeconfig")
	}
}

func TestRunCommandFlags(t *testing.T) {
	// Test that the run command has the confirm flag properly configured
	cmd := runCmd
	
	// Check that the confirm flag exists
	confirmFlag := cmd.Flags().Lookup("confirm")
	if confirmFlag == nil {
		t.Error("Run command should have a --confirm flag")
		return
	}
	
	// Check flag properties
	if confirmFlag.Shorthand != "c" {
		t.Errorf("Confirm flag shorthand = %q, want %q", confirmFlag.Shorthand, "c")
	}
	
	if confirmFlag.DefValue != "false" {
		t.Errorf("Confirm flag default value = %q, want %q", confirmFlag.DefValue, "false")
	}
	
	if confirmFlag.Usage != "Ask for confirmation before executing each tool" {
		t.Errorf("Confirm flag usage = %q, want %q", confirmFlag.Usage, "Ask for confirmation before executing each tool")
	}
}
