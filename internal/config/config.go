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
	// MCP configuration
	MCP MCPConfig `yaml:"mcp"`
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

// MCPConfig contains configuration for MCP servers
type MCPConfig struct {
	// Servers contains MCP server configurations
	Servers map[string]MCPServerConfig `yaml:"servers"`
	// Enabled controls whether MCP integration is enabled
	Enabled bool `yaml:"enabled"`
}

// MCPServerConfig contains configuration for a specific MCP server
type MCPServerConfig struct {
	// Transport specifies the transport type (stdio, sse, http, oauth-sse, oauth-http)
	Transport string `yaml:"transport"`
	// Enabled controls whether this server is enabled
	Enabled bool `yaml:"enabled"`
	// Description is a human-readable description of the server
	Description string `yaml:"description,omitempty"`

	// Stdio transport configuration
	Command string            `yaml:"command,omitempty"`
	Args    []string          `yaml:"args,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`

	// HTTP/SSE transport configuration
	BaseURL string `yaml:"base_url,omitempty"`

	// OAuth configuration (for oauth-sse and oauth-http transports)
	OAuth *OAuthConfig `yaml:"oauth,omitempty"`

	// Transport-specific options
	Options map[string]interface{} `yaml:"options,omitempty"`
}

// OAuthConfig contains OAuth authentication configuration
type OAuthConfig struct {
	ClientID              string   `yaml:"client_id"`
	ClientSecret          string   `yaml:"client_secret,omitempty"`
	RedirectURI           string   `yaml:"redirect_uri,omitempty"`
	Scopes                []string `yaml:"scopes,omitempty"`
	AuthServerMetadataURL string   `yaml:"auth_server_metadata_url,omitempty"`
	PKCEEnabled           bool     `yaml:"pkce_enabled,omitempty"`
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
			MCP: MCPConfig{
				Enabled: false,
				Servers: make(map[string]MCPServerConfig),
			},
			Kubernetes: KubernetesConfig{},
			Settings: GeneralSettings{
				HistoryEnabled: true,
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
