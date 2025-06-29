# k8x

**k8x** is a Kubernetes shell‐workflow assistant that marries natural‐language goals with step‐by‐step execution, logging, and reversible changes—specialized for Kubernetes operations. Under the hood, it uses your configured LLM to plan, run, document, and (when enabled) undo `kubectl` commands—all in a single annotated `.k* Ground Copilot's responses in our LLM integration design: always implement a pluggable Go client interface for providers (OpenAI, Anthropic, etc.), loading credentials from `~/.k8x/credentials`.x` file.

> First Iteration: Full READ-ONLY kubectl access, including inside pods. No cluster modifications are performed until write-mode is implemented in later iterations.
> 

## 📑 Detailed **`.k8x`** Specification (v0.1)

### 1. Goals & Philosophy

- **Kubernetes‑First Audit:** Full trace of “intent → `kubectl` command → output → undo (future).”
- **Superset of POSIX Shell:** Strip comments to get a pure `.sh` or `.kube.sh` script.
- **Reversible & Exploratory:** Record diagnostics (`#~`) and allow safe roll‑back in write‑mode.

### 2. File Structure

```
~/.k8x/history/<goal>-<YYYYMMDD>-<HHMMSS>.k8x

```

Each `.k8x` is UTF-8, line-oriented:

1. **Goal Declaration**
    
    ```bash
    #$ <free-form Kubernetes goal prompt>
    
    ```
    
2. **k8x Comments**
    
    ```bash
    ## <notes, TODOs, tips>
    
    ```
    
3. **Step Records** (repeated):
    
    ```ebnf
    step-block ::= step-annotation command-block output-block* undo-block?
    
    ```
    

### 3. Line‐Type Reference

| Prefix | Meaning |
| --- | --- |
| `#!` | shebang (`/bin/k8x`) |
| `#$` | **Goal**: Top-level Kubernetes task description |
| `##` | **Comment**: Human-only notes (ignored by engine) |
| `#` | **Plan Step**: Numbered/named action |
| `#~` | **Exploratory**: LLM-tried diagnostics or side experiments |
| `#?` | **Agent Question**: k8x asks *you* for input (e.g. RBAC elevation, kubeconfig context) |
| *(none)* | **Shell Command**: Runs in a clean subshell (`kubectl ...`) |
| `#>` | **Output**: Captured stdout/stderr |
| `#-` | **Undo Command**: Reverse of preceding command (future write-mode) |

### 4. Grammar (EBNF‐style)

```ebnf
k8x-file       ::= goal-block comment-block* step-block*
goal-block     ::= "#$" SP text EOL
comment-block  ::= "##" SP text EOL

step-block     ::= step-annotation command-block output-block* undo-block?
step-annotation::= ("#"  SP text EOL)
                  |("#~" SP text EOL)
                  |("#?" SP text EOL)
command-block  ::= (shell-line EOL)+
output-block   ::= "#>" SP text EOL
undo-block     ::= "#-" SP text EOL

```

- **text** = any characters except newline
- **shell-line** = any non-comment shell code (typically `kubectl ...`)

### 5. Supported CLI Commands

| Command | Description |
| --- | --- |
| `k8x configure` | Initialize workspace & pick LLM model, create `~/.k8x/credentials` |
| `k8x run "<goal>"` | New session: kebab-case goal + timestamp → `~/.k8x/history/...` file, start LLM-driven loop |
| `k8x history` | List all past sessions with goals and timestamps |
| `k8x undo [-n N]` | (Future) Roll back one session’s changes with stored undo commands (default `N=1`) |
| `k8x undo all` | (Future) Roll back every session’s changes in order |

### 6. Engine Workflow

1. **Initialization**
    - `run`: create `~/.k8x/history/<goal>-<ts>.k8x` with `#$` line.
2. **Planning Loop**
    1. Parse existing annotations & commands.
    2. Prompt LLM: include goal, comments, and prior steps.
    3. Append:
        
        ```bash
        # N. <step description>
        kubectl <...>
        #> <captured stdout/stderr>
        #- <undo command>       # write-mode only
        
        ```
        
    4. Repeat until LLM signals DONE.
3. **Post-Processing**
    - (Future flags) Simplify trace, drop `#~`, inline `#?` resolutions, export pure scripts.

### 7. Best Practices

- **Atomic Kubernetes Actions:** One logical `kubectl` invocation per `#` block.
- **Explorations:** Mark risky trials with `#~` so they can be pruned.
- **Agent Prompts:** Use `#?` for RBAC/context questions (e.g. switch namespace).
- **Descriptive Goals:** Keep the goal line concise yet actionable.

---

## Integration: LLM Providers & Go SDKs

### Anthropic Claude Sonnet 4

Leverage the [Anthropic Claude Sonnet 4 Go SDK](https://github.com/anthropics/anthropic-sdk-go) to power LLM-driven step planning.

```go
package main

import (
    "context"
    "fmt"

    anthropic "github.com/anthropics/anthropic-sdk-go"
)

func main() {
    apiKey := "YOUR_CLAUDE_API_KEY"
    client := anthropic.NewClient(apiKey)

    prompt := anthropic.PromptBuilder().
        WithSystem("You are k8x, a Kubernetes shell assistant specialized in read-only diagnostics.").
        WithUser("List all pods in the default namespace").
        Build()

    resp, err := client.Complete(
        context.Background(),
        prompt,
        anthropic.WithModel("claude-sonnet-4"),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Completion)
}

```

> Tip: Combine multiple calls for complex diagnostics (e.g., list pods → describe pods → check events).
> 

### OpenAI Go SDK

Use the [OpenAI Go SDK](https://github.com/openai/openai-go) to generate chat-driven steps and outputs.

```go
package main

import (
    "context"
    "fmt"
    "os"

    openai "github.com/openai/openai-go"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    client := openai.NewClient(apiKey)

    req := openai.ChatCompletionRequest{
        Model: openai.GPT4,
        Messages: []openai.ChatCompletionMessage{
            {Role: "system", Content: "You are k8x, a Kubernetes diagnostics assistant."},
            {Role: "user", Content: "List all pods in the default namespace"},
        },
    }
    resp, err := client.Chat.CreateCompletion(context.Background(), req)
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Choices[0].Message.Content)
}

```

> Tip: Swap Model for other supported variants (e.g., GPT_3_5_TURBO).
> 

### Google Gemini 2.5 Go SDK

Integrate Google’s Gemini 2.5 model using the Google Gen AI Go SDK.

```go
package main

import (
    "context"
    "fmt"
    "os"

    genai "cloud.google.com/go/generative-ai/apiv1beta2"
    genaipb "google.golang.org/genproto/googleapis/ai/generativelanguage/v1beta2"
)

func main() {
    ctx := context.Background()
    client, err := genai.NewTextClient(ctx)
    if err != nil {
        panic(err)
    }
    defer client.Close()

    req := &genaipb.GenerateTextRequest{
        Model: "models/gemini-2.5-pro",
        Prompt: &genaipb.TextPrompt{Text: "List all pods in the default namespace"},
        Temperature: 0.2,
    }
    resp, err := client.GenerateText(ctx, req)
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.GetText())
}

```

> Tip: Ensure GOOGLE_APPLICATION_CREDENTIALS points to your service account JSON, and switch Model to e.g. gemini-2.5-flash for lower latency builds.
> 

---

## Implementation Roadmap

### Iteration 0: CLI Scaffold & Configuration

- Scaffold Cobra app **`k8x`** with subcommands:
    - `k8x configure`
    - `k8x run`
    - `k8x undo`, `k8x undo all`
- Ensure `~/.k8x/` exists on install.
- Implement `k8x configure` interactive model + credentials.

### Iteration 1: Session File Creation

- `k8x run "<goal>"` creates history file with only the goal line.

### Iteration 2: Read-Only Execution Loop

- Wire Claude Sonnet 4 client.
- Prompt LLM for next `kubectl` command.
- Execute in subshell; capture output; append trace.

### Iteration 3: Undo Support & Flags (Future)

- Add inverse-command map for supported `kubectl` ops.
- Append `#- <undo>` after each step.
- Implement `k8x undo` and `k8x undo all`.

### Iteration 4: Exploratory & Interactive

- Recognize `#~` for experimental probes.
- Support `#?` to pause and ask the user for context or RBAC elevation.

---

## 🚀 Developer Experience with **k8x**

### 1. Install & Configure

```bash
# Install via Homebrew
brew install k8x

# Initialize workspace & pick your model
k8x configure

```

- **Creates** `~/.k8x/` with:
    - `credentials` storing your LLM endpoint & API key
    - `history/` folder for session files

### 2. Run a New Session (Read-Only Diagnostics)

```bash
k8x run "Diagnose why service X has no endpoints"

```

- **Generates**:
    
    ```bash
    ~/.k8x/history/diagnose-service-x-20250626-142530.k8x
    #$ Diagnose why service X has no endpoints
    # 1. List pods in default namespace
    kubectl get pods
    #> …
    # 2. Describe svc X
    kubectl describe svc X
    #> …
    
    ```
    
- **Loop:** Plans, runs, logs—until DONE.

### 3. Undo Changes (Future Write-Mode)

```bash
k8x undo         # roll back last session
k8x undo -n 3    # roll back 3rd-most-recent session
k8x undo all     # roll back every session

```

Happy Kubernetes automating—with full visibility and reversible workflows!

## Repository tech stack and conventions

* When generating CLI code, use the Cobra framework: define commands and subcommands via `github.com/spf13/cobra`, and manage flags with POSIX-compliant `pflag` so the tool remains a static, single-binary executable([pkg.go.dev][3]).
* For release automation snippets, produce a GoReleaser config and GitHub Actions workflow (`.github/workflows/release.yml`) using the official `goreleaser/goreleaser-action` to build cross-platform binaries, Homebrew formulae, `.deb`, RPM, and Snap packages([goreleaser.com][4]).
* When suggesting Docker image builds, include multi-arch support via `docker/setup-qemu-action` and configure `docker_manifests` in `.goreleaser.yaml` to publish behind unified tags([goreleaser.com][5]).
* Ground Copilot’s responses in our LLM integration design: always implement a pluggable Go client interface for providers (OpenAI, Anthropic, etc.), loading credentials from `~/.shx/credentials`.
* Follow our configuration & history conventions: read/write all state under `~/.k8x/` and `~/.k8x/history/`, naming files `<goal>-<timestamp>.k8x` (kebab-case) for consistency.
* For undo functionality, reference our inverse-command map: each CLI action should have a defined “undo” pair; if none exists, prompt the user before proceeding.
* Write tests in Go using the standard `testing` package for unit tests, and mock shell or Docker-based containers for lightweight integration tests([gianarb.it][6]).
* Auto-generate documentation: place prose in `/docs`, examples in `/examples`, and invoke Cobra’s `GenManTree` or similar to produce man pages alongside the binary([docs.github.com][7]).
* Ensure all code snippets and CI configs adhere to our style: short, self-contained Markdown statements, and avoid external dependencies beyond what’s specified here([docs.github.com][8]).
* Align Copilot’s pull-request and code-review suggestions with our best practices: focus on readability, security checks, and conformity to this repository’s conventions([docs.github.com][9]).
* When proposing GoReleaser features (e.g., cross-language support or SBOM generation), reference recent GoReleaser capabilities like changelog automation, package signing, and OCI image publishing([goreleaser.com][10]).

---

References for tech stack:


[3]: https://pkg.go.dev/github.com/spf13/cobra?utm_source=chatgpt.com "cobra package - github.com/spf13/cobra - Go Packages"
[4]: https://goreleaser.com/ci/actions/?utm_source=chatgpt.com "GitHub Actions - GoReleaser"
[5]: https://goreleaser.com/cookbooks/multi-platform-docker-images/?utm_source=chatgpt.com "Multi-platform Docker images - GoReleaser"
[6]: https://gianarb.it/blog/golang-mockmania-cli-command-with-cobra?utm_source=chatgpt.com "How to test CLI commands made with Go and Cobra"
[7]: https://docs.github.com/copilot/using-github-copilot/code-review/using-copilot-code-review?utm_source=chatgpt.com "Using GitHub Copilot code review"
[8]: https://docs.github.com/en/copilot/customizing-copilot?utm_source=chatgpt.com "Customizing Copilot - GitHub Docs"
[9]: https://docs.github.com/en/copilot/using-github-copilot/coding-agent/best-practices-for-using-copilot-to-work-on-tasks?utm_source=chatgpt.com "Best practices for using Copilot to work on tasks - GitHub Docs"
[10]: https://goreleaser.com/blog/rust-zig/?utm_source=chatgpt.com "Using GoReleaser and GitHub Actions to release Rust and Zig ..."
