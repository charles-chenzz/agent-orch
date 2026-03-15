# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Agent Orchestrator is a desktop application for managing multiple AI coding agents with git worktrees, terminal sessions, and API usage tracking. Built with Wails v2 (Go backend) + React/TypeScript frontend.

## Development Commands

```bash
# Development
make dev          # Start Wails dev server with hot reload
make build        # Build production binary

# Testing
make test         # Run all Go tests
go test -v ./internal/worktree/...              # Run specific package tests
go test -v -tags=integration ./internal/worktree/...  # Run integration tests (requires git)

# Code Quality
make lint         # Run golangci-lint (or go vet as fallback)
make fmt          # Format Go code

# Wails
make generate     # Regenerate Wails bindings after changing Go structs
```

## Architecture

### Frontend (React + TypeScript)
- **State**: Zustand stores (`frontend/src/stores/`)
- **Styling**: TailwindCSS with GitHub Dark Dimmed theme
- **IPC**: Wails auto-generated bindings in `frontend/wailsjs/go/main/`

### Backend (Go)
- **Entry**: `app.go` - Wails bindings expose methods to frontend
- **Modules**: Each `internal/` subdirectory is a self-contained module
  - `worktree/` - Git worktree CRUD (fully implemented)
  - `terminal/` - PTY session management (placeholder)
  - `proxy/` - API proxy for usage tracking (placeholder)
  - `config/` - Configuration management
  - `db/` - SQLite/GORM database

### Key Pattern: Hybrid Git Operations
The `worktree` module uses a hybrid approach due to go-git limitations:
- **Read operations**: Directly read `.git/worktrees/` directory files
- **Status checks**: Use `git status --porcelain` command (more reliable for linked worktrees)
- **Write operations**: Use `git worktree add/remove` commands

## Current Implementation Status

| Phase | Status | Description |
|-------|--------|-------------|
| 0 | ✅ Complete | Foundation, UI layout, config system |
| 1 | 🔄 In Progress | Worktree CRUD (backend done, frontend pending) |
| 2-6 | 📋 Planned | Terminal, API proxy, Agent monitor, Editor |

## Important Conventions

### Go Error Handling
The worktree module uses structured errors with codes:
```go
type WorktreeError struct {
    Code    string `json:"code"`    // e.g., "ERR_NAME_INVALID"
    Message string `json:"message"`
}
```
Error codes are defined in `internal/worktree/errors.go`.

### Frontend-Backend Communication
1. Define method in `app.go`
2. Run `make generate` to create TypeScript bindings
3. Import from `frontend/wailsjs/go/main/App`

### Testing
- Unit tests use `github.com/stretchr/testify`
- Integration tests require `//go:build integration` tag
- Test repos are created with `t.TempDir()` for isolation
