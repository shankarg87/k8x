package cmd

import (
	"fmt"

	"github.com/shankgan/k8x/internal/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage k8x configuration",
	Long: `Manage k8x configuration including LLM providers, 
Kubernetes settings, and general preferences.`,
}

// configInitCmd represents the config init command
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize k8x configuration",
	Long: `Initialize k8x configuration by creating the default
configuration directory and files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureConfigDir(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		configDir, err := config.GetConfigDir()
		if err != nil {
			return err
		}

		fmt.Printf("Configuration initialized in %s\n", configDir)
		fmt.Println("Edit ~/.k8x/config.yaml and ~/.k8x/credentials to configure your settings.")
		fmt.Println("Add your LLM provider API keys to ~/.k8x/credentials")
		return nil
	},
}

// configSetCmd represents the config set command
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  `Set a configuration value using dot notation (e.g., llm.default_provider)`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// TODO: Implement configuration setting logic
		fmt.Printf("Setting %s = %s\n", key, value)
		fmt.Println("Note: Configuration setting not yet implemented")
		return nil
	},
}

// configGetCmd represents the config get command
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value(s)",
	Long:  `Get a specific configuration value or display all configuration`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// TODO: Display all configuration
			fmt.Println("Configuration display not yet implemented")
		} else {
			key := args[0]
			// TODO: Get specific configuration value
			fmt.Printf("Getting value for %s\n", key)
			fmt.Println("Note: Configuration getting not yet implemented")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
}
