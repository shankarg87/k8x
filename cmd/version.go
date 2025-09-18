package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print version, commit, and build date information for k8x`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("k8x version: %s\ncommit: %s\nbuild date: %s\n", version, commit, date)
		return nil
	},
}

func init() {
	// Command is now accessible via /version in console
	// rootCmd.AddCommand(versionCmd)
}
