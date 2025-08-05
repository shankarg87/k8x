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
- `*_test.go`: Test cases for different diagnoses

## Running the Tests

You can run the E2E tests using:

```bash
make test-e2e-all
```

Or run them manually:

```bash
go test -v ./test/e2e/...
```

To run a specific test:

```bash
make test-e2e-single TEST=TestCrashLoopBackoffDiagnosis
```

## Test Cases

1. **CrashLoopBackOff**: Diagnoses a pod in CrashLoopBackOff state due to an exit code 1
2. **ImagePullBackOff**: Diagnoses a pod that cannot pull its container image
3. **Missing ConfigMap**: Diagnoses a pod that depends on a non-existent ConfigMap
4. **Kubectl Exit Code Validation**: Validates that k8x properly analyzes resource scenarios and handles kubectl command failures

## API Key Management

Without a valid API key, tests will compile and run but may be skipped or fail during k8x execution.
**Before running tests, please export your LLM API keys.**

## Adding New Test Cases

To add a new test scenario:

1. Create a new scenario function in `framework/scenarios/pod_failures.go` or `framework/scenarios/opa_gatekeeper.go`
2. Add a new test function in `run_e2e_test.go`
3. Ensure each test uses a unique kind cluster name to avoid conflicts

## CI Integration

The E2E tests are integrated with GitHub Actions as defined in `.github/workflows/test-e2e.yml`. They run:

- On pushes to the `main` branch
- On pull requests targeting the `main` branch
- On manual workflow dispatch

## Debugging Test Runs

Test logs are automatically collected as GitHub Actions artifacts and are available for 7 days. This includes:

- Temporary kubeconfig files
- k8x configuration files (in ~/.k8x/)
- kubectl output and errors

## MCP Integration Tests

The MCP integration test (`mcp_integration_test.go`) validates that k8x can successfully communicate with external MCP servers, currently using the DuckDuckGo MCP server.

### Prerequisites for MCP Integration Tests

- Python 3.x installed
- `pip` available
- Internet access for dependency downloads

### Running MCP Integration Tests

```bash
# Run MCP integration tests
make test-e2e-single TEST=TestDuckDuckGoMCPIntegration
```

### What the MCP Integration Tests Validate

1. **Client Creation**: Verifies MCP client creation using stdio transport
2. **Tool Discovery**: Connects to the DuckDuckGo MCP server and lists available tools
3. **Search Functionality**: Executes search queries and validates results
4. **Error Handling**: Tests behavior with invalid tool calls and parameters

The integration test process automatically:

- Installs the `uv` Python package manager if missing
- Downloads and installs the `duckduckgo-mcp-server` package
- Launches the MCP server as a subprocess, performs operations, and cleans up afterwards

## Skipping e2e Tests

Integration tests are automatically skipped if:

- Running in `-short` mode
- Python dependencies are not available
- Network connectivity issues prevent server installation
