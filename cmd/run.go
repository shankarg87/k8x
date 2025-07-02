package cmd

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"k8x/internal/config"
	"k8x/internal/history"
	"k8x/internal/llm"
	"k8x/internal/llm/providers"

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
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// First load the config to ensure LLM provider is set up
		// Check if ~/.k8x/credentials exists and contains at least one required key
		_, err = config.GetCredentialsPath()
		if err != nil {
			return errors.New("k8x is not configured.\nHint: Please run `k8x configure`")
		}
		creds, err := config.LoadCredentials()
		if err != nil {
			return errors.New("k8x is not configured correctly.\nHint: Please run `k8x configure`")
		}
		if !creds.HasAnyKey("openai_api_key", "anthropic_api_key", "google_application_credentials") {
			return errors.New("k8x cannot find any LLM configured.\nHint: Run 'k8x configure' to set up your LLM provider")
		}

		// Convert schemas.Credentials to providers.Credentials
		provCreds := providers.Credentials{
			SelectedProvider: creds.SelectedProvider,
		}
		provCreds.OpenAI.APIKey = creds.OpenAI.APIKey
		provCreds.Anthropic.APIKey = creds.Anthropic.APIKey
		provCreds.Google.ApplicationCredentials = creds.Google.ApplicationCredentials

		unifiedProvider, err := providers.NewUnifiedProvider(provCreds)
		if err != nil {
			return fmt.Errorf("failed to initialize LLM provider: %w", err)
		}
		fmt.Printf("🤖 Using LLM provider: %s\n", unifiedProvider.Name())

		// Initialize tool manager for shell execution
		toolManager := llm.NewToolManager(".")
		tools := toolManager.GetTools()

		// Prepare system message to set context for k8x
		systemPrompt := `You are k8x, a Kubernetes shell-workflow assistant specialized in read-only diagnostics and operations.

Your role:
1. You help users achieve Kubernetes-related goals through step-by-step kubectl commands
2. For this iteration, you can ONLY perform READ-ONLY operations (get, describe, logs, etc.)
3. Break down complex goals into logical steps
4. Always explain what each kubectl command will do before suggesting it
5. Use the execute_shell_command function to run kubectl commands
6. Provide clear, actionable responses

Available tools:
- execute_shell_command: Execute safe read-only shell commands, primarily kubectl operations

Current mode: READ-ONLY (no cluster modifications)

Guidelines:
- Only use safe, read-only kubectl commands like: get, describe, logs, explain, version, etc.
- Do not use write operations like: create, apply, delete, patch, edit, scale, etc.
- When you want to run a command, use the execute_shell_command function
- Explain what you're going to do before executing commands
- If you achieve the goal or cannot proceed further, say "GOAL_COMPLETE"`

		// Start conversation with system prompt and user goal
		messages := []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("Goal: %s\n\nPlease help me achieve this goal using read-only kubectl commands. Start by suggesting and executing the first step.", entry.Goal)},
		}

		stepCount := 0
		maxSteps := 20 // Maximum number of steps to prevent infinite loops

		for stepCount < maxSteps {
			stepCount++
			fmt.Println(strings.Repeat("=", 40))
			fmt.Printf("📋 Step %d: Consulting LLM...\n", stepCount)

			// Get response from LLM with tools
			response, err := unifiedProvider.ChatWithTools(context.Background(), messages, tools)
			if err != nil {
				return fmt.Errorf("failed to get LLM response: %w", err)
			}

			fmt.Printf("💭 LLM Response:\n%s\n", response.Content)

			// Add LLM response to conversation history
			assistantMsg := llm.Message{
				Role:    "assistant",
				Content: response.Content,
			}

			// Handle tool calls if present
			if len(response.ToolCalls) > 0 {
				assistantMsg.ToolCalls = response.ToolCalls
				messages = append(messages, assistantMsg)

				// Execute tool calls
				for _, toolCall := range response.ToolCalls {
					fmt.Printf("\n🔧 Executing tool: %s\n", toolCall.Function.Name)
					fmt.Printf("📝 Arguments: %s\n", toolCall.Function.Arguments)

					// Execute the tool
					result, err := toolManager.ExecuteTool(toolCall.Function.Name, toolCall.Function.Arguments)
					if err != nil {
						result = fmt.Sprintf("Error: %v", err)
						fmt.Printf("❌ Tool execution failed: %v\n", err)
					} else {
						fmt.Printf("✅ Tool execution successful\n")
					}

					fmt.Printf("📄 Output:\n%s\n", result)

					// Add tool result to conversation
					messages = append(messages, llm.Message{
						Role:       "tool",
						Content:    result,
						ToolCallID: toolCall.ID,
					})

					// Record the step in history
					step := history.Step{
						Description: fmt.Sprintf("Executed: %s", toolCall.Function.Name),
						Command:     toolCall.Function.Arguments,
						Output:      result,
						Type:        "command",
					}

					if err := manager.AddStep(entry, step); err != nil {
						return fmt.Errorf("failed to add step to history: %w", err)
					}
				}
			} else {
				// No tool calls, just add the assistant message
				messages = append(messages, assistantMsg)

				// Add LLM response to history as a step
				step := history.Step{
					Description: fmt.Sprintf("LLM Planning Step %d", stepCount),
					Command:     "", // No command for LLM planning steps
					Output:      response.Content,
					Type:        "step",
				}

				if err := manager.AddStep(entry, step); err != nil {
					return fmt.Errorf("failed to add step to history: %w", err)
				}
			}

			// Check if goal is complete
			if strings.Contains(strings.ToUpper(response.Content), "GOAL_COMPLETE") {
				entry.Status = "completed"
				if err := manager.UpdateEntry(entry); err != nil {
					return fmt.Errorf("failed to update entry status: %w", err)
				}
				fmt.Printf("\n🎉 Goal completed successfully!\n")
				return nil
			}
			fmt.Println(strings.Repeat("=", 40))
			fmt.Println()
		}

		// If we reached max steps, mark as incomplete
		if stepCount >= maxSteps {
			entry.Status = "incomplete"
			if err := manager.UpdateEntry(entry); err != nil {
				return fmt.Errorf("failed to update entry status: %w", err)
			}
			fmt.Printf("⚠️  Reached maximum number of steps (%d). Goal may not be fully achieved.\n", maxSteps)
		} else {
			entry.Status = "pending"
			if err := manager.UpdateEntry(entry); err != nil {
				return fmt.Errorf("failed to update entry status: %w", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

// generateKebabCase converts a string to kebab-case for filenames
func generateKebabCase(s string) string {
	kebab := strings.ToLower(strings.ReplaceAll(s, " ", "-"))
	return regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(kebab, "")
}
