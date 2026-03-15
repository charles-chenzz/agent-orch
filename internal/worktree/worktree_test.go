//go:build integration

package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RequiresGit 检查 git 是否可用
func RequiresGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

// SetupIsolatedRepo 创建隔离的临时仓库
func SetupIsolatedRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	// 使用 --initial-branch=main --no-template
	cmd := exec.Command("git", "init", "--initial-branch=main", "--no-template", repoPath)
	err := cmd.Run()
	require.NoError(t, err, "git init failed")

	exec.Command("git", "-C", repoPath, "config", "user.name", "test").Run()
	exec.Command("git", "-C", repoPath, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", repoPath, "config", "core.hooksPath", "/dev/null").Run()

	// 创建初始提交
	err = os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# Test"), 0644)
	require.NoError(t, err)

	cmd = exec.Command("git", "-C", repoPath, "add", "README.md")
	err = cmd.Run()
	require.NoError(t, err, "git add failed")

	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "Initial commit")
	err = cmd.Run()
	require.NoError(t, err, "git commit failed")

	return repoPath
}

// === Create 测试 ===

func TestIntegration_Create_Success(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	opts := CreateOptions{
		Name:       "test-feature",
		Branch:     "feature/test",
		BaseBranch: "main",
		CreateNew:  true,
	}

	wt, err := m.Create(opts)
	require.NoError(t, err)
	assert.Equal(t, "test-feature", wt.Name)
	assert.DirExists(t, filepath.Join(filepath.Dir(repoPath), wt.Name))
}

func TestIntegration_Create_InvalidName(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	// 测试无效名称
	invalidNames := []string{
		"",            // 空名称
		"test/../hack", // 路径穿越
		"test name",    // 空格
		"test$var",     // 特殊字符
	}

	for _, name := range invalidNames {
		_, err := m.Create(CreateOptions{Name: name, Branch: "test"})
		require.Error(t, err, "name '%s' should be invalid", name)
	}
}

func TestIntegration_Create_PathExists(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	// 创建同名目录
	existingPath := filepath.Join(filepath.Dir(repoPath), "existing")
	os.MkdirAll(existingPath, 0755)

	_, err = m.Create(CreateOptions{Name: "existing", Branch: "test", CreateNew: true})
	require.Error(t, err)
}

func TestIntegration_Create_NameConflict(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	// 先创建一个
	_, err = m.Create(CreateOptions{
		Name:       "conflict",
		Branch:     "feature/conflict",
		BaseBranch: "main",
		CreateNew:  true,
	})
	require.NoError(t, err)

	// 再创建同名
	_, err = m.Create(CreateOptions{Name: "conflict", Branch: "test2", BaseBranch: "main", CreateNew: true})
	require.Error(t, err)
}

func TestIntegration_Create_BranchNotFound(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	// 使用不存在的分支
	_, err = m.Create(CreateOptions{
		Name:      "test",
		Branch:    "nonexistent",
		CreateNew: false,
	})
	require.Error(t, err)
}

// === Delete 测试 ===

func TestIntegration_Delete_Success(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	// 先创建
	_, err = m.Create(CreateOptions{
		Name:       "to-delete",
		Branch:     "feature/delete",
		BaseBranch: "main",
		CreateNew:  true,
	})
	require.NoError(t, err)

	// 再删除
	err = m.Delete("to-delete", false)
	require.NoError(t, err)

	// 验证只剩 main
	worktrees, err := m.List()
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)
}

func TestIntegration_Delete_HasChanges(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	// 创建 worktree
	_, err = m.Create(CreateOptions{
		Name:       "has-changes",
		Branch:     "feature/changes",
		BaseBranch: "main",
		CreateNew:  true,
	})
	require.NoError(t, err)

	// 添加未提交变更
	wtPath := filepath.Join(filepath.Dir(repoPath), "has-changes", "test.txt")
	err = os.WriteFile(wtPath, []byte("test"), 0644)
	require.NoError(t, err)

	// 尝试删除 (应该失败)
	err = m.Delete("has-changes", false)
	require.Error(t, err)

	// 强制删除 (应该成功)
	err = m.Delete("has-changes", true)
	require.NoError(t, err)
}

func TestIntegration_Delete_CannotDeleteMain(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	// 尝试删除 main (应该失败)
	err = m.Delete("main", false)
	require.Error(t, err)
}

func TestIntegration_Delete_NotFound(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	// 尝试删除不存在的 worktree
	err = m.Delete("nonexistent", false)
	require.Error(t, err)
}

// === 测试隔离验证 ===

func TestIntegration_Isolation(t *testing.T) {
	RequiresGit(t)
	repoPath := SetupIsolatedRepo(t)
	m, err := NewManager(repoPath)
	require.NoError(t, err)

	// 1. 验证临时目录
	assert.NotEqual(t, repoPath, filepath.Dir(repoPath))

	// 2. 创建 worktree
	_, err = m.Create(CreateOptions{
		Name:       "isolation-test",
		Branch:     "feature/isolation",
		BaseBranch: "main",
		CreateNew:  true,
	})
	require.NoError(t, err)

	// 3. 验证 worktree 路径在临时目录内
	wtPath := filepath.Join(filepath.Dir(repoPath), "isolation-test")
	assert.DirExists(t, wtPath)

	// 4. 清理
	err = m.Delete("isolation-test", true)
	require.NoError(t, err)

	// 5. 验证清理后目录不存在
	_, err = os.Stat(wtPath)
	assert.True(t, os.IsNotExist(err), "worktree directory should be deleted")

	// 6. 验证全局配置未被修改
	globalConfig, _ := exec.Command("git", "config", "--global", "user.name").Output()
	assert.NotEqual(t, "test", strings.TrimSpace(string(globalConfig)))
}
