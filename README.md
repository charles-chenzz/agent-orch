# Agent Orchestrator

A desktop application for managing multiple AI coding agents with git worktrees, terminal sessions, and API usage tracking.

## Quick Start

### Option 1: Nix (Recommended)

```bash
# Enter development environment
nix develop

# Start development server
make dev
```

### Option 2: Manual Setup

```bash
# Run setup script
./scripts/setup.sh

# Start development server
make dev
```

## Requirements

- Go 1.22+
- Node.js 20+
- Wails CLI v2.10+

### Linux Additional Requirements

- GTK 3 development libraries
- WebKitGTK development libraries

## Development

```bash
make dev      # Start development server
make build    # Build production binary
make test     # Run tests
make lint     # Run linters
make help     # Show all commands
```

## Project Structure

```
agent-orch/
├── frontend/          # React + Vite frontend
├── internal/          # Go business logic
│   ├── config/       # Configuration management
│   ├── db/           # Database (SQLite)
│   ├── worktree/     # Git worktree management
│   ├── terminal/     # Terminal sessions
│   ├── proxy/        # API proxy server
│   ├── agent/        # Agent detection
│   └── github/       # GitHub integration
├── flake.nix         # Nix development environment
├── Makefile          # Build commands
└── scripts/setup.sh  # Manual setup script
```

## License

MIT
