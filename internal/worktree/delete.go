package worktree

import (
	"fmt"
	"os/exec"
)

// Delete 删除 worktree
func (m *Manager) Delete(name string, force bool) error {
	// 1. 禁止删除 main
	if name == "main" {
		return NewWorktreeError(ErrCannotDeleteMain,
			"cannot delete main worktree")
	}

	// 2. 获取 worktree 信息
	wt, err := m.getWorktreeByName(name)
	if err != nil {
		return err
	}

	// 3. 安全检查 (除非 force=true)
	if !force {
		if wt.HasChanges {
			return NewWorktreeError(ErrHasChanges,
				"worktree has uncommitted changes. Use force=true to delete anyway")
		}
		if wt.Unpushed > 0 {
			return NewWorktreeError(ErrHasUnpushed,
				fmt.Sprintf("worktree has %d unpushed commits. Use force=true to delete anyway", wt.Unpushed))
		}
	}

	// 4. 执行 git worktree remove
	return m.execWorktreeRemove(name, force)
}

// getWorktreeByName 根据名称获取 worktree
func (m *Manager) getWorktreeByName(name string) (*Worktree, error) {
	worktrees, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Name == name || wt.ID == name {
			return &wt, nil
		}
	}

	return nil, NewWorktreeError(ErrNotFound,
		fmt.Sprintf("worktree '%s' not found", name))
}

// execWorktreeRemove 执行 git worktree remove 命令
func (m *Manager) execWorktreeRemove(name string, force bool) error {
	args := []string{"worktree", "remove"}

	if force {
		args = append(args, "--force")
	}

	args = append(args, name)

	cmd := exec.Command("git", args...)
	cmd.Dir = m.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return NewWorktreeError(ErrDeleteFailed,
			fmt.Sprintf("git worktree remove failed: %s", string(output)))
	}

	return nil
}
