package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8x",
	Short: "Agentic kubectl - AI-powered Kubernetes operations",
	Long: `k8x is an AI-powered CLI tool that acts as an intelligent layer on top of kubectl.
It helps you manage Kubernetes resources through natural language commands and
provides automated assistance for common operations.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if -c flag was used
		goal, err := cmd.Flags().GetString("c")
		if err != nil {
			return fmt.Errorf("failed to get -c flag: %w", err)
		}

		// Check if -x flag was used (history)
		showHistory, err := cmd.Flags().GetBool("x")
		if err != nil {
			return fmt.Errorf("failed to get -x flag: %w", err)
		}

		// Check if -f flag was used (configure)
		configure, err := cmd.Flags().GetBool("f")
		if err != nil {
			return fmt.Errorf("failed to get -f flag: %w", err)
		}

		// Check if -v flag was used (version)
		showVersion, err := cmd.Flags().GetBool("v")
		if err != nil {
			return fmt.Errorf("failed to get -v flag: %w", err)
		}

		// If -c flag is provided, delegate to run command
		if goal != "" {
			// Get the confirm flag value
			confirm, err := cmd.Flags().GetBool("confirm")
			if err != nil {
				return fmt.Errorf("failed to get confirm flag: %w", err)
			}

			// Set the arguments and flags for the run command
			runCmd.SetArgs([]string{goal})
			if err := runCmd.Flags().Set("confirm", fmt.Sprintf("%t", confirm)); err != nil {
				return fmt.Errorf("failed to set confirm flag: %w", err)
			}

			// Execute the run command
			return runCmd.RunE(runCmd, []string{goal})
		}

		// If -x flag is provided, delegate to history list command
		if showHistory {
			return historyListCmd.RunE(historyListCmd, []string{})
		}

		// If -f flag is provided, delegate to configure command
		if configure {
			return configureCmd.RunE(configureCmd, []string{})
		}

		// If -v flag is provided, delegate to version command
		if showVersion {
			if versionCmd != nil {
				return versionCmd.RunE(versionCmd, []string{})
			}
			return fmt.Errorf("version command not initialized")
		}

		// If no flags, show help
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k8x/config.yaml)")

	// Add -c flag for quick goal execution (delegates to run command)
	rootCmd.Flags().StringP("c", "c", "", "Run a new k8x session with a goal")

	// Add -x flag for history (delegates to history list command)
	rootCmd.Flags().BoolP("x", "x", false, "Show command history")

	// Add -f flag for configure (delegates to configure command)
	rootCmd.Flags().BoolP("f", "f", false, "Initialize k8x workspace and configuration")

	// Add -v flag for version (delegates to version command)
	rootCmd.Flags().BoolP("v", "v", false, "Show k8x version information")

	// Add --confirm flag for confirmation mode
	rootCmd.Flags().BoolP("confirm", "a", false, "Ask for confirmation before executing each tool")

	// Ensure versionCmd is added as a subcommand
	if versionCmd != nil {
		rootCmd.AddCommand(versionCmd)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in ~/.k8x directory with name "config" (without extension).
		k8xDir := filepath.Join(home, ".k8x")
		viper.AddConfigPath(k8xDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
