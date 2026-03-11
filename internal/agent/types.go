// Package agent detects and monitors AI coding agents.
// Phase 5 implementation.
package agent

// Agent represents a detected AI coding agent.
type Agent struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	PID      int    `json:"pid"`
	Worktree string `json:"worktree"`
}

// Detector finds running AI agents.
type Detector struct {
	agents map[string]*Agent
}

// NewDetector creates a new agent detector.
func NewDetector() *Detector {
	return &Detector{
		agents: make(map[string]*Agent),
	}
}
