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
	Short: "Interactive AI-powered Kubernetes console",
	Long: `k8x is an interactive AI-powered console that acts as an intelligent layer on top of kubectl.
It helps you manage Kubernetes resources through natural language commands and
provides automated assistance for common operations.

Simply run 'k8x' to start the interactive console.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Always launch interactive console
		if consoleCmd != nil {
			return consoleCmd.RunE(consoleCmd, []string{})
		}

		// Fallback error if console not available
		return fmt.Errorf("console command not initialized")
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

	// Config file flag is kept for advanced users who want to specify a custom config
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k8x/config.yaml)")

	// Remove all subcommands except console - they're now slash commands
	// This keeps the binary clean and simple
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
