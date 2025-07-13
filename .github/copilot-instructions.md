# Copilot Coding Instructions

Use the Cobra framework for CLI commands with descriptive names and POSIX-compliant flag patterns.

Write Go code with explicit error handling using wrapped errors like `fmt.Errorf("operation failed: %w", err)`.

Implement interfaces for external dependencies like LLM clients and file operations to enable testing with mocks.

Follow Go naming conventions: kebab-case for files, snake_case for packages, camelCase for variables.

Store all configuration under `~/.appname/` directory structure with logical subdirectories.

Write unit tests for public functions using table-driven test patterns with Go's standard testing package.

Keep functions small and single-purpose, extracting common patterns into reusable utilities.

Use structured logging with consistent field names, preferring logrus or zap over standard log package.

Handle file operations with proper error checking for missing directories, permissions, and disk space issues.

Use composition and dependency injection patterns for modular, testable code.

Organize code using standard Go project layout with `cmd/` for CLI commands (one file per command), `internal/` for private packages, and domain-specific packages like `config`, `history`, and `llm`.

Place provider implementations in sub-packages under their domain, such as `internal/llm/providers/` for different LLM client implementations.

Co-locate test files with source code using the `_test.go` suffix in the same package directory.

Keep main.go minimal as an entry point that delegates to cmd package, and place build artifacts in a `build/` directory.

Structure documentation in `docs/` and working examples in `examples/` with realistic configuration files and usage scenarios.

Follow this exact project structure:

```text
.
├── build
│   └── k8x                          # Compiled binary output
├── cmd
│   ├── config.go                    # Config management utilities
│   ├── configure.go                 # `k8x configure` command implementation
│   ├── history.go                   # `k8x history` command implementation
│   ├── root.go                      # Root cobra command and global flags
│   ├── run.go                       # `k8x run` command implementation
│   └── version.go                   # `k8x version` command implementation
├── internal
│   ├── config
│   │   ├── config.go                # Configuration struct and file operations
│   │   ├── config_test.go           # Tests for config operations
│   │   └── credentials.go           # LLM API key management
│   ├── history
│   │   └── manager.go               # .k8x file creation and parsing
│   ├── llm
│   │   ├── client.go                # LLM client interface definition
│   │   ├── client_test.go           # Tests for LLM client interface
│   │   ├── factory.go               # LLM client factory and provider selection
│   │   ├── tools.go                 # LLM tool calling and prompt utilities
│   │   ├── tools_test.go            # Tests for LLM tools
│   │   └── providers
│   │       ├── anthropic.go         # Claude/Anthropic client implementation
│   │       ├── openai.go            # OpenAI GPT client implementation
│   │       └── unified.go           # Common provider utilities
│   └── schemas
│       └── credentials.go           # Credential file format definitions
├── docs
│   ├── README.md                    # Detailed project documentation
│   └── shell-execution-tool.md      # .k8x format specification
├── examples
│   ├── basic-usage.md               # Usage examples and tutorials
│   ├── config.yaml                  # Sample configuration file
│   └── credentials                  # Sample credentials file
├── main.go                          # Application entry point
└── Makefile                         # Build and development tasks
```
