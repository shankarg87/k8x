# Testing k8x

This document covers the testing strategy for k8x, including unit tests, integration tests, and end-to-end tests.

## Test Structure

The k8x project uses multiple layers of testing:

1. **Unit Tests**: Located alongside the code in `*_test.go` files
2. **End-to-End Tests**: Test the complete CLI against real Kubernetes clusters

## Running Tests

### Unit Tests

Run all unit tests with:

```bash
make test
```

### End-to-End Tests

E2E tests require additional setup as they create real kind clusters and test actual k8x functionality.

Prerequisites:

- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) installed
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed
- Docker running
- Valid OpenAI API key (set as `OPENAI_API_KEY` environment variable)

Run E2E tests with:

```bash
# Set your API key
export OPENAI_API_KEY=your-api-key-here

# Run all E2E tests
make test-e2e

# Run specific E2E test
make test-e2e-single TEST=TestCrashLoopBackoffDiagnosis

```

## CI Integration

Both unit tests and E2E tests are integrated into our GitHub Actions workflows:

- **Unit Tests**: Run on all PRs and pushes to main
- **E2E Tests**: Run on PRs from the main repository and pushes to main

For security reasons, E2E tests in PRs from forks will be compiled but skip actual execution since they don't have access to API keys.

## Adding Tests

### Unit Tests

Follow Go's standard testing practices:

- Place tests in the same package as the code being tested
- Use file naming pattern `<filename>_test.go`
- Use table-driven tests where appropriate

### E2E Tests

Add new E2E test scenarios in:

1. Create scenario in `test/e2e/framework/scenarios/`
2. Add test case in `test/e2e/run_e2e_test.go`

See [test/e2e/README.md](test/e2e/README.md) for detailed instructions on E2E testing.
