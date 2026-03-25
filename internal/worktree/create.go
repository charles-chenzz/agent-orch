package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// validNamePattern 有效的 worktree 名称
var validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`)

// branchToName 将分支名转换为有效的目录名
// feature/auth -> feature-auth, fix/bug-123 -> fix-bug-123
func branchToName(branch string) string {
	// 替换 / 为 -
	name := strings.ReplaceAll(branch, "/", "-")
	// 替换多个连续的 - 为单个
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	// 移除首尾的 -
	name = strings.Trim(name, "-")
	// 如果为空或以数字开头前面加 wt- 前缀
	if name == "" || (name[0] >= '0' && name[0] <= '9') {
		name = "wt-" + name
	}
	return name
}

// Create 创建新的 worktree
func (m *Manager) Create(opts CreateOptions) (*Worktree, error) {
	// 1. 如果 Name 为空，自动从 Branch 生成
	if opts.Name == "" {
		opts.Name = branchToName(opts.Branch)
	}

	// 2. 参数验证
	if err := m.validateCreateOptions(opts); err != nil {
		return nil, err
	}

	// 3. 检查名称冲突
	if err := m.checkNameConflict(opts.Name); err != nil {
		return nil, err
	}

	// 4. 计算目标路径（使用 Superset 风格）
	targetPath := m.getWorktreeTargetPath(opts.Name)

	// 5. 确保目标目录存在
	if err := m.ensureWorktreeDir(opts.Name); err != nil {
		return nil, NewWorktreeError(ErrCreateFailed,
			fmt.Sprintf("failed to create worktree directory: %s", err))
	}

	// 6. 检查路径是否已存在
	if _, err := os.Stat(targetPath); err == nil {
		return nil, NewWorktreeError(ErrPathExists,
			fmt.Sprintf("path already exists: %s", targetPath))
	}

	// 7. 检查分支
	if opts.CreateNew {
		if err := m.checkBranchExists(opts.BaseBranch); err != nil {
			return nil, err
		}
	} else {
		if err := m.checkBranchExists(opts.Branch); err != nil {
			return nil, err
		}
	}

	// 8. 执行 git worktree add
	if err := m.execWorktreeAdd(targetPath, opts); err != nil {
		return nil, err
	}

	// 9. 验证创建成功
	worktrees, err := m.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees after creation: %w", err)
	}

	for _, wt := range worktrees {
		if wt.Name == opts.Name {
			return &wt, nil
		}
	}

	return nil, NewWorktreeError(ErrCreateFailed,
		"worktree created but not found in list")
}

// validateCreateOptions 验证创建选项
func (m *Manager) validateCreateOptions(opts CreateOptions) error {
	if opts.Name == "" {
		return NewWorktreeError(ErrNameRequired, "name is required")
	}

	if !validNamePattern.MatchString(opts.Name) {
		return NewWorktreeError(ErrNameInvalid,
			"name must contain only letters, numbers, hyphens and underscores, must start with letter or number")
	}

	if opts.Branch == "" {
		return NewWorktreeError(ErrBranchRequired, "branch is required")
	}

	// 默认基础分支
	if opts.CreateNew && opts.BaseBranch == "" {
		// 尝试自动检测默认分支
		defaultBranch := "main"
		branches, err := m.ListBranches()
		if err == nil {
			for _, b := range branches {
				if b == "master" {
					defaultBranch = "master"
					break
				}
			}
		}
		opts.BaseBranch = defaultBranch
	}

	return nil
}

// checkNameConflict 检查名称冲突
func (m *Manager) checkNameConflict(name string) error {
	worktrees, err := m.List()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	for _, wt := range worktrees {
		if wt.Name == name || wt.ID == name {
			return NewWorktreeError(ErrNameConflict,
				fmt.Sprintf("worktree with name '%s' already exists", name))
		}
	}

	return nil
}

// checkBranchExists 检查分支是否存在
func (m *Manager) checkBranchExists(branch string) error {
	branches, err := m.ListBranches()
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	for _, b := range branches {
		if b == branch {
			return nil
		}
	}

	return NewWorktreeError(ErrBranchNotFound,
		fmt.Sprintf("branch '%s' not found", branch))
}

// execWorktreeAdd 执行 git worktree add 命令
func (m *Manager) execWorktreeAdd(targetPath string, opts CreateOptions) error {
	var args []string

	if opts.CreateNew {
		// git worktree add -b <new-branch> <path> <base-branch>
		args = []string{
			"worktree", "add",
			"-b", opts.Branch,
			targetPath,
			opts.BaseBranch,
		}
	} else {
		// git worktree add <path> <existing-branch>
		args = []string{
			"worktree", "add",
			targetPath,
			opts.Branch,
		}
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = m.repoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return NewWorktreeError(ErrGitFailed,
			fmt.Sprintf("git worktree add failed: %s", string(output)))
	}

	return nil
}
