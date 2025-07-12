package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	// DefaultConfigDir is the default directory for k8x configuration
	DefaultConfigDir = ".k8x"
	// DefaultHistoryDir is the subdirectory for command history
	DefaultHistoryDir = "history"
	// CredentialsFile is the file containing LLM provider credentials
	CredentialsFile = "credentials"
	// DefaultConfigFileName is the default configuration file name
	DefaultConfigFileName = "config.yaml"
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
	BaseURL       string            `yaml:"base_url,omitempty"`
	Model         string            `yaml:"model,omitempty"`
	ContextLength int               `yaml:"context_length,omitempty"` // Context window size in tokens
	Options       map[string]string `yaml:"options,omitempty"`
}

// KubernetesConfig contains Kubernetes-specific settings
type KubernetesConfig struct {
	// Context is the default Kubernetes context to use
	Context string `yaml:"context,omitempty"`
	// Namespace is the default namespace
	Namespace string `yaml:"namespace,omitempty"`
	// KubeConfigPath is the path to the kubeconfig file
	KubeConfigPath string `yaml:"kubeconfig_path,omitempty"`
}

// GeneralSettings contains general application settings
type GeneralSettings struct {
	// Verbose enables verbose output
	Verbose bool `yaml:"verbose"`
	// HistoryEnabled enables command history tracking
	HistoryEnabled bool `yaml:"history_enabled"`
	// UndoEnabled enables undo functionality
	UndoEnabled bool `yaml:"undo_enabled"`
	// Summarizer contains summarization configuration
	Summarizer SummarizerConfig `yaml:"summarizer"`
}

// SummarizerConfig contains configuration for conversation summarization
type SummarizerConfig struct {
	// SummarizeAtPercent is the percentage of context length at which to trigger summarization (default 70%)
	SummarizeAtPercent int `yaml:"summarize_at_percent,omitempty"`
	// KeepConversations is the number of recent conversation exchanges to keep intact (default 1)
	KeepConversations int `yaml:"keep_conversations,omitempty"`
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

// GetConfigPath returns the configuration file path
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, DefaultConfigFileName), nil
}

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &Config{
			LLM: LLMConfig{
				DefaultProvider: "openai",
				Providers: map[string]ProviderConfig{
					"openai": {
						Model: "gpt-4",
					},
				},
			},
			Kubernetes: KubernetesConfig{},
			Settings: GeneralSettings{
				HistoryEnabled: true,
				Summarizer: SummarizerConfig{
					SummarizeAtPercent: 70,
					KeepConversations:  1,
				},
			},
		}, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetContextLength returns the context length for a provider, with fallback to model defaults
func (c *Config) GetContextLength(providerName string) int {
	if provider, exists := c.LLM.Providers[providerName]; exists && provider.ContextLength > 0 {
		return provider.ContextLength
	}
	
	// Fallback to model-specific defaults
	if provider, exists := c.LLM.Providers[providerName]; exists {
		return getDefaultContextLength(providerName, provider.Model)
	}
	
	// Ultimate fallback
	return 4096
}

// getDefaultContextLength returns default context lengths for known models
func getDefaultContextLength(providerName, model string) int {
	defaults := map[string]map[string]int{
		"openai": {
			"gpt-4":           8192,
			"gpt-4-turbo":     128000,
			"gpt-4o":          128000,
			"gpt-4o-mini":     128000,
			"o1":              200000,
			"o1-mini":         128000,
			"o3-mini":         128000,
			"gpt-3.5-turbo":   16385,
		},
		"anthropic": {
			"claude-3-haiku-20240307":   200000,
			"claude-3-sonnet-20240229":  200000,
			"claude-3-opus-20240229":    200000,
			"claude-3-5-sonnet-20241022": 200000,
			"claude-3-5-haiku-20241022":  200000,
		},
	}
	
	if providerDefaults, exists := defaults[providerName]; exists {
		if contextLength, exists := providerDefaults[model]; exists {
			return contextLength
		}
	}
	
	// Provider-specific fallbacks
	switch providerName {
	case "openai":
		return 8192
	case "anthropic":
		return 200000
	default:
		return 4096
	}
}
