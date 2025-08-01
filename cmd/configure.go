package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"k8x/internal/config"
	"k8x/internal/schemas"

	"github.com/spf13/cobra"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Initialize k8x workspace and configuration",
	Long: `Initialize k8x workspace by creating the default configuration directory,
example configuration files, and credentials template. This is the first command
you should run after installing k8x.

This command will:
- Create ~/.k8x/ directory structure
- Copy example configuration files
- Create credentials template
- Display setup instructions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure config directory exists
		if err := config.EnsureConfigDir(); err != nil {
			return fmt.Errorf("failed to initialize config directory: %w", err)
		}

		configDir, err := config.GetConfigDir()
		if err != nil {
			return err
		}

		// Create example config file if it doesn't exist
		configPath := filepath.Join(configDir, "config.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := copyExampleFile("examples/config.yaml", configPath); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
			fmt.Printf("✓ Created configuration file: %s\n", configPath)
		} else {
			fmt.Printf("✓ Configuration file exists: %s\n", configPath)
		}

		// Create example credentials file if it doesn't exist
		credentialsPath := filepath.Join(configDir, "credentials")
		if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
			if err := copyExampleFile("examples/credentials", credentialsPath); err != nil {
				return fmt.Errorf("failed to create credentials file: %w", err)
			}
			fmt.Printf("✓ Created credentials template: %s\n", credentialsPath)
		} else {
			fmt.Printf("✓ Credentials file exists: %s\n", credentialsPath)
		}

		// Create history directory
		historyDir, err := config.GetHistoryDir()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(historyDir, 0755); err != nil {
			return fmt.Errorf("failed to create history directory: %w", err)
		}
		fmt.Printf("✓ History directory ready: %s\n", historyDir)

		// Step 1: Ask user for LLM provider
		fmt.Println("\nSelect your preferred LLM provider:")
		fmt.Println("  1. OpenAI")
		fmt.Println("  2. Anthropic")
		fmt.Println("  3. Google (Gemini API)")
		fmt.Print("Enter choice [1-3]: ")
		var providerChoice int
		_, err = fmt.Scanln(&providerChoice)
		if err != nil || providerChoice < 1 || providerChoice > 3 {
			return fmt.Errorf("invalid provider selection")
		}

		var provider string
		var providerLabel string
		switch providerChoice {
		case 1:
			provider = "openai"
			providerLabel = "OpenAI"
		case 2:
			provider = "anthropic"
			providerLabel = "Anthropic"
		case 3:
			provider = "google"
			providerLabel = "Google (Gemini API)"
		}

		// Step 2: Ask for API key
		fmt.Printf("Enter your %s API key: ", providerLabel)
		var apiKey string
		_, err = fmt.Scanln(&apiKey)
		if err != nil || apiKey == "" {
			return fmt.Errorf("invalid API key")
		}

		// Load existing credentials or create new ones
		creds, err := config.LoadCredentials()
		if err != nil {
			// If file doesn't exist or is invalid, create new credentials
			creds = &schemas.Credentials{}
		}

		// Update credentials with the new provider and API key
		creds.SetProviderAPIKey(provider, apiKey)

		// Save the updated credentials
		if err := config.SaveCredentials(creds); err != nil {
			return fmt.Errorf("failed to update credentials: %w", err)
		}
		fmt.Printf("✓ Updated credentials for %s\n", provider)

		fmt.Println("\n🚀 k8x workspace initialized successfully!")
		fmt.Println("\nNext steps:")
		fmt.Println("Run 'k8x run \"<your kubernetes goal>\"' to start using k8x")

		fmt.Println("\nExample:")
		fmt.Println("  k8x run \"List all pods in the default namespace\"")

		return nil
	},
}

// copyExampleFile copies a file from the examples directory to the destination
func copyExampleFile(srcPath, destPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read example file %s: %w", srcPath, err)
	}

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
