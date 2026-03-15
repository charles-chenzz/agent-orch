// Package worktree manages git worktrees for parallel agent development.
// Phase 1 implementation.
package worktree

import "time"

// Worktree represents a git worktree.
type Worktree struct {
	ID           string    `json:"id"`           // 唯一标识 (目录名或 "main")
	Name         string    `json:"name"`         // 显示名称
	Path         string    `json:"path"`         // 文件系统路径
	Branch       string    `json:"branch"`       // 当前分支
	Head         string    `json:"head"`         // 当前 commit hash (前7位)
	IsMain       bool      `json:"isMain"`       // 是否为主 worktree
	HasChanges   bool      `json:"hasChanges"`   // 是否有未提交变更
	Unpushed     int       `json:"unpushed"`     // 未推送提交数
	LastActivity time.Time `json:"lastActivity"` // 最后活动时间
}

// WorktreeStatus 包含 worktree 的详细状态信息.
type WorktreeStatus struct {
	WorktreeID string       `json:"worktreeId"`
	Branch     string       `json:"branch"`
	Head       string       `json:"head"`
	Ahead      int          `json:"ahead"`     // 领先远程的提交数
	Behind     int          `json:"behind"`    // 落后远程的提交数
	Staged     []FileStat   `json:"staged"`    // 已暂存文件
	Unstaged   []FileStat   `json:"unstaged"`  // 未暂存文件
	Untracked  []string     `json:"untracked"` // 未跟踪文件
	LastCommit *CommitInfo  `json:"lastCommit"`
}

// FileStat 表示文件变更状态.
type FileStat struct {
	Path   string `json:"path"`
	Status string `json:"status"` // M, A, D, R
}

// CommitInfo 包含提交信息.
type CommitInfo struct {
	Hash    string    `json:"hash"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Time    time.Time `json:"time"`
}

// TerminalExecutor interface for creating terminal sessions.
// Will be implemented by terminal module in Phase 2.
type TerminalExecutor interface {
	CreateSession(id, cwd string) error
}
