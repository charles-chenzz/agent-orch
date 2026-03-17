// Package terminal manages PTY sessions and terminal emulation.
// Phase 2 implementation.
package terminal

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"time"
)

// SessionState 会话状态
type SessionState string

const (
	StateCreating  SessionState = "creating"
	StateRunning   SessionState = "running"
	StateDetached  SessionState = "detached"
	StateExited    SessionState = "exited"
	StateDestroyed SessionState = "destroyed"
)

// Session 终端会话
type Session struct {
	ID          string
	WorktreeID  string
	CWD         string
	State       SessionState
	PTY         *os.File
	Cmd         *exec.Cmd
	TmuxSession string // tmux session 名称
	CreatedAt   time.Time
	LastActive  time.Time
}

// Manager 终端会话管理器
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ctx      context.Context
	tmuxPath string // tmux 可执行文件路径
	hasTmux  bool   // tmux 是否可用
}

// TerminalConfig 终端配置
type TerminalConfig struct {
	Shell      string
	FontFamily string
	FontSize   int
	Theme      TerminalTheme
}

// TerminalTheme 终端主题
type TerminalTheme struct {
	Background string
	Foreground string
	Cursor     string
	Selection  string
}

// EventPayload 统一事件负载
type EventPayload struct {
	SessionID string `json:"sessionId"`
	Type      string `json:"type"` // output/state/error/exit
	Data      string `json:"data,omitempty"`
	State     string `json:"state,omitempty"`
	Error     string `json:"error,omitempty"`
	Timestamp int64  `json:"ts"`
}

// SessionInfo 会话信息（用于列表展示）
type SessionInfo struct {
	ID         string    `json:"id"`
	WorktreeID string    `json:"worktreeId"`
	CWD        string    `json:"cwd"`
	State      string    `json:"state"`
	CreatedAt  time.Time `json:"createdAt"`
	LastActive time.Time `json:"lastActive"`
}
