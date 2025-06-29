package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Credentials represents the credentials file structure
type Credentials struct {
	OpenAI struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"openai"`
	Anthropic struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"anthropic"`
}

// LoadCredentials loads credentials from the credentials file
func LoadCredentials() (*Credentials, error) {
	credentialsPath, err := GetCredentialsPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials path: %w", err)
	}

	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		// Return empty credentials if file doesn't exist
		return &Credentials{}, nil
	}

	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	return &creds, nil
}

// SaveCredentials saves credentials to the credentials file
func SaveCredentials(creds *Credentials) error {
	credentialsPath, err := GetCredentialsPath()
	if err != nil {
		return fmt.Errorf("failed to get credentials path: %w", err)
	}

	// Ensure the directory exists
	credentialsDir := filepath.Dir(credentialsPath)
	if err := os.MkdirAll(credentialsDir, 0700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	data, err := yaml.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(credentialsPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// CreateDefaultCredentialsFile creates a default credentials file with examples
func CreateDefaultCredentialsFile() error {
	credentialsPath, err := GetCredentialsPath()
	if err != nil {
		return fmt.Errorf("failed to get credentials path: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(credentialsPath); err == nil {
		return nil // File already exists, don't overwrite
	}

	defaultCredentials := `# k8x Credentials File
# Add your LLM provider API keys here

openai:
  api_key: "your-openai-api-key"

anthropic:
  api_key: "your-anthropic-api-key"

# Uncomment and configure the providers you want to use
# Make sure to keep this file secure and don't commit it to version control
`

	// Ensure the directory exists
	credentialsDir := filepath.Dir(credentialsPath)
	if err := os.MkdirAll(credentialsDir, 0700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	if err := os.WriteFile(credentialsPath, []byte(defaultCredentials), 0600); err != nil {
		return fmt.Errorf("failed to write default credentials file: %w", err)
	}

	return nil
}
