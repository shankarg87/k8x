
## Repository tech stack and conventions

* When generating CLI code, use the Cobra framework: define commands and subcommands via `github.com/spf13/cobra`, and manage flags with POSIX-compliant `pflag` so the tool remains a static, single-binary executable([pkg.go.dev][3]).
* For release automation snippets, produce a GoReleaser config and GitHub Actions workflow (`.github/workflows/release.yml`) using the official `goreleaser/goreleaser-action` to build cross-platform binaries, Homebrew formulae, `.deb`, RPM, and Snap packages([goreleaser.com][4]).
* When suggesting Docker image builds, include multi-arch support via `docker/setup-qemu-action` and configure `docker_manifests` in `.goreleaser.yaml` to publish behind unified tags([goreleaser.com][5]).
* Ground Copilot’s responses in our LLM integration design: always implement a pluggable Go client interface for providers (OpenAI, Anthropic, etc.), loading credentials from `~/.shx/credentials`.
* Follow our configuration & history conventions: read/write all state under `~/.shx/` and `~/.shx/history/`, naming files `<goal>-<timestamp>.json` (kebab-case) for consistency.
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
