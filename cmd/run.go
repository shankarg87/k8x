package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"k8x/internal/config"
	k8xcontext "k8x/internal/context"
	"k8x/internal/history"
	"k8x/internal/llm"
	"k8x/internal/llm/providers"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run \"<goal>\"",
	Aliases: []string{"command", "-c"},
	Short:   "Run a new k8x session with a goal (preferred: k8x -c)",
	Long: `Start a new k8x session with a natural language goal.
This will create a new .k8x history file and begin an LLM-driven
planning and execution loop.

Preferred usage:
  k8x -c "Diagnose why my nginx pod is failing"

Also supported:
  k8x run "Diagnose why my nginx pod is failing"
  k8x command "Diagnose why my nginx pod is failing"
  k8x -c "Diagnose why my nginx pod is failing" --confirm
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		goal := args[0]
		if strings.TrimSpace(goal) == "" {
			return fmt.Errorf("goal cannot be empty")
		}

		// Get the confirm flag value
		confirm, err := cmd.Flags().GetBool("confirm")
		if err != nil {
			return fmt.Errorf("failed to get confirm flag: %w", err)
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

		// Save the new session file
		if err := manager.Save(entry); err != nil {
			return fmt.Errorf("failed to create session file: %w", err)
		}

		// First load the config to ensure LLM provider and Kubernetes are set up
		// Check if ~/.k8x/credentials exists and contains at least one required key
		_, err = config.GetCredentialsPath()
		if err != nil {
			return errors.New("k8x is not configured.\nHint: Please run `k8x configure`")
		}

		// Load configuration for Kubernetes settings
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Load credentials for LLM
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

		// Set confirmation mode
		toolManager.SetConfirmationMode(confirm)

		// Set Kubernetes configuration for the tool manager's shell executor
		toolManager.SetKubernetesConfig(&cfg.Kubernetes)
		tools := toolManager.GetTools()

		// Gather cluster context information before starting
		fmt.Println("🔍 Gathering cluster information...")

		// Build context info string using new function (prints as it gathers)
		contextInfo, err := k8xcontext.BuildContextInfoString(toolManager, []string{"~/.zsh_history", "~/.bash_history"})
		if err != nil {
			return fmt.Errorf("failed to build context info: %w", err)
		}

		// Prepare system message to set context for k8x
		systemPrompt := fmt.Sprintf(`You are k8x, a Kubernetes shell-workflow assistant specialized in read-only diagnostics and operations.

%s

Your role:
1. You help users achieve Kubernetes-related goals through step-by-step kubectl commands
2. You can ONLY perform READ-ONLY operations (get, describe, logs, etc.)
3.a. Break down complex goals into logical steps, but be fast and efficient.
3.b. You can always run commands to gather additional details if needed.
3.c. You can use pipe '|' to chain commands for efficiency.
3.d. You also have access to jq for JSON processing and can use it in commands.
4. Always explain what each kubectl command will do before suggesting it
5. Use the execute_shell_command function to run kubectl commands
6. Provide clear, actionable responses.
7. Your responses will be printed to the console.
Use colors and emojis to enhance readability.
YOU MUST NOT USE MARKDOWN formatting in your responses.

Available tools:
- execute_shell_command: Execute safe read-only shell commands, primarily kubectl operations

Current mode: READ-ONLY (no cluster modifications, no installations, no changes)

Guidelines:
- Only use safe, read-only commands: e.g. for kubectl - get, describe, logs, explain, version, etc.
- Do not use write operations: e.g. for kubectl - create, apply, delete, patch, edit, scale, etc.
- When you want to run a command, use the execute_shell_command function
- Explain what you're going to do before executing commands.
- If you achieve the goal or cannot proceed further, say "**DONE**."`, contextInfo)

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
			fmt.Printf("📋 Step %d:\n", stepCount)

			// Animated 'Thinking...' spinner
			thinkingDone := make(chan struct{})
			spinnerLine := ""
			go func() {
				spinner := []string{"   ", ".  ", ".. ", "..."}
				idx := 0
				for {
					select {
					case <-thinkingDone:
						return
					default:
						spinnerLine = fmt.Sprintf("\r💭 Thinking%s", spinner[idx])
						fmt.Print(spinnerLine)
						idx = (idx + 1) % len(spinner)
						time.Sleep(350 * time.Millisecond)
					}
				}
			}()

			// Get response from LLM with tools
			response, err := unifiedProvider.ChatWithTools(context.Background(), messages, tools)
			close(thinkingDone)
			// Clear spinner line and print response in its place
			fmt.Printf("\r%40s\r", "") // Clear spinner line
			if err != nil {
				return fmt.Errorf("failed to get LLM response: %w", err)
			}
			fmt.Printf("💭 %s\n", response.Content)

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

					// Parse arguments as JSON for better readability
					var argsMap map[string]interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &argsMap); err == nil {
						fmt.Println("📝 Arguments:")
						for k, v := range argsMap {
							fmt.Printf("  %s: %v\n", k, v)
						}
					} else {
						fmt.Printf("📝 Arguments: %s\n", toolCall.Function.Arguments)
					}

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
			if strings.Contains(strings.ToUpper(response.Content), "**DONE**") {
				entry.Status = "completed"
				if err := manager.UpdateEntry(entry); err != nil {
					return fmt.Errorf("failed to update entry status: %w", err)
				}
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

		// Final summary step: ask LLM to summarize session
		summaryPrompt := `To tell the user what was done in this session on shell console,
summarize each step taken as a simple checklist.
e.g. "Step 02: ✅ All pods in the production namespace listed." (use cross emoji for failed tool calls)
The last line should be a single sentence saying what was done. Followed by **DONE**.
- be clear and concise to summarize the user's original question/command.`

		messages = append(messages, llm.Message{
			Role:    "user",
			Content: summaryPrompt,
		})
		response, err := unifiedProvider.ChatWithTools(context.Background(), messages, tools)
		if err != nil {
			fmt.Printf("\n❌ Failed to get summary from LLM: %v\n", err)
		} else {
			fmt.Println("\n==============================")
			fmt.Println("📋 Session Summary Checklist")
			fmt.Println("==============================")
			fmt.Printf("%s\n", response.Content)
			fmt.Println("==============================")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Add confirm flag with alias a
	runCmd.Flags().BoolP("confirm", "a", false, "Ask for confirmation before executing each tool")
}
