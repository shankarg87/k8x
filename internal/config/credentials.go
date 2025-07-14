package config

import (
	"fmt"
	"os"
	"path/filepath"

	"k8x/internal/schemas"

	"gopkg.in/yaml.v3"
)

// LoadCredentials loads credentials from the credentials file
func LoadCredentials() (*schemas.Credentials, error) {
	credentialsPath, err := GetCredentialsPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials path: %w", err)
	}

	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var creds schemas.Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file: %w", err)
	}

	return &creds, nil
}

// SaveCredentials saves credentials to the credentials file
func SaveCredentials(creds *schemas.Credentials) error {
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

	// Create default credentials structure
	defaultCreds := &schemas.Credentials{
		SelectedProvider: "openai",
	}
	defaultCreds.OpenAI.APIKey = "your-openai-api-key-here"
	defaultCreds.Anthropic.APIKey = "your-anthropic-api-key-here"
	defaultCreds.Google.APIKey = "your-gemini-api-key-here"
	defaultCreds.Google.ApplicationCredentials = "/path/to/service-account.json"

	// Ensure the directory exists
	credentialsDir := filepath.Dir(credentialsPath)
	if err := os.MkdirAll(credentialsDir, 0700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	// Save the default credentials
	return SaveCredentials(defaultCreds)
}
