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

	// Apply the CrashLoop scenario
	if err := cluster.ApplyManifest(scenarios.CrashLoopBackoffScenario()); err != nil {
		t.Fatalf("Failed to apply CrashLoop scenario: %v", err)
	}

	// Wait for the pod to enter CrashLoopBackOff state
	if err := cluster.WaitForPodStatus("k8x-test", "crash-loop", "CrashLoopBackOff", 2*time.Minute); err != nil {
		t.Fatalf("Pod failed to reach CrashLoopBackOff state: %v", err)
	}

	// Run k8x against the cluster
	output, err := cluster.RunK8x("Is the crash-loop running? Answer only Yes or No.")
	if err != nil {
		t.Logf("K8x output: %s", output)
		t.Fatalf("Failed to run k8x: %v", err)
	}

	// Verify that k8x correctly diagnosed the issue
	lowerOutput := strings.ToLower(output)
	if !strings.Contains(lowerOutput, "no") {
		t.Errorf("K8x failed to diagnose CrashLoopBackOff. Output: %s", output)
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

func TestOPAGatekeeperHostPortDeny(t *testing.T) {
	// Create a unique cluster name
	clusterName := fmt.Sprintf("k8x-test-%d", time.Now().Unix())

	// Create the test cluster
	cluster, err := framework.NewClusterFixture(t, clusterName)
	if err != nil {
		t.Fatalf("Failed to create test cluster: %v", err)
	}

	// Step 1: Install Gatekeeper using official Helm chart
	t.Log("Installing Gatekeeper using Helm...")
	releaseName := fmt.Sprintf("gatekeeper-%d", time.Now().Unix())
	if err := cluster.ExecuteCommands(scenarios.GatekeeperInstallationCommands(releaseName)); err != nil {
		t.Fatalf("Failed to install Gatekeeper: %v", err)
	}

	// Wait for Gatekeeper to be ready
	t.Log("Waiting for Gatekeeper deployment to be ready...")
	if err := cluster.WaitForDeployment("gatekeeper-system", "gatekeeper-controller-manager", 5*time.Minute); err != nil {
		t.Logf("Warning: Gatekeeper deployment may not be fully ready: %v", err)
		// Continue anyway - we'll test the policy enforcement
	}

	// Wait for ConstraintTemplate CRD to be established
	t.Log("Waiting for ConstraintTemplate CRD to be established...")
	if err := cluster.WaitForCRD("constrainttemplates.templates.gatekeeper.sh", 2*time.Minute); err != nil {
		t.Fatalf("ConstraintTemplate CRD not ready: %v", err)
	}

	// Step 2: Apply the ConstraintTemplate
	t.Log("Installing HostPort ConstraintTemplate...")
	if err := cluster.ApplyManifest(scenarios.HostPortConstraintTemplateScenario()); err != nil {
		t.Fatalf("Failed to apply ConstraintTemplate: %v", err)
	}

	// Wait for ConstraintTemplate to be established
	if err := cluster.WaitForConstraintTemplate("k8xhostportforbidden", 2*time.Minute); err != nil {
		t.Logf("Warning: ConstraintTemplate may not be fully ready: %v", err)
	}

	// Wait for the constraint CRD to be created by Gatekeeper
	t.Log("Waiting for K8xHostPortForbidden CRD to be established...")
	if err := cluster.WaitForCRD("k8xhostportforbidden.constraints.gatekeeper.sh", 5*time.Minute); err != nil {
		t.Fatalf("K8xHostPortForbidden CRD not ready: %v", err)
	}

	// Step 3: Apply the Constraint
	t.Log("Installing HostPort Constraint...")
	if err := cluster.ApplyManifest(scenarios.HostPortConstraintScenario()); err != nil {
		t.Fatalf("Failed to apply Constraint: %v", err)
	}

	// Wait for constraint to be ready
	if err := cluster.WaitForConstraint("K8xHostPortForbidden", "hostport-not-allowed", 1*time.Minute); err != nil {
		t.Logf("Warning: Constraint may not be fully ready: %v", err)
	}

	// Give some time for the admission controller to be ready
	t.Log("Waiting for admission controller to be ready...")
	time.Sleep(30 * time.Second)

	// Step 4: Test that rejected deployment fails
	t.Log("Testing forbidden deployment (with hostPort)...")
	if err := cluster.ApplyManifest(scenarios.ForbiddenDeploymentScenario()); err == nil {
		t.Errorf("Forbidden deployment was allowed when it should have been rejected")
	} else {
		t.Logf("✓ Forbidden deployment was correctly rejected: %v", err)
	}

	// Step 5: Run k8x to diagnose why the deployment was rejected
	t.Log("Running k8x to diagnose the deployment rejection...")
	output, err := cluster.RunK8x("Why was my most recent deployment rejected? Getting some weird errors.")
	if err != nil {
		t.Logf("K8x output: %s", output)
		t.Fatalf("Failed to run k8x: %v", err)
	}

	// Step 6: Verify that k8x correctly diagnosed the OPA Gatekeeper policy violation
	lowerOutput := strings.ToLower(output)

	// Check for OPA/Gatekeeper related terms
	if !strings.Contains(lowerOutput, "gatekeeper") && !strings.Contains(lowerOutput, "opa") && !strings.Contains(lowerOutput, "policy") {
		t.Errorf("K8x failed to identify OPA Gatekeeper policy violation. Output: %s", output)
	}

	// Check for hostPort specific violation
	if !strings.Contains(lowerOutput, "hostport") {
		t.Errorf("K8x failed to identify hostPort restriction. Output: %s", output)
	}

	// Check for constraint/admission controller related terms
	if !strings.Contains(lowerOutput, "constraint") && !strings.Contains(lowerOutput, "admission") && !strings.Contains(lowerOutput, "forbidden") {
		t.Errorf("K8x failed to identify admission controller constraint. Output: %s", output)
	}

	t.Logf("✓ K8x successfully diagnosed the OPA Gatekeeper policy violation")

	// Cleanup Gatekeeper installation
	if err := cluster.ExecuteCommands(scenarios.GatekeeperUninstallCommands(releaseName)); err != nil {
		t.Logf("Warning: Failed to cleanup Gatekeeper: %v", err)
	}
}
