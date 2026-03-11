// Package worktree manages git worktrees for parallel agent development.
// Phase 1 implementation.
package worktree

// Worktree represents a git worktree.
type Worktree struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

// TerminalExecutor interface for creating terminal sessions.
// Will be implemented by terminal module in Phase 2.
type TerminalExecutor interface {
	CreateSession(id, cwd string) error
}

// Manager handles worktree operations.
type Manager struct {
	executor TerminalExecutor
}

// NewManager creates a new worktree manager.
func NewManager(executor TerminalExecutor) *Manager {
	return &Manager{executor: executor}
}
