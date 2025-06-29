package config

import (
	"os"
	"path/filepath"
)

const (
	// DefaultConfigDir is the default directory for k8x configuration
	DefaultConfigDir = ".k8x"
	// DefaultHistoryDir is the subdirectory for command history
	DefaultHistoryDir = "history"
	// CredentialsFile is the file containing LLM provider credentials
	CredentialsFile = "credentials"
)

// Config represents the application configuration
type Config struct {
	// LLM configuration
	LLM LLMConfig `yaml:"llm"`
	// Kubernetes configuration
	Kubernetes KubernetesConfig `yaml:"kubernetes"`
	// General settings
	Settings GeneralSettings `yaml:"settings"`
}

// LLMConfig contains configuration for LLM providers
type LLMConfig struct {
	// DefaultProvider is the default LLM provider to use
	DefaultProvider string `yaml:"default_provider"`
	// Providers contains provider-specific configurations
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// ProviderConfig contains configuration for a specific LLM provider
type ProviderConfig struct {
	// APIKey is stored separately in credentials file
	BaseURL string            `yaml:"base_url,omitempty"`
	Model   string            `yaml:"model,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
}

// KubernetesConfig contains Kubernetes-specific settings
type KubernetesConfig struct {
	// Context is the default Kubernetes context to use
	Context string `yaml:"context,omitempty"`
	// Namespace is the default namespace
	Namespace string `yaml:"namespace,omitempty"`
}

// GeneralSettings contains general application settings
type GeneralSettings struct {
	// Verbose enables verbose output
	Verbose bool `yaml:"verbose"`
	// HistoryEnabled enables command history tracking
	HistoryEnabled bool `yaml:"history_enabled"`
	// UndoEnabled enables undo functionality
	UndoEnabled bool `yaml:"undo_enabled"`
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, DefaultConfigDir), nil
}

// GetHistoryDir returns the history directory path
func GetHistoryDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, DefaultHistoryDir), nil
}

// GetCredentialsPath returns the credentials file path
func GetCredentialsPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, CredentialsFile), nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	historyDir, err := GetHistoryDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(historyDir, 0755)
}
