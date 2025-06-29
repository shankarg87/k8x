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
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

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
