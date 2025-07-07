# End-to-End Tests for k8x

This directory contains end-to-end tests that validate k8x functionality against real Kubernetes clusters.

## Prerequisites

- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) installed
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed
- Go 1.19+
- Docker running

## Test Structure

The E2E tests are organized into:

- `framework/`: Framework for cluster creation and test utilities
- `framework/scenarios/`: Kubernetes resource manifests for problem scenarios
- `*_e2e_test.go`: Test cases for different diagnoses

## Running the Tests

You can run the E2E tests using:

```bash
make test-e2e
```

Or run them manually:

```bash
go test -v ./test/e2e/...
```

To run a specific test:

```bash
go test -v ./test/e2e/... -run TestCrashLoopBackoffDiagnosis
```

## Test Cases

1. **CrashLoopBackOff**: Tests k8x's ability to diagnose a pod in CrashLoopBackOff state due to an exit code 1
2. **ImagePullBackOff**: Tests diagnosing a pod that can't pull its container image
3. **Missing ConfigMap**: Tests diagnosing a pod that depends on a non-existent ConfigMap

## Adding New Test Cases

To add a new test scenario:

1. Create a new scenario function in `framework/scenarios/pod_failures.go`
2. Add a new test function in `run_e2e_test.go`
3. Ensure the test creates a unique kind cluster name to avoid conflicts with existing tests

## CI Integration

The E2E tests are integrated with GitHub Actions in the `.github/workflows/e2e-tests.yml` workflow file. They run:

- On pushes to the `main` branch
- On pull requests targeting the `main` branch
- On manual workflow dispatch

### API Key Management

These tests require an OpenAI API key to work properly. The API key is stored securely as a GitHub Actions secret named `OPENAI_API_KEY`.

For security reasons, secrets are not available to pull requests from forks. When a PR comes from a fork:

1. The tests will still be compiled and executed
2. They'll automatically detect they're running in a fork PR without access to API keys
3. They'll mark themselves as skipped with an appropriate message
4. After the PR is merged, the full E2E tests will run with proper API keys

### Debugging Test Runs

Test logs are automatically collected as GitHub Actions artifacts and are available for 7 days. This includes:

- Temporary kubeconfig files
- k8x configuration files
- kubectl output and errors

### Testing Locally

When running locally, you should set the `OPENAI_API_KEY` environment variable to your own API key:

```bash
export OPENAI_API_KEY=your-api-key-here
make test-e2e
```

Without a valid API key, the tests will run but likely fail during the actual k8x execution phase.
