package framework

import (
	"os"
	"strings"
	"testing"
)

// IsRunningInCI returns true if the test is running in a CI environment
func IsRunningInCI() bool {
	return os.Getenv("CI") == "true" ||
		os.Getenv("GITHUB_ACTIONS") == "true"
}

// IsExternalPR returns true if the current build is from an external PR
// that doesn't have access to secrets
func IsExternalPR() bool {
	event := os.Getenv("GITHUB_EVENT_NAME")
	eventPath := os.Getenv("GITHUB_EVENT_PATH")

	// PR from fork (no access to secrets)
	return event == "pull_request" && eventPath != "" &&
		strings.Contains(os.Getenv("GITHUB_REPOSITORY"), "/") &&
		!strings.HasPrefix(os.Getenv("GITHUB_HEAD_REF"), "refs/heads/")
}

// HasAPIKeys returns true if the necessary API keys are available
func HasAPIKeys() bool {
	return os.Getenv("OPENAI_API_KEY") != ""
}

// ShouldSkipE2ETests returns true if E2E tests should be skipped
func ShouldSkipE2ETests() bool {
	// Skip E2E tests in short mode
	if testing.Short() {
		return true
	}

	// Skip if in external PR without API keys
	if IsRunningInCI() && IsExternalPR() && !HasAPIKeys() {
		return true
	}

	return false
}
