package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"k8x/internal/config"
	k8xcontext "k8x/internal/context"
	"k8x/internal/history"
	"k8x/internal/llm"
	"k8x/internal/llm/providers"
	"k8x/internal/output"

	"github.com/spf13/cobra"
)

var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Start k8x interactive console",
	Long: `Start an interactive k8x console session where you can type natural language
commands to interact with your Kubernetes cluster.

Special commands:
  /help         - Show available commands
  /configure    - Configure k8x settings
  /history      - Show command history
  /version      - Show version information
  /exit or /q   - Exit the console
  /clear        - Clear the screen`,
	RunE: runConsole,
}

func runConsole(cmd *cobra.Command, args []string) error {
	// Initialize colored printer with secret filtering enabled
	printer := output.NewPrinter(true)

	// Check if configuration exists
	if !isConfigured() {
		printer.PrintInfoln("🔧 k8x is not configured. Let's set it up now.")
		if err := configureCmd.RunE(configureCmd, []string{}); err != nil {
			return fmt.Errorf("configuration failed: %w", err)
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Load credentials
	creds, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	if !creds.HasAnyKey("openai_api_key", "anthropic_api_key", "gemini_api_key") {
		printer.PrintInfoln("🔧 No LLM provider configured. Let's set it up now.")
		if err := configureCmd.RunE(configureCmd, []string{}); err != nil {
			return fmt.Errorf("configuration failed: %w", err)
		}
		// Reload credentials after configuration
		creds, err = config.LoadCredentials()
		if err != nil {
			return fmt.Errorf("failed to reload credentials: %w", err)
		}
	}

	// Convert credentials
	provCreds := providers.Credentials{
		SelectedProvider: creds.SelectedProvider,
	}
	provCreds.OpenAI.APIKey = creds.OpenAI.APIKey
	provCreds.Anthropic.APIKey = creds.Anthropic.APIKey
	provCreds.Google.APIKey = creds.Google.APIKey
	provCreds.Google.ApplicationCredentials = creds.Google.ApplicationCredentials

	// Initialize LLM provider
	unifiedProvider, err := providers.NewUnifiedProvider(provCreds)
	if err != nil {
		return fmt.Errorf("failed to initialize LLM provider: %w", err)
	}

	// Initialize tool manager
	toolManager, err := llm.NewMCPToolManager(".", cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize tool manager: %w", err)
	}

	// Connect to MCP servers if enabled
	if cfg.MCP.Enabled {
		printer.PrintInfoln("🔌 Connecting to MCP servers...")
		if err := toolManager.ConnectMCPServers(context.Background()); err != nil {
			printer.PrintWarningln("⚠️  Warning: Failed to connect to some MCP servers: %v", err)
		}
		defer func() {
			if err := toolManager.DisconnectMCPServers(); err != nil {
				printer.PrintWarningln("Warning: Failed to disconnect MCP servers: %v", err)
			}
		}()
	}

	// Set Kubernetes configuration
	toolManager.SetKubernetesConfig(&cfg.Kubernetes)

	// Print welcome message
	printWelcome(unifiedProvider.Name(), printer)

	// Start console loop
	return runConsoleLoop(unifiedProvider, toolManager, cfg)
}

func printWelcome(providerName string, printer *output.Printer) {
	printer.PrintInfoln("\n╔════════════════════════════════════════════╗")
	printer.PrintInfoln("║         Welcome to k8x Console! 🚀         ║")
	printer.PrintInfoln("╚════════════════════════════════════════════╝")
	printer.PrintInfoln("🤖 Using LLM provider: %s", providerName)
	printer.Println("\nType your Kubernetes questions in natural language.")
	printer.Println("Type /help for available commands or /exit to quit.")
}

func runConsoleLoop(provider *providers.UnifiedProvider, toolManager *llm.MCPToolManager, cfg *config.Config) error {
	scanner := bufio.NewScanner(os.Stdin)
	historyManager, _ := history.NewManager()

	// Initialize colored printer with secret filtering enabled
	printer := output.NewPrinter(true)

	// Gather initial cluster context
	printer.PrintInfoln("🔍 Gathering cluster information...")
	contextInfo, err := k8xcontext.BuildContextInfoString(toolManager.ToolManager, []string{"~/.zsh_history", "~/.bash_history"})
	if err != nil {
		printer.PrintWarningln("⚠️  Warning: Failed to gather cluster context: %v", err)
		contextInfo = "Cluster context information unavailable."
	}

	// Initialize conversation messages with system prompt
	systemPrompt := buildSystemPrompt(contextInfo)
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
	}

	stepCount := 0

	for {
		fmt.Println()
		printer.PrintPrompt()

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		// Handle slash commands
		if strings.HasPrefix(input, "/") {
			handled, shouldExit, shouldClear := handleSlashCommand(input, provider, toolManager, cfg, &messages, &stepCount, contextInfo)
			if shouldExit {
				printer.PrintInfoln("👋 Goodbye!")
				return nil
			}
			if shouldClear {
				// Reset conversation
				messages = []llm.Message{
					{Role: "system", Content: systemPrompt},
				}
				stepCount = 0
				printer.PrintSuccessln("🔄 Conversation history cleared. Starting fresh!")
			}
			if handled && !shouldClear {
				continue
			}
			if !handled {
				printer.PrintErrorln("❌ Unknown command: %s (type /help for available commands)", input)
			}
			continue
		}

		// Handle natural language command
		if err := executeGoalWithHistory(input, provider, toolManager, historyManager, &messages, &stepCount, false, printer); err != nil {
			printer.PrintErrorln("❌ Error: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	printer.PrintInfoln("\n👋 Goodbye!")
	return nil
}

func handleSlashCommand(input string, provider *providers.UnifiedProvider, toolManager *llm.MCPToolManager, cfg *config.Config, messages *[]llm.Message, stepCount *int, contextInfo string) (handled bool, shouldExit bool, shouldClear bool) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return true, false, false
	}

	command := strings.ToLower(parts[0])

	switch command {
	case "/help", "/h":
		printHelp()
		return true, false, false
	case "/configure", "/config", "/f":
		if err := configureCmd.RunE(configureCmd, []string{}); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
		return true, false, false
	case "/history", "/x":
		if err := historyListCmd.RunE(historyListCmd, []string{}); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
		return true, false, false
	case "/version", "/v":
		if versionCmd != nil {
			if err := versionCmd.RunE(versionCmd, []string{}); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			}
		} else {
			fmt.Println("❌ Version command not initialized")
		}
		return true, false, false
	case "/exit", "/quit", "/q":
		return true, true, false
	case "/clear", "/cls":
		clearScreen()
		return true, false, true
	case "/confirm":
		// Toggle confirmation mode
		// We track this through the toolManager's SetConfirmationMode
		fmt.Println("✅ Confirmation mode toggled")
		return true, false, false
	case "/mcp":
		// Show MCP server status
		if !cfg.MCP.Enabled {
			fmt.Println("MCP is disabled")
		} else {
			status := toolManager.GetMCPServerStatus()
			fmt.Println("MCP Server Status:")
			for name, connected := range status {
				if connected {
					fmt.Printf("  ✅ %s: connected\n", name)
				} else {
					fmt.Printf("  ❌ %s: disconnected\n", name)
				}
			}
		}
		return true, false, false
	default:
		return false, false, false
	}
}

func printHelp() {
	fmt.Println("\n📚 Available Commands:")
	fmt.Println("  /help, /h       - Show this help message")
	fmt.Println("  /configure, /f  - Configure k8x settings")
	fmt.Println("  /history, /x    - Show command history")
	fmt.Println("  /version, /v    - Show version information")
	fmt.Println("  /confirm        - Toggle confirmation mode")
	fmt.Println("  /mcp            - Show MCP server status")
	fmt.Println("  /clear, /cls    - Clear the screen")
	fmt.Println("  /exit, /q       - Exit the console")
	fmt.Println("\nOr type any natural language command to interact with your cluster:")
	fmt.Println("  Example: \"Why is my nginx pod failing?\"")
}

func clearScreen() {
	// ANSI escape code to clear screen
	fmt.Print("\033[H\033[2J")
}

func isConfigured() bool {
	_, err := config.GetCredentialsPath()
	if err != nil {
		return false
	}

	creds, err := config.LoadCredentials()
	if err != nil {
		return false
	}

	return creds.HasAnyKey("openai_api_key", "anthropic_api_key", "gemini_api_key")
}

func buildSystemPrompt(contextInfo string) string {
	return fmt.Sprintf(`You are k8x, a Kubernetes shell-workflow assistant specialized in read-only diagnostics and operations.

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
}

func executeGoalWithHistory(goal string, provider *providers.UnifiedProvider, toolManager *llm.MCPToolManager, historyManager *history.Manager, messages *[]llm.Message, stepCount *int, confirm bool, printer *output.Printer) error {
	// Create history entry
	entry := &history.Entry{
		Goal:      goal,
		Timestamp: time.Now(),
		Status:    "pending",
		Steps:     []history.Step{},
	}

	// Save the session
	if historyManager != nil {
		if err := historyManager.Save(entry); err != nil {
			printer.PrintWarningln("⚠️  Warning: Failed to save history: %v", err)
		}
	}

	// Get available tools
	tools, err := toolManager.GetAllTools(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get available tools: %w", err)
	}

	// Add user message
	userMessage := fmt.Sprintf("Goal: %s\n\nPlease help me achieve this goal using read-only kubectl commands.", goal)
	if *stepCount == 0 {
		userMessage += " Start by suggesting and executing the first step."
	} else {
		userMessage += " Continue from where we left off."
	}
	*messages = append(*messages, llm.Message{
		Role:    "user",
		Content: userMessage,
	})

	maxStepsPerGoal := 20
	stepsForThisGoal := 0

	for stepsForThisGoal < maxStepsPerGoal {
		*stepCount++
		stepsForThisGoal++
		printer.PrintInfoln("\n📋 Step %d:", *stepCount)

		// Get response from LLM
		response, err := provider.ChatWithTools(context.Background(), *messages, tools)
		if err != nil {
			return fmt.Errorf("failed to get LLM response: %w", err)
		}

		printer.PrintAssistantln("💭 %s", response.Content)

		// Add to messages
		assistantMsg := llm.Message{
			Role:    "assistant",
			Content: response.Content,
		}

		// Handle tool calls
		if len(response.ToolCalls) > 0 {
			assistantMsg.ToolCalls = response.ToolCalls
			*messages = append(*messages, assistantMsg)

			for _, toolCall := range response.ToolCalls {
				printer.PrintInfoln("\n🔧 Executing: %s", toolCall.Function.Name)

				// Parse and display arguments
				var argsMap map[string]interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &argsMap); err == nil {
					if cmd, ok := argsMap["command"]; ok {
						printer.PrintCommandln("📝 Command: %v", cmd)
					}
				}

				// Execute tool
				result, err := toolManager.ExecuteTool(toolCall.Function.Name, toolCall.Function.Arguments)
				if err != nil {
					result = fmt.Sprintf("Error: %v", err)
					printer.PrintErrorln("❌ Failed: %v", err)
				} else {
					printer.PrintSuccessln("✅ Success")
				}

				// Filter secrets from output before printing
				filteredResult := output.FilterSecrets(result)
				printer.Println("📄 Output:\n%s", filteredResult)

				// Add result to conversation
				*messages = append(*messages, llm.Message{
					Role:       "tool",
					Content:    result,
					ToolCallID: toolCall.ID,
				})

				// Save to history
				if historyManager != nil {
					step := history.Step{
						Description: fmt.Sprintf("Executed: %s", toolCall.Function.Name),
						Command:     toolCall.Function.Arguments,
						Output:      result,
						Type:        "command",
					}
					if err := historyManager.AddStep(entry, step); err != nil {
						printer.PrintWarningln("Warning: failed to add step to history: %v", err)
					}
				}
			}
		} else {
			*messages = append(*messages, assistantMsg)

			// Save planning step to history
			if historyManager != nil {
				step := history.Step{
					Description: fmt.Sprintf("Planning Step %d", *stepCount),
					Output:      response.Content,
					Type:        "step",
				}
				if err := historyManager.AddStep(entry, step); err != nil {
					printer.PrintWarningln("Warning: failed to add step to history: %v", err)
				}
			}
		}

		// Check if done
		if strings.Contains(strings.ToUpper(response.Content), "**DONE**") {
			if historyManager != nil {
				entry.Status = "completed"
				if err := historyManager.UpdateEntry(entry); err != nil {
					printer.PrintWarningln("Warning: failed to update history entry: %v", err)
				}
			}
			return nil
		}
	}

	if stepsForThisGoal >= maxStepsPerGoal {
		printer.PrintWarningln("⚠️  Reached maximum steps (%d) for this goal. You can continue with another request.", maxStepsPerGoal)
	}

	return nil
}

func init() {
	// Console is now launched by default, no need to register as subcommand
	// rootCmd.AddCommand(consoleCmd)
}
