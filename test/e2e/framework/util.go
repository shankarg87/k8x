package framework

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CheckRequirements verifies that all required tools are installed
func CheckRequirements() error {
	requiredCommands := []string{"kind", "kubectl", "docker"}

	for _, cmd := range requiredCommands {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("required command not found: %s - please install it before running E2E tests", cmd)
		}
	}

	// Check that the k8x binary exists
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up to find the repository root
	repoRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			return fmt.Errorf("could not find repository root")
		}
		repoRoot = parent
	}

	k8xPath := filepath.Join(repoRoot, "build", "k8x")
	if _, err := os.Stat(k8xPath); os.IsNotExist(err) {
		return fmt.Errorf("k8x binary not found at %s, please run 'make build' first", k8xPath)
	}

	return nil
}

// CheckEnvironment checks the environment and prints warnings
func CheckEnvironment() error {
	// Check if Docker is running
	if err := exec.Command("docker", "info").Run(); err != nil {
		fmt.Println("Docker does not seem to be running. Please start Docker before running E2E tests.")
		return fmt.Errorf("Warning: Docker does not seem to be running. E2E tests will fail.")
	}

	// Check for API key
	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Println("To run complete tests, set the environment variable:")
		fmt.Println("    export OPENAI_API_KEY=your-api-key")
		return fmt.Errorf("Warning: OPENAI_API_KEY is not set. E2E tests will use a dummy API key and may fail.")
	}

	// Check if in CI
	if IsRunningInCI() {
		fmt.Println("Running in CI environment.")
		if IsExternalPR() && !HasAPIKeys() {
			fmt.Println("External PR detected without API keys available.")
			fmt.Println("Tests will be compiled but skipped at runtime.")
			return fmt.Errorf("Warning: Running in external PR without API keys. E2E tests will be skipped.")
		}
	}

	return nil
}
