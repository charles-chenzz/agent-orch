// Package terminal manages PTY sessions and terminal emulation.
// Phase 2 implementation.
package terminal

// Session represents a terminal session.
type Session struct {
	ID      string `json:"id"`
	CWD     string `json:"cwd"`
	Running bool   `json:"running"`
}

// Manager handles terminal sessions.
type Manager struct {
	sessions map[string]*Session
}

// NewManager creates a new terminal manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}
