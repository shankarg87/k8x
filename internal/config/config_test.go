package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigDir(t *testing.T) {
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() failed: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}

	expected := filepath.Join(home, DefaultConfigDir)
	if configDir != expected {
		t.Errorf("GetConfigDir() = %v, want %v", configDir, expected)
	}
}

func TestGetHistoryDir(t *testing.T) {
	historyDir, err := GetHistoryDir()
	if err != nil {
		t.Fatalf("GetHistoryDir() failed: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}

	expected := filepath.Join(home, DefaultConfigDir, DefaultHistoryDir)
	if historyDir != expected {
		t.Errorf("GetHistoryDir() = %v, want %v", historyDir, expected)
	}
}

func TestGetCredentialsPath(t *testing.T) {
	credentialsPath, err := GetCredentialsPath()
	if err != nil {
		t.Fatalf("GetCredentialsPath() failed: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}

	expected := filepath.Join(home, DefaultConfigDir, CredentialsFile)
	if credentialsPath != expected {
		t.Errorf("GetCredentialsPath() = %v, want %v", credentialsPath, expected)
	}
}

func TestEnsureConfigDir(t *testing.T) {
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

	err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() failed: %v", err)
	}

	// Check if config directory was created
	configDir := filepath.Join(tempDir, DefaultConfigDir)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created: %v", configDir)
	}

	// Check if history directory was created
	historyDir := filepath.Join(tempDir, DefaultConfigDir, DefaultHistoryDir)
	if _, err := os.Stat(historyDir); os.IsNotExist(err) {
		t.Errorf("History directory was not created: %v", historyDir)
	}
}

func TestLoadConfig(t *testing.T) {
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

	// Test case 1: Config file doesn't exist, should return default config
	t.Run("default config when file doesn't exist", func(t *testing.T) {
		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		if cfg.LLM.DefaultProvider != "openai" {
			t.Errorf("Default provider = %v, want openai", cfg.LLM.DefaultProvider)
		}

		if len(cfg.LLM.Providers) != 1 || cfg.LLM.Providers["openai"].Model != "gpt-4" {
			t.Errorf("Default providers not correctly initialized")
		}

		if !cfg.Settings.HistoryEnabled {
			t.Errorf("HistoryEnabled = %v, want true", cfg.Settings.HistoryEnabled)
		}
	})

	// Test case 2: Create a config file with Kubernetes settings
	t.Run("load config with kubernetes settings", func(t *testing.T) {
		// Create config directory
		configDir := filepath.Join(tempDir, DefaultConfigDir)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create a test config file
		configPath := filepath.Join(configDir, DefaultConfigFileName)
		configContent := `
llm:
  default_provider: "anthropic"
  providers:
    anthropic:
      model: "claude-3-5-sonnet"
kubernetes:
  context: "test-context"
  namespace: "test-namespace"
  kubeconfig_path: "/path/to/kubeconfig"
settings:
  verbose: true
  history_enabled: false
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Load the config
		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() failed: %v", err)
		}

		// Verify the loaded config
		if cfg.LLM.DefaultProvider != "anthropic" {
			t.Errorf("DefaultProvider = %v, want anthropic", cfg.LLM.DefaultProvider)
		}

		if cfg.Kubernetes.Context != "test-context" {
			t.Errorf("Kubernetes.Context = %v, want test-context", cfg.Kubernetes.Context)
		}

		if cfg.Kubernetes.Namespace != "test-namespace" {
			t.Errorf("Kubernetes.Namespace = %v, want test-namespace", cfg.Kubernetes.Namespace)
		}

		if cfg.Kubernetes.KubeConfigPath != "/path/to/kubeconfig" {
			t.Errorf("Kubernetes.KubeConfigPath = %v, want /path/to/kubeconfig", cfg.Kubernetes.KubeConfigPath)
		}

		if !cfg.Settings.Verbose {
			t.Errorf("Settings.Verbose = %v, want true", cfg.Settings.Verbose)
		}

		if cfg.Settings.HistoryEnabled {
			t.Errorf("Settings.HistoryEnabled = %v, want false", cfg.Settings.HistoryEnabled)
		}
	})
}
