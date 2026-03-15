package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Manager 管理 Git worktree 操作.
type Manager struct {
	repoPath string
	repo     *git.Repository
}

// NewManager 创建新的 worktree manager.
func NewManager(repoPath string) (*Manager, error) {
	// 转换为绝对路径
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 验证路径存在
	if _, err := os.Stat(absPath); err != nil {
		return nil, fmt.Errorf("repository path does not exist: %s", absPath)
	}

	// 打开仓库
	repo, err := git.PlainOpen(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return &Manager{
		repoPath: absPath,
		repo:     repo,
	}, nil
}

// List 返回所有 worktrees.
func (m *Manager) List() ([]Worktree, error) {
	worktrees := []Worktree{}

	// 获取 main worktree
	mainWt, err := m.getMainWorktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get main worktree: %w", err)
	}
	worktrees = append(worktrees, *mainWt)

	// 读取 .git/worktrees 目录
	worktreesPath := filepath.Join(m.repoPath, ".git", "worktrees")
	entries, err := os.ReadDir(worktreesPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 没有额外的 worktrees，只返回 main
			return worktrees, nil
		}
		return nil, fmt.Errorf("failed to read worktrees directory: %w", err)
	}

	// 添加其他 worktrees
	for _, entry := range entries {
		if entry.IsDir() {
			wt, err := m.parseWorktreeDir(worktreesPath, entry.Name())
			if err != nil {
				// 跳过无效的 worktree，记录错误但不中断
				continue
			}
			worktrees = append(worktrees, *wt)
		}
	}

	return worktrees, nil
}

// getMainWorktree 获取主 worktree 信息.
func (m *Manager) getMainWorktree() (*Worktree, error) {
	head, err := m.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	branch := ""
	if head.Name().IsBranch() {
		branch = head.Name().Short()
	}

	// 获取状态
	status, _ := m.GetStatus(m.repoPath)
	hasChanges := false
	unpushed := 0
	if status != nil {
		hasChanges = len(status.Staged) > 0 || len(status.Unstaged) > 0 || len(status.Untracked) > 0
		unpushed = status.Ahead
	}

	// 获取最后活动时间
	lastActivity := m.getLastActivity(m.repoPath)

	return &Worktree{
		ID:           "main",
		Name:         filepath.Base(m.repoPath),
		Path:         m.repoPath,
		Branch:       branch,
		Head:         head.Hash().String()[:7],
		IsMain:       true,
		HasChanges:   hasChanges,
		Unpushed:     unpushed,
		LastActivity: lastActivity,
	}, nil
}

// parseWorktreeDir 解析 worktree 目录信息.
// 直接读取 .git/worktrees/name/ 目录中的文件，而不是使用 go-git，
// 因为 go-git 对 linked worktree 的支持有限。
func (m *Manager) parseWorktreeDir(worktreesPath, name string) (*Worktree, error) {
	wtDir := filepath.Join(worktreesPath, name)

	// 读取 gitdir 文件获取实际路径
	gitdirFile := filepath.Join(wtDir, "gitdir")
	data, err := os.ReadFile(gitdirFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read gitdir: %w", err)
	}

	// 解析实际路径
	// gitdir 文件内容如: /path/to/worktree/.git
	// 实际 worktree 路径是其父目录
	gitdir := strings.TrimSpace(string(data))
	actualPath := filepath.Dir(gitdir)

	// 验证路径存在
	if _, err := os.Stat(actualPath); err != nil {
		return nil, fmt.Errorf("worktree path does not exist: %s", actualPath)
	}

	// 直接读取 HEAD 文件 (在 .git/worktrees/name/HEAD)
	headFile := filepath.Join(wtDir, "HEAD")
	headData, err := os.ReadFile(headFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read HEAD: %w", err)
	}

	headContent := strings.TrimSpace(string(headData))
	var branch, headHash string

	// HEAD 内容格式:
	// - "ref: refs/heads/branch-name" (在分支上)
	// - "abc1234..." (detached HEAD)
	if strings.HasPrefix(headContent, "ref: refs/heads/") {
		branch = strings.TrimPrefix(headContent, "ref: refs/heads/")
		// 需要从 refs/heads/branch 获取 commit hash
		refFile := filepath.Join(m.repoPath, ".git", "refs", "heads", branch)
		if hashData, err := os.ReadFile(refFile); err == nil {
			headHash = strings.TrimSpace(string(hashData))[:7]
		} else {
			// 可能在 packed-refs 中
			headHash = m.getCommitHashFromPackedRefs(branch)
		}
	} else {
		// Detached HEAD
		headHash = headContent[:7]
	}

	// 使用 git 命令获取状态 (go-git 对 worktree 支持有限)
	hasChanges, unpushed := m.getWorktreeStatusFromGit(actualPath, branch)

	// 获取最后活动时间
	lastActivity := m.getLastActivity(actualPath)

	return &Worktree{
		ID:           name,
		Name:         filepath.Base(actualPath),
		Path:         actualPath,
		Branch:       branch,
		Head:         headHash,
		IsMain:       false,
		HasChanges:   hasChanges,
		Unpushed:     unpushed,
		LastActivity: lastActivity,
	}, nil
}

// getWorktreeStatusFromGit 使用 git 命令获取 worktree 状态.
// go-git 对 linked worktree 的支持有限，使用 git 命令更可靠。
func (m *Manager) getWorktreeStatusFromGit(wtPath, branch string) (hasChanges bool, unpushed int) {
	// 检查是否有未提交的变更
	// git status --porcelain 返回空则无变更
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = wtPath
	output, err := cmd.Output()
	if err == nil {
		hasChanges = len(strings.TrimSpace(string(output))) > 0
	}

	// 检查 unpushed 提交数
	// git rev-list --count @{upstream}..HEAD
	if branch != "" {
		cmd = exec.Command("git", "rev-list", "--count", "@{upstream}..HEAD")
		cmd.Dir = wtPath
		output, err = cmd.Output()
		if err == nil {
			count := strings.TrimSpace(string(output))
			if count != "" && count != "0" {
				fmt.Sscanf(count, "%d", &unpushed)
			}
		}
	}

	return hasChanges, unpushed
}

// getCommitHashFromPackedRefs 从 packed-refs 文件获取 commit hash.
func (m *Manager) getCommitHashFromPackedRefs(branch string) string {
	packedRefsPath := filepath.Join(m.repoPath, ".git", "packed-refs")
	data, err := os.ReadFile(packedRefsPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	targetRef := "refs/heads/" + branch
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == targetRef {
			if len(parts[0]) >= 7 {
				return parts[0][:7]
			}
			return parts[0]
		}
	}
	return ""
}

// GetStatus 获取 worktree 的详细状态.
func (m *Manager) GetStatus(wtPath string) (*WorktreeStatus, error) {
	repo, err := git.PlainOpen(wtPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	status := &WorktreeStatus{
		Staged:    []FileStat{},
		Unstaged:  []FileStat{},
		Untracked: []string{},
	}

	// 获取 HEAD 信息
	head, err := repo.Head()
	if err == nil {
		status.Head = head.Hash().String()[:7]
		if head.Name().IsBranch() {
			status.Branch = head.Name().Short()
		}
	}

	// 获取工作区状态
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	s, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// 解析文件状态
	for file, fs := range s {
		// 暂存区状态
		if fs.Staging != git.Unmodified {
			stat := FileStat{Path: file}
			switch fs.Staging {
			case git.Added:
				stat.Status = "A"
			case git.Modified:
				stat.Status = "M"
			case git.Deleted:
				stat.Status = "D"
			case git.Renamed:
				stat.Status = "R"
			}
			if stat.Status != "" {
				status.Staged = append(status.Staged, stat)
			}
		}

		// 工作区状态
		if fs.Worktree != git.Unmodified {
			stat := FileStat{Path: file}
			switch fs.Worktree {
			case git.Modified:
				stat.Status = "M"
			case git.Deleted:
				stat.Status = "D"
			}
			if stat.Status != "" {
				status.Unstaged = append(status.Unstaged, stat)
			}
		}

		// 未跟踪文件
		if fs.Staging == git.Untracked && fs.Worktree == git.Untracked {
			status.Untracked = append(status.Untracked, file)
		}
	}

	// 计算 ahead/behind
	if head != nil && status.Branch != "" {
		status.Ahead, status.Behind = m.calculateAheadBehind(repo, head)
	}

	// 获取最后一次提交
	lastCommit, err := m.getLastCommit(repo)
	if err == nil {
		status.LastCommit = lastCommit
	}

	return status, nil
}

// calculateAheadBehind 计算 ahead/behind 提交数.
func (m *Manager) calculateAheadBehind(repo *git.Repository, head *plumbing.Reference) (ahead, behind int) {
	branch := head.Name().Short()
	remoteRef := plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", branch))

	remote, err := repo.Reference(remoteRef, true)
	if err != nil {
		return 0, 0
	}

	localHash := head.Hash()
	remoteHash := remote.Hash()

	// 使用 git 命令计算 (go-git 的 rev-list 实现较复杂)
	// git rev-list --left-right --count origin/branch...HEAD
	cmd := exec.Command("git", "rev-list", "--left-right", "--count",
		fmt.Sprintf("%s...%s", remoteHash.String(), localHash.String()))
	cmd.Dir = m.repoPath

	output, err := cmd.Output()
	if err != nil {
		// 回退到简单比较
		if localHash != remoteHash {
			return 1, 0
		}
		return 0, 0
	}

	// 解析输出 "behind\tahead"
	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) == 2 {
		fmt.Sscanf(parts[0], "%d", &behind)
		fmt.Sscanf(parts[1], "%d", &ahead)
	}

	return ahead, behind
}

// getLastCommit 获取最后一次提交信息.
func (m *Manager) getLastCommit(repo *git.Repository) (*CommitInfo, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	return &CommitInfo{
		Hash:    commit.Hash.String()[:7],
		Message: strings.Split(commit.Message, "\n")[0],
		Author:  commit.Author.Name,
		Time:    commit.Author.When,
	}, nil
}

// getLastActivity 获取最后 Git 活动时间.
func (m *Manager) getLastActivity(wtPath string) time.Time {
	// 检查几个可能的活动文件
	files := []string{
		filepath.Join(wtPath, ".git", "logs", "HEAD"),
		filepath.Join(wtPath, ".git", "index"),
	}

	var latestTime time.Time
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
			}
		}
	}

	return latestTime
}

// GetStatusByName 根据 worktree 名称获取状态.
func (m *Manager) GetStatusByName(name string) (*WorktreeStatus, error) {
	// 获取 worktree 列表
	worktrees, err := m.List()
	if err != nil {
		return nil, err
	}

	// 查找对应的 worktree
	for _, wt := range worktrees {
		if wt.Name == name || wt.ID == name {
			return m.GetStatus(wt.Path)
		}
	}

	return nil, fmt.Errorf("worktree not found: %s", name)
}

// ListBranches 列出所有分支.
func (m *Manager) ListBranches() ([]string, error) {
	branches := []string{}

	refs, err := m.repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branches = append(branches, ref.Name().Short())
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate references: %w", err)
	}

	return branches, nil
}

// GetRepoPath 返回仓库路径.
func (m *Manager) GetRepoPath() string {
	return m.repoPath
}
