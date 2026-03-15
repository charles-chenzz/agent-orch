# Repository Guidelines

## Project Structure & Module Organization
- `app.go` and `main.go` are the Wails entry points.
- `internal/` contains Go modules: `config/`, `db/`, `worktree/`, `terminal/`, `proxy/`, `agent/`, `github/`.
- `frontend/` holds the React + Vite UI; shared layout lives in `frontend/src/components/Layout/`.
- `docs/` tracks phase plans and design decisions; `scripts/` contains setup helpers.
- `build/` is for compiled artifacts; `wails.json` stores app configuration.

## Build, Test, and Development Commands
- `nix develop` (recommended) or `./scripts/setup.sh` for manual setup.
- `make dev` runs `wails dev` with hot reload.
- `make build` builds the production binary.
- `make test` runs Go tests.
- `make lint` runs `golangci-lint` (or `go vet`) and `npm run build` in `frontend/`.
- `make install-deps` downloads Go modules and frontend npm dependencies.
- `make generate` regenerates Wails bindings.

## Coding Style & Naming Conventions
- Go: format with `gofmt` (tabs), keep package names short and lowercase.
- TypeScript/React: follow existing file style (no semicolons in current files), use `PascalCase` for components and `useX` for hooks.
- Prefer `camelCase` for variables/functions and descriptive names for Zustand stores (for example `appStore`).
- Tailwind-first styling; keep layout primitives under `frontend/src/components/Layout/`.

## Testing Guidelines
- Go tests live in `internal/**/**/*_test.go`; run with `make test` or `go test ./...`.
- There is no frontend test harness yet; avoid adding JS tests unless you introduce a framework.

## Commit & Pull Request Guidelines
- Recent history follows conventional-style prefixes like `feat(scope): ...`, `docs: ...`, `chore: ...`; match that pattern.
- PRs should include a concise summary, test commands run (or “not run”), and UI screenshots/GIFs for visual changes.

## Configuration & Security
- Do not commit API keys or tokens. Document any new configuration in `docs/`.
