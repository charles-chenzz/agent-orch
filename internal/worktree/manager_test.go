package worktree

import (
	"encoding/json"
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

// ============ Create Tests ============

func TestManager_Create_ValidateOptions(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	tests := []struct {
		name    string
		opts    CreateOptions
		errCode string
	}{
		{
			name:    "empty name",
			opts:    CreateOptions{Name: "", Branch: "feature/test"},
			errCode: ErrNameRequired,
		},
		{
			name:    "invalid name with special chars",
			opts:    CreateOptions{Name: "test@invalid", Branch: "feature/test"},
			errCode: ErrNameInvalid,
		},
		{
			name:    "empty branch",
			opts:    CreateOptions{Name: "test-worktree", Branch: ""},
			errCode: ErrBranchRequired,
		},
		{
			name:    "name starts with hyphen",
			opts:    CreateOptions{Name: "-invalid", Branch: "feature/test"},
			errCode: ErrNameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Create(tt.opts)
			require.Error(t, err)

			var wtErr *WorktreeError
			require.ErrorAs(t, err, &wtErr)
			assert.Equal(t, tt.errCode, wtErr.Code)
		})
	}
}

func TestManager_Create_NameConflict(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 尝试使用 main 作为名称（已存在）
	opts := CreateOptions{
		Name:      "main",
		Branch:    "feature/test",
		CreateNew: true,
	}

	_, err = mgr.Create(opts)
	require.Error(t, err)

	var wtErr *WorktreeError
	require.ErrorAs(t, err, &wtErr)
	assert.Equal(t, ErrNameConflict, wtErr.Code)
}

func TestManager_Create_BranchNotFound(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	opts := CreateOptions{
		Name:      "test-worktree",
		Branch:    "nonexistent-branch",
		CreateNew: false, // 使用现有分支
	}

	_, err = mgr.Create(opts)
	require.Error(t, err)

	var wtErr *WorktreeError
	require.ErrorAs(t, err, &wtErr)
	assert.Equal(t, ErrBranchNotFound, wtErr.Code)
}

func TestManager_Create_ValidName(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 测试各种有效名称
	validNames := []string{
		"feature-auth",
		"fix_123",
		"worktree1",
		"TEST_WORKTREE",
		"a1-b2_c3",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			err := mgr.validateCreateOptions(CreateOptions{
				Name:   name,
				Branch: "main",
			})
			assert.NoError(t, err, "name '%s' should be valid", name)
		})
	}
}

// ============ Delete Tests ============

func TestManager_Delete_CannotDeleteMain(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	err = mgr.Delete("main", false)
	require.Error(t, err)

	var wtErr *WorktreeError
	require.ErrorAs(t, err, &wtErr)
	assert.Equal(t, ErrCannotDeleteMain, wtErr.Code)
}

func TestManager_Delete_NotFound(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	err = mgr.Delete("nonexistent-worktree", false)
	require.Error(t, err)

	var wtErr *WorktreeError
	require.ErrorAs(t, err, &wtErr)
	assert.Equal(t, ErrNotFound, wtErr.Code)
}

func TestManager_Delete_HasChanges(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 模拟有变更的 worktree（通过修改 main）
	// 在实际场景中，这会是一个真实的 worktree
	// 这里我们测试 getWorktreeByName 的逻辑

	// 修改文件
	err = os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# Modified"), 0644)
	require.NoError(t, err)

	// 获取 main worktree 信息
	wt, err := mgr.getWorktreeByName("main")
	require.NoError(t, err)
	assert.True(t, wt.HasChanges)
}

func TestManager_Delete_Force(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 删除 main 应该即使 force=true 也失败
	err = mgr.Delete("main", true)
	require.Error(t, err)

	var wtErr *WorktreeError
	require.ErrorAs(t, err, &wtErr)
	assert.Equal(t, ErrCannotDeleteMain, wtErr.Code)
}

func TestManager_GetWorktreeByName(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	// 通过 ID 查找
	wt, err := mgr.getWorktreeByName("main")
	require.NoError(t, err)
	assert.Equal(t, "main", wt.ID)
	assert.True(t, wt.IsMain)

	// 通过名称查找
	repoName := filepath.Base(repoPath)
	wt, err = mgr.getWorktreeByName(repoName)
	require.NoError(t, err)
	assert.Equal(t, repoName, wt.Name)
}

func TestManager_GetWorktreeByName_NotFound(t *testing.T) {
	repoPath := setupTestRepo(t)

	mgr, err := NewManager(repoPath)
	require.NoError(t, err)

	_, err = mgr.getWorktreeByName("nonexistent")
	require.Error(t, err)

	var wtErr *WorktreeError
	require.ErrorAs(t, err, &wtErr)
	assert.Equal(t, ErrNotFound, wtErr.Code)
}

// ============ Error Tests ============

func TestWorktreeError_Error(t *testing.T) {
	err := NewWorktreeError(ErrNameRequired, "name is required")
	assert.Equal(t, "[ERR_NAME_REQUIRED] name is required", err.Error())
}

func TestWorktreeError_Json(t *testing.T) {
	err := NewWorktreeError(ErrNameInvalid, "invalid name format")

	// 验证 JSON 序列化
	jsonData, jsonErr := json.Marshal(err)
	require.NoError(t, jsonErr)
	assert.Contains(t, string(jsonData), "ERR_NAME_INVALID")
	assert.Contains(t, string(jsonData), "invalid name format")
}
