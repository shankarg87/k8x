package cmd

import (
	"fmt"

	"k8x/internal/config"

	"github.com/spf13/cobra"
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server operations",
	Long:  `Model Context Protocol server operations for k8x integration.`,
}

// mcpListCmd represents the mcp list command
var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	Long:  `List all configured MCP servers and their status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if !cfg.MCP.Enabled {
			fmt.Println("MCP integration is disabled")
			fmt.Println("Enable it with: k8x mcp enable")
			return nil
		}

		if len(cfg.MCP.Servers) == 0 {
			fmt.Println("No MCP servers configured")
			return nil
		}

		fmt.Println("Configured MCP servers:")
		for name, server := range cfg.MCP.Servers {
			status := "disabled"
			if server.Enabled {
				status = "enabled"
			}
			fmt.Printf("  %s (%s)\n", name, status)
			fmt.Printf("    Command: %s %v\n", server.Command, server.Args)
			if server.Description != "" {
				fmt.Printf("    Description: %s\n", server.Description)
			}
		}
		return nil
	},
}

// mcpEnableCmd represents the mcp enable command
var mcpEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable MCP integration",
	Long:  `Enable Model Context Protocol integration to use MCP servers as tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement MCP enable functionality
		fmt.Println("MCP integration enabled")
		fmt.Println("Note: Configuration modification not yet implemented")
		return nil
	},
}

// mcpDisableCmd represents the mcp disable command
var mcpDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable MCP integration",
	Long:  `Disable Model Context Protocol integration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement MCP disable functionality
		fmt.Println("MCP integration disabled")
		fmt.Println("Note: Configuration modification not yet implemented")
		return nil
	},
}

// mcpAddCmd represents the mcp add command
var mcpAddCmd = &cobra.Command{
	Use:   "add <name> <command> [args...]",
	Short: "Add an MCP server configuration",
	Long:  `Add a new MCP server configuration with the specified command and arguments.`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		command := args[1]
		serverArgs := args[2:]

		description, _ := cmd.Flags().GetString("description")

		fmt.Printf("Adding MCP server '%s'\n", name)
		fmt.Printf("  Command: %s %v\n", command, serverArgs)
		if description != "" {
			fmt.Printf("  Description: %s\n", description)
		}
		fmt.Println("Note: Configuration modification not yet implemented")
		return nil
	},
}

// mcpRemoveCmd represents the mcp remove command
var mcpRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an MCP server configuration",
	Long:  `Remove an MCP server configuration by name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		fmt.Printf("Removing MCP server '%s'\n", name)
		fmt.Println("Note: Configuration modification not yet implemented")
		return nil
	},
}

func init() {
	// Command is now accessible via /mcp in console
	// rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpListCmd)
	mcpCmd.AddCommand(mcpEnableCmd)
	mcpCmd.AddCommand(mcpDisableCmd)
	mcpCmd.AddCommand(mcpAddCmd)
	mcpCmd.AddCommand(mcpRemoveCmd)

	// Add flags for MCP commands
	mcpAddCmd.Flags().StringP("description", "d", "", "Description of the MCP server")
}
