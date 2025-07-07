package e2e

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"k8x/test/e2e/framework"
)

// TestMain is used to set up shared resources for all E2E tests
func TestMain(m *testing.M) {
	// Parse the go-lang flags
	flag.Parse()

	if framework.ShouldSkipE2ETests() {
		fmt.Println("Short mode enabled; skipping all E2E tests.")
		os.Exit(0)
	}

	// Ensure required commands are available
	if err := checkRequirements(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Check environment and print any warnings
	checkEnvironment()

	// Run tests
	exitCode := m.Run()

	// Always clean up any leftover clusters
	cleanupClusters()

	os.Exit(exitCode)
}

// checkRequirements verifies that all required tools are installed
func checkRequirements() error {
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

// checkEnvironment checks the environment and prints warnings
func checkEnvironment() {
	// Check if Docker is running
	if err := exec.Command("docker", "info").Run(); err != nil {
		fmt.Println("Warning: Docker does not seem to be running. E2E tests will fail.")
	}

	// Check for API key
	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Println("Warning: OPENAI_API_KEY is not set. E2E tests will use a dummy API key and may fail.")
		fmt.Println("To run complete tests, set the environment variable:")
		fmt.Println("    export OPENAI_API_KEY=your-api-key")
	}

	// Check if in CI
	if framework.IsRunningInCI() {
		fmt.Println("Running in CI environment.")
		if framework.IsExternalPR() && !framework.HasAPIKeys() {
			fmt.Println("External PR detected without API keys available.")
			fmt.Println("Tests will be compiled but skipped at runtime.")
		}
	}
}

// cleanupClusters attempts to clean up any leftover kind clusters from failed tests
func cleanupClusters() {
	prefix := "k8x-test-"
	out, err := exec.Command("kind", "get", "clusters").Output()
	if err != nil {
		return // Ignore errors
	}

	clusters := strings.Split(string(out), "\n")
	for _, cluster := range clusters {
		cluster = strings.TrimSpace(cluster)
		if cluster == "" {
			continue
		}

		if strings.HasPrefix(cluster, prefix) {
			fmt.Printf("Cleaning up leftover cluster: %s\n", cluster)
			if err := exec.Command("kind", "delete", "cluster", "--name", cluster).Run(); err != nil {
				fmt.Printf("Warning: failed to delete cluster %s: %v\n", cluster, err)
			}
		}
	}
}
