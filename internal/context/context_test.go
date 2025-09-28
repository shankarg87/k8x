package context

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// parseShellHistory extracts the shell history parsing logic for testing
func parseShellHistory(historyPath string) ([]string, error) {
	if historyPath == "" {
		return nil, nil
	}

	file, err := os.Open(historyPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	cmds := []string{}

	// Updated logic that handles multi-line commands
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

	return cmds, nil
}

func TestShellHistoryMultilineCommands(t *testing.T) {
	// Create a temporary history file with multi-line commands
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "test_history")

	// Sample history content with multi-line commands (using backslash continuation)
	historyContent := `cd /tmp
kubectl get pods \
  --namespace=default \
  --output=wide
kubectl describe pod my-pod \
  --namespace=production
helm install my-release \
  --namespace=production \
  --set key=value \
  my-chart
kubectl apply -f deployment.yaml
: 1234567890:0;ls -la
kubectl delete pod \
  old-pod \
  --grace-period=0
`

	err := os.WriteFile(historyFile, []byte(historyContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test history file: %v", err)
	}

	// Test the updated shell history parsing
	cmds, err := parseShellHistory(historyFile)
	if err != nil {
		t.Fatalf("parseShellHistory failed: %v", err)
	}

	// After fix, multi-line commands should be captured completely
	for _, cmd := range cmds {
		t.Logf("Captured command: %q", cmd)
	}

	// Verify that multi-line commands are now captured completely
	expectedCommands := map[string]bool{
		"kubectl get pods \\\n  --namespace=default \\\n  --output=wide":                            false,
		"kubectl describe pod my-pod \\\n  --namespace=production":                                  false,
		"helm install my-release \\\n  --namespace=production \\\n  --set key=value \\\n  my-chart": false,
		"kubectl delete pod \\\n  old-pod \\\n  --grace-period=0":                                   false,
		"kubectl apply -f deployment.yaml":                                                          false,
	}

	for _, cmd := range cmds {
		for expectedCmd := range expectedCommands {
			if cmd == expectedCmd {
				expectedCommands[expectedCmd] = true
			}
		}
	}

	// Check that all expected commands were found
	for expectedCmd, found := range expectedCommands {
		if !found {
			t.Errorf("Expected multi-line command not found: %q", expectedCmd)
		}
	}
}

func TestShellHistoryWithZshTimestamps(t *testing.T) {
	// Create a temporary history file with zsh timestamp format
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "test_zsh_history")

	// Sample zsh history content with timestamps and multi-line commands
	historyContent := `: 1234567890:0;kubectl get pods
: 1234567891:0;kubectl describe pod \
my-pod \
--namespace=production
: 1234567892:0;ls -la
`

	err := os.WriteFile(historyFile, []byte(historyContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test zsh history file: %v", err)
	}

	// Test the shell history parsing
	cmds, err := parseShellHistory(historyFile)
	if err != nil {
		t.Fatalf("parseShellHistory failed: %v", err)
	}

	// After fix, zsh timestamp commands should be captured
	for _, cmd := range cmds {
		t.Logf("Captured command: %q", cmd)
	}

	// Verify that both single-line and multi-line zsh commands are captured
	expectedCommands := map[string]bool{
		"kubectl get pods": false,
		"kubectl describe pod \\\nmy-pod \\\n--namespace=production": false,
	}

	for _, cmd := range cmds {
		for expectedCmd := range expectedCommands {
			if cmd == expectedCmd {
				expectedCommands[expectedCmd] = true
			}
		}
	}

	// Check that all expected commands were found
	for expectedCmd, found := range expectedCommands {
		if !found {
			t.Errorf("Expected zsh command not found: %q", expectedCmd)
		}
	}
}

// Test edge cases for shell history parsing
func TestShellHistoryEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "edge_cases_history")

	// Test various edge cases
	historyContent := `: 1234567890:0;
: 1234567891:0;kubectl get pods
kubectl apply -f \
file1.yaml \
file2.yaml \
file3.yaml
: 1234567892:0;some-other-command
helm upgrade release-name \
  chart-name \
  --set value1=test \
  --set value2=test
: 1234567893:0;
kubectl delete pod \
  some-pod
kustomize build .
`

	err := os.WriteFile(historyFile, []byte(historyContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create edge cases history file: %v", err)
	}

	cmds, err := parseShellHistory(historyFile)
	if err != nil {
		t.Fatalf("parseShellHistory failed: %v", err)
	}

	for _, cmd := range cmds {
		t.Logf("Captured command: %q", cmd)
	}

	// Verify expected commands
	expectedCommands := []string{
		"kustomize build .",
		"kubectl delete pod \\\n  some-pod",
		"helm upgrade release-name \\\n  chart-name \\\n  --set value1=test \\\n  --set value2=test",
		"kubectl apply -f \\\nfile1.yaml \\\nfile2.yaml \\\nfile3.yaml",
		"kubectl get pods",
	}

	if len(cmds) != len(expectedCommands) {
		t.Errorf("Expected %d commands, got %d", len(expectedCommands), len(cmds))
	}

	for i, expectedCmd := range expectedCommands {
		if i >= len(cmds) {
			t.Errorf("Missing expected command: %q", expectedCmd)
			continue
		}
		if cmds[i] != expectedCmd {
			t.Errorf("Command %d mismatch:\nExpected: %q\nGot: %q", i, expectedCmd, cmds[i])
		}
	}
}
