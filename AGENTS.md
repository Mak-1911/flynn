# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` contains CLI entry points (`cmd/flynn/`, `cmd/flynnd/`).
- `internal/` holds private application code (agents, models, memory, tools, UI).
- `pkg/` exposes public libraries (shared protocol and client types).
- `api/` includes API definitions; `web/` contains frontend assets for the desktop app.
- `desktop/` is the Tauri desktop wrapper.
- `models/` stores local model configs (not weights).
- `scripts/` provides build and release helpers.
- `test/` contains integration and end-to-end fixtures (`test/e2e/`, `test/fixtures/`).
- `docs/` and `documentation/` provide user docs and architecture/PRD references.

## Build, Test, and Development Commands
- `make build`: Build the CLI for the current platform into `build/`.
- `make run`: Run the CLI via `go run ./cmd/flynn`.
- `make test`: Execute Go tests across the repo (`go test -v ./...`).
- `make test-coverage`: Generate `coverage.out` and `coverage.html`.
- `make fmt`: Format Go code with `go fmt ./...`.
- `make lint`: Run `golangci-lint` if installed.
- `make desktop`: Build the Tauri desktop app (requires Rust/Tauri tooling).

## Coding Style & Naming Conventions
- Use Go’s standard formatting (`make fmt`) and keep files gofmt-compliant.
- Naming follows Go package conventions and repository structure, e.g. `internal/model/router.go`.
- Test files use the `_test.go` suffix and live alongside the code they validate.

## Testing Guidelines
- Primary testing uses Go’s built-in toolchain (`go test`).
- Integration/E2E assets live under `test/`.
- When adding tests, name them descriptively and keep them close to the package under test.

## Commit & Pull Request Guidelines
- Recent commit history uses short, descriptive messages (e.g., “architectural changes”).
- Keep commits focused and prefer imperative summaries.
- PRs should include: a clear description of intent, linked issues (if any), and screenshots for UI changes in `web/` or `desktop/`.

## Configuration & Security
- Copy `.env.example` when adding env vars; document new keys.
- Local config lives in `config.toml`; runtime data goes under `~/.flynn/`.
- Avoid committing secrets or model weights.
