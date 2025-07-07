package framework

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ClusterFixture represents a kind cluster for e2e testing
type ClusterFixture struct {
	Name           string
	KubeConfigPath string
	t              *testing.T
	ctx            context.Context
}

// NewClusterFixture creates a new kind cluster fixture
func NewClusterFixture(t *testing.T, name string) (*ClusterFixture, error) {
	t.Helper()

	// Create a temp directory for the kubeconfig
	tempDir, err := os.MkdirTemp("", "k8x-e2e-test")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	kubeConfigPath := filepath.Join(tempDir, "kubeconfig")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	t.Logf("Creating kind cluster %q...", name)

	// Create the kind cluster
	cmd := exec.CommandContext(
		ctx,
		"kind",
		"create", "cluster",
		"--name", name,
		"--kubeconfig", kubeConfigPath,
		"--wait", "60s",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create kind cluster: %w, stderr: %s", err, stderr.String())
	}

	t.Logf("Kind cluster %q created successfully", name)

	return &ClusterFixture{
		Name:           name,
		KubeConfigPath: kubeConfigPath,
		t:              t,
		ctx:            context.Background(),
	}, nil
}

// Cleanup deletes the kind cluster and temporary files
func (c *ClusterFixture) Cleanup() {
	c.t.Helper()
	c.t.Logf("Cleaning up kind cluster %q...", c.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"kind",
		"delete", "cluster",
		"--name", c.Name,
	)

	if err := cmd.Run(); err != nil {
		c.t.Logf("Warning: failed to delete kind cluster: %v", err)
	}

	// Remove temp directory containing kubeconfig
	if c.KubeConfigPath != "" {
		tempDir := filepath.Dir(c.KubeConfigPath)
		if err := os.RemoveAll(tempDir); err != nil {
			c.t.Logf("Warning: failed to remove temp directory: %v", err)
		}
	}
}

// ApplyManifest applies a YAML manifest to the cluster
func (c *ClusterFixture) ApplyManifest(yamlContent string) error {
	c.t.Helper()

	// Create a temporary file for the manifest
	tmpfile, err := os.CreateTemp("", "manifest-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if removeErr := os.Remove(tmpfile.Name()); removeErr != nil {
			c.t.Logf("Warning: failed to remove temp file: %v", removeErr)
		}
	}()

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"kubectl",
		"--kubeconfig", c.KubeConfigPath,
		"apply", "-f", tmpfile.Name(),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply manifest: %w, stderr: %s", err, stderr.String())
	}

	return nil
}

// WaitForPod waits for a pod to be ready or in a specific state
func (c *ClusterFixture) WaitForPod(namespace, podNamePrefix, condition string, timeout time.Duration) error {
	c.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.t.Logf("Waiting for pod %s in namespace %s to be %s...", podNamePrefix, namespace, condition)

	args := []string{
		"--kubeconfig", c.KubeConfigPath,
		"-n", namespace,
		"wait", "pod",
		"--for", "condition=" + condition,
		"--selector", fmt.Sprintf("app=%s", podNamePrefix),
		"--timeout", timeout.String(),
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pod never reached %s state: %w, stderr: %s", condition, err, stderr.String())
	}

	return nil
}

// WaitForPodStatus waits for a pod to reach a specific status condition like CrashLoopBackOff
func (c *ClusterFixture) WaitForPodStatus(namespace, podNamePrefix, status string, timeout time.Duration) error {
	c.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.t.Logf("Waiting for pod %s in namespace %s to reach status %s...", podNamePrefix, namespace, status)

	// Poll the pod status until it matches what we expect
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pod %s to reach status %s", podNamePrefix, status)
		case <-ticker.C:
			// Get pod status
			cmd := exec.CommandContext(
				context.Background(),
				"kubectl",
				"--kubeconfig", c.KubeConfigPath,
				"-n", namespace,
				"get", "pods",
				"--selector", fmt.Sprintf("app=%s", podNamePrefix),
				"-o", "jsonpath={.items[0].status.containerStatuses[0].state}",
			)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				c.t.Logf("Error getting pod status: %v, stderr: %s", err, stderr.String())
				continue
			}

			statusOutput := strings.TrimSpace(stdout.String())
			c.t.Logf("Pod status: %s", statusOutput)

			// Check if the status contains what we're looking for
			if strings.Contains(strings.ToLower(statusOutput), strings.ToLower(status)) {
				c.t.Logf("Pod %s reached status %s", podNamePrefix, status)
				return nil
			}

			// Also check phase for some conditions
			phaseCmd := exec.CommandContext(
				context.Background(),
				"kubectl",
				"--kubeconfig", c.KubeConfigPath,
				"-n", namespace,
				"get", "pods",
				"--selector", fmt.Sprintf("app=%s", podNamePrefix),
				"-o", "jsonpath={.items[0].status.phase}",
			)

			var phaseStdout, phaseStderr bytes.Buffer
			phaseCmd.Stdout = &phaseStdout
			phaseCmd.Stderr = &phaseStderr

			if err := phaseCmd.Run(); err == nil {
				phaseOutput := strings.TrimSpace(phaseStdout.String())
				if strings.Contains(strings.ToLower(phaseOutput), strings.ToLower(status)) {
					c.t.Logf("Pod %s reached phase %s", podNamePrefix, status)
					return nil
				}
			}
		}
	}
}

// RunK8x executes the k8x command against the cluster
func (c *ClusterFixture) RunK8x(args ...string) (string, error) {
	c.t.Helper()

	// Determine the path to the k8x binary (assumed to be in the build directory)
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up to find the repository root
	repoRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			return "", fmt.Errorf("could not find repository root")
		}
		repoRoot = parent
	}

	k8xPath := filepath.Join(repoRoot, "build", "k8x")
	if _, err := os.Stat(k8xPath); os.IsNotExist(err) {
		return "", fmt.Errorf("k8x binary not found at %s, please build it first", k8xPath)
	}

	// Create a temporary directory for k8x config
	tempDir, err := os.MkdirTemp("", "k8x-config")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory for k8x config: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			c.t.Logf("Warning: failed to remove temp directory: %v", removeErr)
		}
	}()

	// Set HOME to the temporary directory for isolated config
	env := os.Environ()
	env = append(env, fmt.Sprintf("HOME=%s", tempDir))
	env = append(env, fmt.Sprintf("KUBECONFIG=%s", c.KubeConfigPath))

	// Set up k8x config
	k8xConfigDir := filepath.Join(tempDir, ".k8x")
	if err := os.MkdirAll(k8xConfigDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create k8x config directory: %w", err)
	}

	// Create a minimal config file
	configContent := fmt.Sprintf(`
llm:
  default_provider: "openai"
  providers:
    openai:
      model: "gpt-4"
kubernetes:
  context: "kind-%s"
  namespace: "k8x-test"
settings:
  history_enabled: true
`, c.Name)

	if err := os.WriteFile(filepath.Join(k8xConfigDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write k8x config file: %w", err)
	}

	// Create credentials file
	var apiKey string

	// Check for API key in environment variables (CI)
	apiKey = os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Use a dummy key for local development - this won't work but allows tests to run
		apiKey = "test-api-key"
		c.t.Log("Warning: Using dummy API key. Tests will fail without valid API keys.")
	}

	credsContent := fmt.Sprintf(`
selected_provider: openai
openai:
  api_key: "%s"
`, apiKey)

	if err := os.WriteFile(filepath.Join(k8xConfigDir, "credentials"), []byte(credsContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write k8x credentials file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Prepare command
	fullArgs := append([]string{"run"}, args...)
	cmd := exec.CommandContext(ctx, k8xPath, fullArgs...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String()

	if err != nil {
		return output, fmt.Errorf("k8x command failed: %w\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String())
	}

	return output, nil
}

// GetPodLogs retrieves logs from a pod
func (c *ClusterFixture) GetPodLogs(namespace, podNamePrefix string) (string, error) {
	c.t.Helper()

	// First get the actual pod name
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"kubectl",
		"--kubeconfig", c.KubeConfigPath,
		"-n", namespace,
		"get", "pods",
		"--selector", fmt.Sprintf("app=%s", podNamePrefix),
		"-o", "jsonpath={.items[0].metadata.name}",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get pod name: %w, stderr: %s", err, stderr.String())
	}

	podName := strings.TrimSpace(stdout.String())
	if podName == "" {
		return "", fmt.Errorf("pod name not found")
	}

	// Get the logs
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Minute)
	defer cancel2()

	logsCmd := exec.CommandContext(
		ctx2,
		"kubectl",
		"--kubeconfig", c.KubeConfigPath,
		"-n", namespace,
		"logs", podName,
	)

	var logsStdout, logsStderr bytes.Buffer
	logsCmd.Stdout = &logsStdout
	logsCmd.Stderr = &logsStderr

	if err := logsCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w, stderr: %s", err, logsStderr.String())
	}

	return logsStdout.String(), nil
}
