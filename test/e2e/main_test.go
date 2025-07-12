package e2e

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	if err := framework.CheckRequirements(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Check environment and print any warnings
	if err := framework.CheckEnvironment(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}

	// Run tests
	exitCode := m.Run()

	// Clean up any leftover clusters unless cleanup is disabled
	cleanupClusters(exitCode != 0)

	os.Exit(exitCode)
}

// cleanupClusters attempts to clean up any leftover kind clusters from failed tests
func cleanupClusters(testFailed bool) {

	// If preserve-on-failure is set, warn about potential leftover clusters
	if framework.ShouldPreserveOnFailure() && testFailed {
		fmt.Println("Warning: preserve-on-failure is enabled. Clusters will be preserved.")
		fmt.Println("To see all clusters: kind get clusters")
		fmt.Println("To delete a specific cluster: kind delete cluster --name <cluster-name>")
		return // Skip cleanup if preserve-on-failure is set
	}

	prefix := "k8x-test-"

	fmt.Printf("Looking for clusters with prefix: %s\n", prefix)

	out, err := exec.Command("kind", "get", "clusters").Output()
	if err != nil {
		fmt.Printf("Failed to get clusters: %v\n", err)
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
			} else {
				fmt.Printf("Successfully deleted cluster: %s\n", cluster)
			}
		}
	}
}
