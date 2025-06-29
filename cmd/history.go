package cmd

import (
	"fmt"

	"github.com/shankgan/k8x/internal/history"
	"github.com/spf13/cobra"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Manage command history",
	Long: `View and manage k8x command history. All commands are automatically
tracked and can be reviewed, replayed, or undone.`,
}

// historyListCmd represents the history list command
var historyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List command history",
	Long:  `List all recorded command history entries`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := history.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create history manager: %w", err)
		}

		files, err := manager.List()
		if err != nil {
			return fmt.Errorf("failed to list history: %w", err)
		}

		if len(files) == 0 {
			fmt.Println("No command history found")
			return nil
		}

		fmt.Printf("Found %d history entries:\n", len(files))
		for i, file := range files {
			fmt.Printf("%d. %s\n", i+1, file)
		}

		return nil
	},
}

// historyShowCmd represents the history show command
var historyShowCmd = &cobra.Command{
	Use:   "show <filename>",
	Short: "Show details of a history entry",
	Long:  `Show detailed information about a specific history entry`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		manager, err := history.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create history manager: %w", err)
		}

		entry, err := manager.Load(filename)
		if err != nil {
			return fmt.Errorf("failed to load history entry: %w", err)
		}

		fmt.Printf("ID: %s\n", entry.ID)
		fmt.Printf("Goal: %s\n", entry.Goal)
		fmt.Printf("Command: %s\n", entry.Command)
		fmt.Printf("Args: %v\n", entry.Args)
		fmt.Printf("Timestamp: %s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("Status: %s\n", entry.Status)

		if entry.Output != "" {
			fmt.Printf("Output:\n%s\n", entry.Output)
		}

		if entry.Error != "" {
			fmt.Printf("Error:\n%s\n", entry.Error)
		}

		return nil
	},
}

// historyClearCmd represents the history clear command
var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear command history",
	Long:  `Clear all command history entries`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement history clearing
		fmt.Println("History clearing not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.AddCommand(historyListCmd)
	historyCmd.AddCommand(historyShowCmd)
	historyCmd.AddCommand(historyClearCmd)
}
