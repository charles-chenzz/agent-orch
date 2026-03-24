package main

import (
	"context"
	"fmt"

	"agent-orch/internal/config"
	"agent-orch/internal/db"
	"agent-orch/internal/terminal"
	"agent-orch/internal/worktree"
)

// App struct
type App struct {
	ctx       context.Context
	worktree  *worktree.Manager
	terminal  *terminal.Manager
	db        *db.Database
	repoPath  string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// TODO: 从配置获取仓库路径，暂时使用当前目录
	a.repoPath = "."

	// 初始化配置和数据库
	cfg, _ := config.Load()
	if cfg != nil {
		if database, err := db.Init(cfg.DatabasePath()); err == nil {
			a.db = database
		}
	}

	// 初始化终端管理器（传入 db 用于会话持久化）
	a.terminal = terminal.NewManager(ctx, a.db)

	// 初始化 worktree 管理器（使用当前目录）
	if mgr, err := worktree.NewManager("."); err == nil {
		a.worktree = mgr
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// === Worktree API ===

// SetRepoPath 设置当前管理的仓库路径
func (a *App) SetRepoPath(path string) error {
	mgr, err := worktree.NewManager(path)
	if err != nil {
		return err
	}
	a.repoPath = path
	a.worktree = mgr
	return nil
}

// ListWorktrees 返回所有 worktrees
func (a *App) ListWorktrees() ([]worktree.Worktree, error) {
	if a.worktree == nil {
		return nil, fmt.Errorf("repository not initialized, call SetRepoPath first")
	}
	return a.worktree.List()
}

// GetWorktreeStatus 获取指定 worktree 的详细状态
func (a *App) GetWorktreeStatus(name string) (*worktree.WorktreeStatus, error) {
	if a.worktree == nil {
		return nil, fmt.Errorf("repository not initialized, call SetRepoPath first")
	}
	return a.worktree.GetStatusByName(name)
}

// ListBranches 返回所有分支
func (a *App) ListBranches() ([]string, error) {
	if a.worktree == nil {
		return nil, fmt.Errorf("repository not initialized, call SetRepoPath first")
	}
	return a.worktree.ListBranches()
}

// CreateWorktree 创建新的 worktree
func (a *App) CreateWorktree(opts worktree.CreateOptions) (*worktree.Worktree, error) {
	if a.worktree == nil {
		return nil, fmt.Errorf("repository not initialized, call SetRepoPath first")
	}
	return a.worktree.Create(opts)
}

// DeleteWorktree 删除 worktree
func (a *App) DeleteWorktree(name string, force bool) error {
	if a.worktree == nil {
		return fmt.Errorf("repository not initialized, call SetRepoPath first")
	}
	return a.worktree.Delete(name, force)
}

// GetRepoPath 返回当前仓库路径
func (a *App) GetRepoPath() string {
	return a.repoPath
}

// === Terminal API ===

// CreateOrAttachTerminal 创建或附加到终端
func (a *App) CreateOrAttachTerminal(id, worktreeId string) error {
	// 检查 worktree manager 是否初始化
	if a.worktree == nil {
		return fmt.Errorf("worktree manager not initialized")
	}

	// 获取 worktree 路径
	worktrees, err := a.worktree.List()
	if err != nil {
		return err
	}

	var cwd string
	for _, wt := range worktrees {
		if wt.ID == worktreeId || wt.Name == worktreeId {
			cwd = wt.Path
			break
		}
	}

	if cwd == "" {
		return fmt.Errorf("worktree not found: %s", worktreeId)
	}

	return a.terminal.CreateOrAttachSession(id, worktreeId, cwd)
}

// SendTerminalInput 发送终端输入
func (a *App) SendTerminalInput(id, data string) error {
	return a.terminal.SendInput(id, data)
}

// ResizeTerminal 调整终端大小
func (a *App) ResizeTerminal(id string, cols, rows uint16) error {
	return a.terminal.Resize(id, cols, rows)
}

// DetachTerminal 断开终端（保活，可重连）
func (a *App) DetachTerminal(id string) error {
	return a.terminal.DetachSession(id)
}

// DestroyTerminal 彻底销毁终端
func (a *App) DestroyTerminal(id string) error {
	return a.terminal.DestroySession(id)
}

// ListTerminalSessions 列出终端会话
func (a *App) ListTerminalSessions() []terminal.SessionInfo {
	return a.terminal.ListSessions()
}

// GetTerminalState 获取终端状态
func (a *App) GetTerminalState(id string) (string, error) {
	state, err := a.terminal.GetSessionState(id)
	if err != nil {
		return "", err
	}
	return string(state), nil
}

// HasTmux 返回 tmux 是否可用
func (a *App) HasTmux() bool {
	return a.terminal.HasTmux()
}
