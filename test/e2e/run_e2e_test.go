package e2e

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"k8x/test/e2e/framework"
	"k8x/test/e2e/framework/scenarios"
)

func TestCrashLoopBackoffDiagnosis(t *testing.T) {

	// Create a unique cluster name
	clusterName := fmt.Sprintf("k8x-test-%d", time.Now().Unix())

	// Create the test cluster
	cluster, err := framework.NewClusterFixture(t, clusterName)
	if err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}
	defer cluster.Cleanup()

	// Apply the CrashLoop scenario
	if err := cluster.ApplyManifest(scenarios.CrashLoopBackoffScenario()); err != nil {
		t.Fatalf("Failed to apply CrashLoop scenario: %v", err)
	}

	// Wait for the pod to enter CrashLoopBackOff state
	if err := cluster.WaitForPodStatus("k8x-test", "crash-loop", "CrashLoopBackOff", 2*time.Minute); err != nil {
		t.Fatalf("Pod failed to reach CrashLoopBackOff state: %v", err)
	}

	// Run k8x against the cluster
	output, err := cluster.RunK8x("What's wrong with the crash-loop pod?")
	if err != nil {
		t.Logf("K8x output: %s", output)
		t.Fatalf("Failed to run k8x: %v", err)
	}

	// Verify that k8x correctly diagnosed the issue
	lowerOutput := strings.ToLower(output)
	if !strings.Contains(lowerOutput, "crash") && !strings.Contains(lowerOutput, "loop") {
		t.Errorf("K8x failed to diagnose CrashLoopBackOff. Output: %s", output)
	}

	if !strings.Contains(lowerOutput, "exit code 1") {
		t.Errorf("K8x failed to identify exit code. Output: %s", output)
	}
}

func TestImagePullBackoffDiagnosis(t *testing.T) {

	// Create a unique cluster name
	clusterName := fmt.Sprintf("k8x-test-%d", time.Now().Unix())

	// Create the test cluster
	cluster, err := framework.NewClusterFixture(t, clusterName)
	if err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}
	defer cluster.Cleanup()

	// Apply the ImagePullBackoff scenario
	if err := cluster.ApplyManifest(scenarios.ImagePullBackoffScenario()); err != nil {
		t.Fatalf("Failed to apply ImagePullBackoff scenario: %v", err)
	}

	// Wait for the pod to enter ImagePullBackOff state
	if err := cluster.WaitForPodStatus("k8x-test", "image-pull-error", "ImagePullBackOff", 2*time.Minute); err != nil {
		t.Fatalf("Pod failed to reach ImagePullBackOff state: %v", err)
	}

	// Run k8x against the cluster
	output, err := cluster.RunK8x("What's wrong with the image-pull-error pod?")
	if err != nil {
		t.Logf("K8x output: %s", output)
		t.Fatalf("Failed to run k8x: %v", err)
	}

	// Verify that k8x correctly diagnosed the issue
	if !strings.Contains(strings.ToLower(output), "imagepullbackoff") {
		t.Errorf("K8x failed to diagnose ImagePullBackOff. Output: %s", output)
	}

	if !strings.Contains(strings.ToLower(output), "image") && !strings.Contains(strings.ToLower(output), "pull") {
		t.Errorf("K8x failed to identify image pull issue. Output: %s", output)
	}
}

func TestMissingConfigMapDiagnosis(t *testing.T) {

	// Create a unique cluster name
	clusterName := fmt.Sprintf("k8x-test-%d", time.Now().Unix())

	// Create the test cluster
	cluster, err := framework.NewClusterFixture(t, clusterName)
	if err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}
	defer cluster.Cleanup()

	// Apply the ConfigMap missing scenario
	if err := cluster.ApplyManifest(scenarios.ConfigMapMissingScenario()); err != nil {
		t.Fatalf("Failed to apply ConfigMap missing scenario: %v", err)
	}

	// Wait for the pod to be created and show the volume mount error
	if err := cluster.WaitForPodStatus("k8x-test", "missing-config", "CreateContainerConfigError", 1*time.Minute); err != nil {
		// If CreateContainerConfigError is not detected, wait for pod to be created at least
		t.Logf("Pod may not have reached CreateContainerConfigError state, waiting briefly: %v", err)
		time.Sleep(5 * time.Second)
	}

	// Run k8x against the cluster
	output, err := cluster.RunK8x("Why isn't my missing-config-pod running?")
	if err != nil {
		t.Logf("K8x output: %s", output)
		t.Fatalf("Failed to run k8x: %v", err)
	}

	// Verify that k8x correctly diagnosed the issue
	if !strings.Contains(strings.ToLower(output), "configmap") {
		t.Errorf("K8x failed to diagnose missing ConfigMap. Output: %s", output)
	}

	if !strings.Contains(strings.ToLower(output), "non-existent-config") {
		t.Errorf("K8x failed to identify the missing ConfigMap name. Output: %s", output)
	}
}
