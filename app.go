package main

import (
	"context"
	"fmt"

	"agent-orch/internal/worktree"
)

// App struct
type App struct {
	ctx       context.Context
	worktree  *worktree.Manager
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
