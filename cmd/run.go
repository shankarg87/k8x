package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/shankgan/k8x/internal/history"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run \"<goal>\"",
	Short: "Run a new k8x session with a goal",
	Long: `Start a new k8x session with a natural language goal.
This will create a new .k8x history file and begin an LLM-driven
planning and execution loop.

Example:
  k8x run "Diagnose why my nginx pod is failing"
  k8x run "List all pods in the production namespace"
  k8x run "Check resource usage across all nodes"`,
	Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		goal := args[0]
		if strings.TrimSpace(goal) == "" {
			return fmt.Errorf("goal cannot be empty")
		}

		manager, err := history.NewManager()
		if err != nil {
			return fmt.Errorf("failed to create history manager: %w", err)
		}

		// Create new session entry
		entry := &history.Entry{
			Goal:      goal,
			Timestamp: time.Now(),
			Status:    "pending",
			Steps:     []history.Step{},
		}

		// Save the initial entry with just the goal
		if err := manager.Save(entry); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		fmt.Printf("Created new k8x session: %s\n", goal)
		fmt.Println("Session file created in ~/.k8x/history/")
		fmt.Println("\nNote: LLM-driven execution loop not yet implemented.")
		fmt.Println("This will be added in a future iteration.")

		// TODO: Implement LLM-driven planning and execution loop
		// 1. Load LLM provider from config
		// 2. Generate kubectl commands based on goal
		// 3. Execute commands and capture output
		// 4. Update session file with steps
		// 5. Continue until goal is achieved

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
