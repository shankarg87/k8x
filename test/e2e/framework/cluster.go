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

// CleanupConditional deletes the kind cluster based on failure state and flags
func (c *ClusterFixture) Cleanup(isFailure bool) {
	c.t.Helper()

	// Check if we should preserve on failure
	if isFailure && ShouldPreserveOnFailure() {
		c.t.Logf("Preserving kind cluster %q after test failure due to --preserve-on-failure flag", c.Name)
		c.t.Logf("To manually clean up later, run: kind delete cluster --name %s", c.Name)
		c.t.Logf("Kubeconfig available at: %s", c.KubeConfigPath)
		return
	}

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
	} else {
		c.t.Logf("Successfully deleted kind cluster %q", c.Name)
	}

	// Remove temp directory containing kubeconfig
	if c.KubeConfigPath != "" {
		tempDir := filepath.Dir(c.KubeConfigPath)
		if err := os.RemoveAll(tempDir); err != nil {
			c.t.Logf("Warning: failed to remove temp directory: %v", err)
		} else {
			c.t.Logf("Successfully removed temp directory: %s", tempDir)
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

// ExecuteCommands executes a list of shell commands sequentially
func (c *ClusterFixture) ExecuteCommands(commands [][]string) error {
	c.t.Helper()

	for i, cmdArgs := range commands {
		c.t.Logf("Executing command %d/%d: %s", i+1, len(commands), strings.Join(cmdArgs, " "))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)

		// Set KUBECONFIG environment for all commands (kubectl, helm, etc.)
		cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", c.KubeConfigPath))

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %d failed (%s): %w\nstdout: %s\nstderr: %s",
				i+1, strings.Join(cmdArgs, " "), err, stdout.String(), stderr.String())
		}

		c.t.Logf("Command %d completed successfully", i+1)
		if stdout.Len() > 0 {
			c.t.Logf("Command output: %s", stdout.String())
		}
	}

	return nil
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
		return "", fmt.Errorf("failed to write to k8x config file: %w", err)
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

// ResourceConditionChecker defines how to check if a resource meets a condition
type ResourceConditionChecker interface {
	Check(c *ClusterFixture) (bool, error)
	Description() string
}

// KubectlWaitConditionChecker uses kubectl wait command for standard conditions
type KubectlWaitConditionChecker struct {
	ResourceType string
	Name         string
	Namespace    string
	Condition    string
	Selector     string
	description  string
}

func (k *KubectlWaitConditionChecker) Check(c *ClusterFixture) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []string{
		"--kubeconfig", c.KubeConfigPath,
		"wait", k.ResourceType,
		"--for", k.Condition,
		"--timeout", "5s",
	}

	if k.Namespace != "" {
		args = append(args, "-n", k.Namespace)
	}

	if k.Selector != "" {
		args = append(args, "--selector", k.Selector)
	} else if k.Name != "" {
		args = append(args, k.Name)
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Log the error for debugging but don't fail - this is expected during waiting
		if stderr.Len() > 0 {
			c.t.Logf("kubectl wait failed (expected during polling): %v, stderr: %s", err, stderr.String())
		}
		return false, nil
	}
	return true, nil
}

func (k *KubectlWaitConditionChecker) Description() string {
	if k.description != "" {
		return k.description
	}
	return fmt.Sprintf("%s to have condition %s", k.ResourceType, k.Condition)
}

// WaitForResourceCondition is the unified waiting function
func (c *ClusterFixture) WaitForResourceCondition(checker ResourceConditionChecker, timeout time.Duration) error {
	c.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c.t.Logf("Waiting for %s...", checker.Description())

	// Check immediately first
	if satisfied, err := checker.Check(c); err != nil {
		return fmt.Errorf("initial check failed for %s: %w", checker.Description(), err)
	} else if satisfied {
		c.t.Logf("Condition already met: %s", checker.Description())
		return nil
	}

	ticker := time.NewTicker(5 * time.Second) // Reduced from 10s for faster polling
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s", checker.Description())
		case <-ticker.C:
			if satisfied, err := checker.Check(c); err != nil {
				c.t.Logf("Error checking condition: %v", err)
				continue
			} else if satisfied {
				c.t.Logf("Condition met: %s", checker.Description())
				return nil
			}
		}
	}
}

// PodStatusChecker checks for specific container states in pods
type PodStatusChecker struct {
	Namespace      string
	PodNamePrefix  string
	ExpectedStatus string
	description    string
}

func (p *PodStatusChecker) Check(c *ClusterFixture) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// For CrashLoopBackOff, we need to check both the waiting reason and restart count
	if p.ExpectedStatus == "CrashLoopBackOff" {
		// Get both waiting reason and restart count separately for more robust checking
		waitingCmd := exec.CommandContext(
			ctx,
			"kubectl",
			"--kubeconfig", c.KubeConfigPath,
			"-n", p.Namespace,
			"get", "pods",
			"--selector", fmt.Sprintf("app=%s", p.PodNamePrefix),
			"-o", "jsonpath={.items[*].status.containerStatuses[*].state.waiting.reason}",
		)

		var waitingStdout, waitingStderr bytes.Buffer
		waitingCmd.Stdout = &waitingStdout
		waitingCmd.Stderr = &waitingStderr

		restartCmd := exec.CommandContext(
			ctx,
			"kubectl",
			"--kubeconfig", c.KubeConfigPath,
			"-n", p.Namespace,
			"get", "pods",
			"--selector", fmt.Sprintf("app=%s", p.PodNamePrefix),
			"-o", "jsonpath={.items[*].status.containerStatuses[*].restartCount}",
		)

		var restartStdout, restartStderr bytes.Buffer
		restartCmd.Stdout = &restartStdout
		restartCmd.Stderr = &restartStderr

		waitingErr := waitingCmd.Run()
		restartErr := restartCmd.Run()

		waitingReason := strings.TrimSpace(waitingStdout.String())
		restartCountStr := strings.TrimSpace(restartStdout.String())

		c.t.Logf("Pod status check: waiting reason='%s', restart count='%s'", waitingReason, restartCountStr)

		// Check if the waiting reason is CrashLoopBackOff
		if waitingErr == nil && waitingReason == "CrashLoopBackOff" {
			c.t.Logf("Found CrashLoopBackOff state with restart count: %s", restartCountStr)
			return true, nil
		}

		// For early CrashLoopBackOff detection, if restart count > 0, check for Error state
		if restartErr == nil && restartCountStr != "" && restartCountStr != "0" {
			// Also check if the last state was terminated with Error
			terminatedCmd := exec.CommandContext(
				ctx,
				"kubectl",
				"--kubeconfig", c.KubeConfigPath,
				"-n", p.Namespace,
				"get", "pods",
				"--selector", fmt.Sprintf("app=%s", p.PodNamePrefix),
				"-o", "jsonpath={.items[*].status.containerStatuses[*].lastState.terminated.reason}",
			)

			var terminatedStdout bytes.Buffer
			terminatedCmd.Stdout = &terminatedStdout

			if terminatedCmd.Run() == nil {
				terminatedReason := strings.TrimSpace(terminatedStdout.String())
				if terminatedReason == "Error" {
					c.t.Logf("Found Error state with restart count %s, treating as CrashLoopBackOff", restartCountStr)
					return true, nil
				}
			}
		}

		return false, nil
	}

	// For other statuses, check waiting reason first
	cmd := exec.CommandContext(
		ctx,
		"kubectl",
		"--kubeconfig", c.KubeConfigPath,
		"-n", p.Namespace,
		"get", "pods",
		"--selector", fmt.Sprintf("app=%s", p.PodNamePrefix),
		"-o", "jsonpath={.items[*].status.containerStatuses[*].state.waiting.reason}",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Also check terminated state for some container states
		terminatedCmd := exec.CommandContext(
			ctx,
			"kubectl",
			"--kubeconfig", c.KubeConfigPath,
			"-n", p.Namespace,
			"get", "pods",
			"--selector", fmt.Sprintf("app=%s", p.PodNamePrefix),
			"-o", "jsonpath={.items[*].status.containerStatuses[*].lastState.terminated.reason}",
		)

		var terminatedStdout bytes.Buffer
		terminatedCmd.Stdout = &terminatedStdout

		if terminatedErr := terminatedCmd.Run(); terminatedErr != nil {
			c.t.Logf("Failed to get container state: waiting: %v, terminated: %v", err, terminatedErr)
			return false, nil
		}

		containerState := strings.TrimSpace(terminatedStdout.String())
		return containerState == p.ExpectedStatus, nil
	}

	containerState := strings.TrimSpace(stdout.String())
	return containerState == p.ExpectedStatus, nil
}

func (p *PodStatusChecker) Description() string {
	if p.description != "" {
		return p.description
	}
	return fmt.Sprintf("pod %s to have container status %s", p.PodNamePrefix, p.ExpectedStatus)
}

// WaitForPodStatus waits for a pod to reach a specific status condition
func (c *ClusterFixture) WaitForPodStatus(namespace, podNamePrefix, status string, timeout time.Duration) error {
	// Handle container states vs pod conditions differently
	if status == "CrashLoopBackOff" || status == "ImagePullBackOff" || status == "ErrImagePull" {
		// These are container states, use PodStatusChecker
		checker := &PodStatusChecker{
			Namespace:      namespace,
			PodNamePrefix:  podNamePrefix,
			ExpectedStatus: status,
			description:    fmt.Sprintf("pod %s to have container status %s", podNamePrefix, status),
		}
		return c.WaitForResourceCondition(checker, timeout)
	} else {
		// These are pod conditions, use KubectlWaitConditionChecker
		checker := &KubectlWaitConditionChecker{
			ResourceType: "pod",
			Namespace:    namespace,
			Condition:    fmt.Sprintf("condition=%s", status), // This should be "Ready", "PodScheduled", etc.
			Selector:     fmt.Sprintf("app=%s", podNamePrefix),
			description:  fmt.Sprintf("pod %s to have condition %s", podNamePrefix, status),
		}
		return c.WaitForResourceCondition(checker, timeout)
	}
}

// WaitForDeployment waits for a deployment to be ready
func (c *ClusterFixture) WaitForDeployment(namespace, deploymentName string, timeout time.Duration) error {
	checker := &KubectlWaitConditionChecker{
		ResourceType: "deployment",
		Name:         deploymentName,
		Namespace:    namespace,
		Condition:    "condition=Available",
		description:  fmt.Sprintf("deployment %s to be ready", deploymentName),
	}
	return c.WaitForResourceCondition(checker, timeout)
}

// WaitForConstraintTemplate waits for a ConstraintTemplate to be established
func (c *ClusterFixture) WaitForConstraintTemplate(name string, timeout time.Duration) error {
	checker := &KubectlWaitConditionChecker{
		ResourceType: "constrainttemplate",
		Name:         name,
		Condition:    "condition=Ready",
		description:  fmt.Sprintf("ConstraintTemplate %s to be established", name),
	}
	return c.WaitForResourceCondition(checker, timeout)
}

// WaitForConstraint waits for a constraint to be ready
func (c *ClusterFixture) WaitForConstraint(kind, name string, timeout time.Duration) error {
	checker := &KubectlWaitConditionChecker{
		ResourceType: strings.ToLower(kind),
		Name:         name,
		Condition:    "condition=Ready",
		description:  fmt.Sprintf("%s constraint %s to be ready", kind, name),
	}
	return c.WaitForResourceCondition(checker, timeout)
}

// WaitForCRD waits for a CRD to be established
func (c *ClusterFixture) WaitForCRD(crdName string, timeout time.Duration) error {
	checker := &KubectlWaitConditionChecker{
		ResourceType: "crd",
		Name:         crdName,
		Condition:    "condition=Established",
		description:  fmt.Sprintf("CRD %s to be established", crdName),
	}
	return c.WaitForResourceCondition(checker, timeout)
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
