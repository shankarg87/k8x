package context

import (
	"bufio"
	"fmt"
	"k8x/internal/llm"
	"os"
	"strings"
)

// ContextInfo holds all relevant cluster and shell context for k8x
// Add more fields as needed for future context expansion
// This struct is designed for easy extension and testing
// You can add more fields for additional context sources

type ContextInfo struct {
	KubectlVersion string
	ClusterVersion string
	Namespaces     string
	ToolsCheck     string
	HelmReleases   string
	RecentExamples string
}

// BuildContextInfo gathers cluster, tool, helm, and shell history context for k8x
// All shell execution and file operations should be injected for testability
func BuildContextInfo(
	toolManager *llm.ToolManager,
	historyFiles []string,
) (*ContextInfo, error) {
	var (
		kubectlVersion string
		clusterVersion string
		namespaces     string
		toolsCheck     string
		helmReleases   string
		recentExamples string
	)

	// Get kubectl version
	kubectlVersion, err := toolManager.ExecuteTool("execute_shell_command", `{"command": "kubectl version --client --output=yaml | grep 'gitVersion:' | head -1 | awk '{print $2}'"}`)
	if err != nil || strings.TrimSpace(kubectlVersion) == "" {
		kubectlVersionFallback, errFallback := toolManager.ExecuteTool("execute_shell_command", `{"command": "kubectl version --client --short"}`)
		if errFallback != nil {
			kubectlVersion = fmt.Sprintf("Error getting kubectl version: %v (fallback error: %v)", err, errFallback)
		} else {
			kubectlVersion = strings.TrimSpace(kubectlVersionFallback)
		}
	} else {
		kubectlVersion = strings.TrimSpace(kubectlVersion)
	}

	// Get cluster version
	clusterVersion, err = toolManager.ExecuteTool("execute_shell_command", `{"command": "kubectl version --output=yaml 2>/dev/null | grep 'gitVersion:' | tail -1 | awk '{print $2}'"}`)
	if err != nil || strings.TrimSpace(clusterVersion) == "" {
		clusterVersion = "No cluster connection available"
	} else {
		clusterVersion = strings.TrimSpace(clusterVersion)
	}

	// Get namespaces
	namespaces, err = toolManager.ExecuteTool("execute_shell_command", `{"command": "kubectl get namespaces --output=name"}`)
	if err != nil {
		namespaces = "No cluster connection available"
	} else {
		namespaceList := strings.Split(strings.TrimSpace(namespaces), "\n")
		namespaces = strings.Join(namespaceList, ", ")
	}

	// Check for common tools
	toolsCheck = ""
	var helmAvailable bool
	for _, tool := range []string{"kubectl", "helm", "kustomize", "jq"} {
		result, err := toolManager.ExecuteTool("execute_shell_command", fmt.Sprintf(`{"command": "which %s"}`, tool))
		if err == nil && strings.TrimSpace(result) != "" {
			version, _ := toolManager.ExecuteTool("execute_shell_command", fmt.Sprintf(`{"command": "%s version --short 2>/dev/null || %s --version 2>/dev/null || echo 'version unknown'"}`, tool, tool))
			toolsCheck += fmt.Sprintf("- %s: %s (%s)\n", tool, strings.TrimSpace(result), strings.TrimSpace(version))
			if tool == "helm" {
				helmAvailable = true
			}
		} else {
			toolsCheck += fmt.Sprintf("- %s: not available\n", tool)
		}
	}

	// Helm releases
	if helmAvailable {
		releases, err := toolManager.ExecuteTool("execute_shell_command", `{"command": "helm list --all-namespaces"}`)
		if err != nil {
			helmReleases = fmt.Sprintf("Error getting Helm releases: %v", err)
		} else {
			helmReleases = releases
		}
	} else {
		helmReleases = "Helm not available"
	}

	// Recent CLI examples from shell history
	historyPath := ""
	for _, file := range historyFiles {
		path := file
		if strings.HasPrefix(file, "~") {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				path = strings.Replace(file, "~", homeDir, 1)
			}
		}
		if _, err := os.Stat(path); err == nil {
			historyPath = path
			break
		}
	}
	cmds := []string{}
	if historyPath != "" {
		file, err := os.Open(historyPath)
		if err == nil {
			defer func() {
				if err := file.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "error closing file %s: %v\n", file.Name(), err)
				}
			}()
			scanner := bufio.NewScanner(file)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			if err := scanner.Err(); err == nil {
				// Process commands from oldest to newest to maintain history order
				// but collect them in reverse to show most recent first
				var allCommands []string
				allCmdSet := make(map[string]struct{})
				
				i := 0
				for i < len(lines) {
					line := lines[i]
					
					// Skip empty lines
					if line == "" {
						i++
						continue
					}
					
					// Handle zsh timestamp format: ": timestamp:duration;command"
					if strings.HasPrefix(line, ":") {
						semicolonIndex := strings.Index(line, ";")
						if semicolonIndex != -1 && semicolonIndex < len(line)-1 {
							line = line[semicolonIndex+1:]
						} else {
							i++
							continue
						}
					}
					
					// Check if this line starts a command we're interested in
					if strings.HasPrefix(line, "kubectl") || strings.HasPrefix(line, "helm") || strings.HasPrefix(line, "kustomize") {
						// Reconstruct the complete command, handling multi-line continuations
						fullCommand := line
						j := i + 1
						
						// Look forward for continuation lines (current line ends with \)
						for strings.HasSuffix(fullCommand, "\\") && j < len(lines) {
							nextLine := lines[j]
							
							// Handle zsh timestamp format in continuation lines
							if strings.HasPrefix(nextLine, ":") {
								semicolonIndex := strings.Index(nextLine, ";")
								if semicolonIndex != -1 && semicolonIndex < len(nextLine)-1 {
									nextLine = nextLine[semicolonIndex+1:]
								} else {
									// If it's just a timestamp line without command, stop
									break
								}
							}
							
							// Add the continuation line
							fullCommand += "\n" + nextLine
							j++
						}
						
						// Add the complete command if we haven't seen it before
						if _, exists := allCmdSet[fullCommand]; !exists {
							allCmdSet[fullCommand] = struct{}{}
							allCommands = append(allCommands, fullCommand)
						}
						
						// Skip the lines we've already processed as part of this command
						i = j
					} else {
						i++
					}
				}
				
				// Reverse the commands to show most recent first, and limit to 20
				for i := len(allCommands) - 1; i >= 0 && len(cmds) < 20; i-- {
					cmds = append(cmds, allCommands[i])
				}
				if len(cmds) > 20 {
					cmds = cmds[:20]
				}
				if len(cmds) > 0 {
					recentExamples = "Recent Examples:\n" + strings.Join(cmds, "\n")
				}
			}
		}
	}

	// Ensure context values have (unavailable) as default if empty
	if strings.TrimSpace(kubectlVersion) == "" {
		kubectlVersion = "(unavailable)"
	}
	if strings.TrimSpace(clusterVersion) == "" {
		clusterVersion = "(unavailable)"
	}
	if strings.TrimSpace(namespaces) == "" {
		namespaces = "(unavailable)"
	}
	if strings.TrimSpace(toolsCheck) == "" {
		toolsCheck = "(unavailable)"
	}
	if strings.TrimSpace(helmReleases) == "" {
		helmReleases = "(unavailable)"
	}
	if strings.TrimSpace(recentExamples) == "" {
		recentExamples = "(unavailable)"
	}

	return &ContextInfo{
		KubectlVersion: kubectlVersion,
		ClusterVersion: clusterVersion,
		Namespaces:     namespaces,
		ToolsCheck:     toolsCheck,
		HelmReleases:   helmReleases,
		RecentExamples: recentExamples,
	}, nil
}

// BuildContextInfoString gathers cluster context and returns a formatted string for LLM prompt
func BuildContextInfoString(toolManager *llm.ToolManager, historyFiles []string) (string, error) {
	fmt.Println("==============================")
	fmt.Println("📋 Cluster Information")
	fmt.Println("==============================")

	ctxInfo, err := BuildContextInfo(toolManager, historyFiles)
	if err != nil {
		return "", fmt.Errorf("failed to build context info: %w", err)
	}

	fmt.Printf("kubectl Version:\n%s\n\n", ctxInfo.KubectlVersion)
	fmt.Printf("Cluster Version:\n%s\n\n", ctxInfo.ClusterVersion)
	fmt.Printf("Available Namespaces:\n%s\n\n", ctxInfo.Namespaces)
	fmt.Printf("Available CLI Commands:\n%s\n\n", ctxInfo.ToolsCheck)
	fmt.Printf("Helm Releases:\n%s\n\n", ctxInfo.HelmReleases)
	fmt.Printf("Recent CLI Examples (may be unoptimized, but useful for context):\n%s\n", ctxInfo.RecentExamples)
	fmt.Println("==============================")

	contextInfo := fmt.Sprintf(`Here's the current cluster context information: (use only the relevant information towards the goal)
================

kubectl Version:
%s

Cluster Version:
%s

Available Namespaces:
%s

Available CLI Commands:
%s

Helm Releases:
%s

Recent CLI Examples (may be unoptimized, but useful for context):
%s
`, ctxInfo.KubectlVersion, ctxInfo.ClusterVersion, ctxInfo.Namespaces, ctxInfo.ToolsCheck, ctxInfo.HelmReleases, ctxInfo.RecentExamples)

	return contextInfo, nil
}
