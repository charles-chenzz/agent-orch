package worktree

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test Repository"), 0644)
	require.NoError(t, err)

	_, err = wt.Add("README.md")
	require.NoError(t, err)

	_, err = wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@test.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	return tmpDir
}

func TestNewManager(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)
	require.NotNil(t, mgr)
	assert.Equal(t, repoPath, mgr.GetRepoPath())
}

func TestNewManager_InvalidPath(t *testing.T) {
	_, err := NewManager("/nonexistent/path")
	assert.Error(t, err)
}

func TestGetStatus(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	status, err := mgr.GetStatus(repoPath)
	require.NoError(t, err)
	assert.NotEmpty(t, status.Branch)
	assert.NotEmpty(t, status.Head)
}

func TestManager_List(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	worktrees, err := mgr.List()
	require.NoError(t, err)
	require.Len(t, worktrees, 1)

	// 验证 main worktree
	mainWt := worktrees[0]
	assert.Equal(t, "main", mainWt.ID)
	assert.True(t, mainWt.IsMain)
	assert.Equal(t, repoPath, mainWt.Path)
	assert.NotEmpty(t, mainWt.Branch)
	assert.NotEmpty(t, mainWt.Head)
	assert.Equal(t, filepath.Base(repoPath), mainWt.Name)
}

func TestManager_List_EmptyWorktreesDir(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 没有 .git/worktrees 目录时，只返回 main
	worktrees, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)
	assert.Equal(t, "main", worktrees[0].ID)
}

func TestManager_GetStatusByName(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 使用 "main" 作为名称
	status, err := mgr.GetStatusByName("main")
	require.NoError(t, err)
	assert.NotEmpty(t, status.Branch)
	assert.NotEmpty(t, status.Head)
	assert.NotNil(t, status.LastCommit)
	assert.Equal(t, "Initial commit", status.LastCommit.Message)
}

func TestManager_GetStatusByName_WithRepoName(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 使用仓库名称作为名称
	repoName := filepath.Base(repoPath)
	status, err := mgr.GetStatusByName(repoName)
	require.NoError(t, err)
	assert.NotEmpty(t, status.Branch)
}

func TestManager_GetStatusByName_NotFound(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 使用不存在的名称
	_, err = mgr.GetStatusByName("nonexistent-worktree")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found")
}

func TestManager_GetStatus_WithChanges(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 创建新文件（未跟踪）
	err = os.WriteFile(filepath.Join(repoPath, "newfile.txt"), []byte("test"), 0644)
	require.NoError(t, err)

	status, err := mgr.GetStatus(repoPath)
	require.NoError(t, err)
	assert.Len(t, status.Untracked, 1)
	assert.Contains(t, status.Untracked, "newfile.txt")
}

func TestManager_GetStatus_WithStagedChanges(t *testing.T) {
	repoPath := setupTestRepo(t)

	// 打开仓库添加暂存文件
	repo, err := git.PlainOpen(repoPath)
	require.NoError(t, err)

	wt, err := repo.Worktree()
	require.NoError(t, err)

	// 修改文件
	err = os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# Modified"), 0644)
	require.NoError(t, err)

	// 暂存
	_, err = wt.Add("README.md")
	require.NoError(t, err)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	status, err := mgr.GetStatus(repoPath)
	require.NoError(t, err)
	assert.Len(t, status.Staged, 1)
	assert.Equal(t, "README.md", status.Staged[0].Path)
	assert.Equal(t, "M", status.Staged[0].Status)
}

func TestManager_ListBranches(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	branches, err := mgr.ListBranches()
	require.NoError(t, err)
	assert.NotEmpty(t, branches)
	// 默认分支可能是 master 或 main
	assert.Contains(t, []string{"main", "master"}, branches[0])
}

func TestManager_GetLastActivity(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	activity := mgr.getLastActivity(repoPath)
	// 刚创建的仓库，活动时间应该是最近的
	assert.False(t, activity.IsZero())
}
