# Phase 1: Git Worktree 管理

> **周期**：2 周
> **目标**：实现完整的 Git Worktree CRUD 功能
> **依赖**：Phase 0 (Foundation)
> **交付**：v0.2.0-alpha
> **状态**：🔄 **进行中** (2026-03-15)

---

## 1. Feature List

### 1.1 Go 后端 - Worktree 核心

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F1.1 | worktree.Manager 结构定义 | P0 | ✅ 完成 |
| F1.2 | List() - 列出所有 worktrees | P0 | ✅ 完成 |
| F1.3 | Create(name, branch string) - 创建 worktree | P0 | ✅ 完成 |
| F1.4 | Delete(name string) - 删除 worktree | P0 | ✅ 完成 |
| F1.5 | GetStatus(name string) - 获取 worktree 状态 | P0 | ✅ 完成 |

### 1.2 Go 后端 - Git 操作

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F1.6 | 分支安全检查 (未提交变更检测) | P0 | ✅ 完成 |
| F1.7 | 未推送提交检测 | P1 | ✅ 完成 |
| F1.8 | 最后 Git 活动时间戳 | P1 | ✅ 完成 |
| F1.9 | 当前分支信息 | P0 | ✅ 完成 |
| F1.10 | 远程分支列表 | P1 | ✅ 完成 |

### 1.3 前端 - Worktree 组件

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F1.11 | WorktreeList.tsx - 列表组件 | P0 | ⏸️ 待实现 |
| F1.12 | WorktreeItem.tsx - 单项组件 | P0 | ⏸️ 待实现 |
| F1.13 | CreateModal.tsx - 创建弹窗 | P0 | ⏸️ 待实现 |
| F1.14 | DeleteConfirm.tsx - 删除确认 | P0 | ⏸️ 待实现 |
| F1.15 | BranchStatusIndicator - 分支状态指示器 | P1 | ⏸️ 待实现 |

### 1.4 前端 - 状态管理

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F1.16 | useWorktree hook | P0 | ⏸️ 待实现 |
| F1.17 | worktreeStore (Zustand) | P0 | ⏸️ 待实现 |
| F1.18 | Wails IPC 绑定类型 | P0 | ✅ 完成 |

---

## 1.5 已实现功能 (2026-03-15)

### Go 后端实现

**文件结构**：
```
internal/worktree/
├── types.go           # 数据结构定义
├── manager.go         # Manager 实现 (List, GetStatus)
├── create.go          # Create 实现
├── delete.go          # Delete 实现
├── errors.go          # 结构化错误定义
├── manager_test.go    # 单元测试 (25+ 测试用例)
└── worktree_test.go   # 集成测试 (10+ 测试用例)
```

**核心功能**：

| 方法 | 功能 | 说明 |
|------|------|------|
| `NewManager(repoPath)` | 创建 Manager | 路径验证 + git.PlainOpen |
| `List()` | 列出所有 worktrees | main + .git/worktrees 目录 |
| `Create(opts)` | 创建新 worktree | 参数验证 + git worktree add |
| `Delete(name, force)` | 删除 worktree | 安全检查 + git worktree remove |
| `GetStatus(path)` | 获取详细状态 | staged/unstaged/untracked |
| `GetStatusByName(name)` | 按名称获取状态 | 便捷方法 |
| `ListBranches()` | 列出所有分支 | 额外功能 |
| `calculateAheadBehind()` | 计算 ahead/behind | 使用 git rev-list |
| `getLastCommit()` | 获取最后提交 | CommitInfo |
| `getLastActivity()` | 获取最后活动时间 | 读取 .git/logs/HEAD |

**Create/Delete 实现细节**：

```go
// Create 创建新的 worktree
func (m *Manager) Create(opts CreateOptions) (*Worktree, error) {
    // 1. 参数验证 (名称格式、分支要求)
    // 2. 检查名称冲突
    // 3. 检查路径是否已存在
    // 4. 检查分支是否存在
    // 5. 执行 git worktree add
    // 6. 验证创建成功并返回
}

// Delete 删除 worktree
func (m *Manager) Delete(name string, force bool) error {
    // 1. 禁止删除 main
    // 2. 获取 worktree 信息
    // 3. 安全检查 (未提交变更、未推送提交)
    // 4. 执行 git worktree remove (可选 --force)
}
```

**结构化错误处理**：

```go
type WorktreeError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// 错误码定义
const (
    ErrNameRequired     = "ERR_NAME_REQUIRED"
    ErrNameInvalid      = "ERR_NAME_INVALID"
    ErrBranchRequired   = "ERR_BRANCH_REQUIRED"
    ErrBranchNotFound   = "ERR_BRANCH_NOT_FOUND"
    ErrNameConflict     = "ERR_NAME_CONFLICT"
    ErrPathExists       = "ERR_PATH_EXISTS"
    ErrCreateFailed     = "ERR_CREATE_FAILED"
    ErrCannotDeleteMain = "ERR_CANNOT_DELETE_MAIN"
    ErrHasChanges       = "ERR_HAS_CHANGES"
    ErrHasUnpushed      = "ERR_HAS_UNPUSHED"
    ErrNotFound         = "ERR_NOT_FOUND"
    ErrDeleteFailed     = "ERR_DELETE_FAILED"
)
```

**go-git Worktree 限制处理**：

由于 go-git 对 linked worktree 支持有限，我们采用混合策略：
- **读取操作**：直接读取 `.git/worktrees/name/` 目录中的文件
- **状态检查**：使用 `git status --porcelain` 命令（更可靠）
- **创建/删除**：使用 `git worktree add/remove` 命令

```go
// parseWorktreeDir 直接读取 worktree 元数据
func (m *Manager) parseWorktreeDir(worktreesPath, name string) (*Worktree, error) {
    // 读取 gitdir 文件获取实际路径
    // 读取 HEAD 文件获取分支信息
    // 使用 git 命令获取状态（go-git 不支持 linked worktree）
}

// getWorktreeStatusFromGit 使用 git 命令获取状态
func (m *Manager) getWorktreeStatusFromGit(wtPath, branch string) (hasChanges bool, unpushed int) {
    // git status --porcelain (检查变更)
    // git rev-list --count @{upstream}..HEAD (检查未推送)
}
```

**数据结构**：

```go
// Worktree - worktree 基本信息
type Worktree struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Path         string    `json:"path"`
    Branch       string    `json:"branch"`
    Head         string    `json:"head"`         // commit hash 前7位
    IsMain       bool      `json:"isMain"`
    HasChanges   bool      `json:"hasChanges"`
    Unpushed     int       `json:"unpushed"`
    LastActivity time.Time `json:"lastActivity"`
}

// WorktreeStatus - 详细状态
type WorktreeStatus struct {
    WorktreeID string       `json:"worktreeId"`
    Branch     string       `json:"branch"`
    Head       string       `json:"head"`
    Ahead      int          `json:"ahead"`
    Behind     int          `json:"behind"`
    Staged     []FileStat   `json:"staged"`
    Unstaged   []FileStat   `json:"unstaged"`
    Untracked  []string     `json:"untracked"`
    LastCommit *CommitInfo  `json:"lastCommit"`
}
```

**Wails 绑定 (app.go)**：

```go
func (a *App) SetRepoPath(path string) error
func (a *App) ListWorktrees() ([]worktree.Worktree, error)
func (a *App) GetWorktreeStatus(name string) (*worktree.WorktreeStatus, error)
func (a *App) ListBranches() ([]string, error)
```

**测试覆盖**：

单元测试 (25+ 测试用例):

| 测试 | 覆盖场景 |
|------|----------|
| TestNewManager | 正常创建 |
| TestNewManager_InvalidPath | 无效路径 |
| TestGetStatus | 状态获取 |
| TestManager_List | 列表返回 |
| TestManager_List_EmptyWorktreesDir | 无额外 worktrees |
| TestManager_GetStatusByName | 按名称查询 |
| TestManager_GetStatusByName_WithRepoName | 仓库名查询 |
| TestManager_GetStatusByName_NotFound | 不存在处理 |
| TestManager_GetStatus_WithChanges | 未跟踪文件 |
| TestManager_GetStatus_WithStagedChanges | 已暂存文件 |
| TestManager_ListBranches | 分支列表 |
| TestManager_GetLastActivity | 活动时间 |
| TestManager_Create_ValidateOptions | 创建参数验证 |
| TestManager_Create_NameConflict | 名称冲突 |
| TestManager_Create_BranchNotFound | 分支不存在 |
| TestManager_Create_ValidName | 有效名称 |
| TestManager_Delete_CannotDeleteMain | 禁止删除 main |
| TestManager_Delete_NotFound | 不存在处理 |
| TestManager_Delete_HasChanges | 有变更检查 |
| TestManager_Delete_Force | 强制删除 |
| TestManager_GetWorktreeByName | 按名称获取 |
| TestWorktreeError_* | 错误处理 |

集成测试 (10+ 测试用例，需要 `go test -tags=integration`):

| 测试 | 覆盖场景 |
|------|----------|
| TestIntegration_Create_Success | 创建成功 |
| TestIntegration_Create_InvalidName | 无效名称 |
| TestIntegration_Create_PathExists | 路径存在 |
| TestIntegration_Create_NameConflict | 名称冲突 |
| TestIntegration_Create_BranchNotFound | 分支不存在 |
| TestIntegration_Delete_Success | 删除成功 |
| TestIntegration_Delete_HasChanges | 有变更检查 |
| TestIntegration_Delete_CannotDeleteMain | 禁止删除 main |
| TestIntegration_Delete_NotFound | 不存在处理 |
| TestIntegration_Isolation | 测试隔离验证 |

---

## 2. 实现细节

### 2.1 Go 后端 - 数据结构

```go
// internal/worktree/types.go
package worktree

import "time"

type Worktree struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Path         string    `json:"path"`
    Branch       string    `json:"branch"`
    Head         string    `json:"head"`         // 当前 commit hash
    IsMain       bool      `json:"isMain"`       // 是否为主 worktree
    HasChanges   bool      `json:"hasChanges"`   // 是否有未提交变更
    Unpushed     int       `json:"unpushed"`     // 未推送提交数
    LastActivity time.Time `json:"lastActivity"` // 最后活动时间
}

type CreateOptions struct {
    Name       string `json:"name"`
    Branch     string `json:"branch"`
    BaseBranch string `json:"baseBranch"` // 创建分支的基础分支
    CreateNew  bool   `json:"createNew"`  // 是否创建新分支
}

type WorktreeStatus struct {
    WorktreeID  string      `json:"worktreeId"`
    Branch      string      `json:"branch"`
    Head        string      `json:"head"`
    Ahead       int         `json:"ahead"`       // 领先远程的提交数
    Behind      int         `json:"behind"`      // 落后远程的提交数
    Staged      []FileStat  `json:"staged"`      // 已暂存文件
    Unstaged    []FileStat  `json:"unstaged"`    // 未暂存文件
    Untracked   []string    `json:"untracked"`   // 未跟踪文件
    LastCommit  *CommitInfo `json:"lastCommit"`  // 最后一次提交
}

type FileStat struct {
    Path    string `json:"path"`
    Status  string `json:"status"` // M, A, D, R
}

type CommitInfo struct {
    Hash    string    `json:"hash"`
    Message string    `json:"message"`
    Author  string    `json:"author"`
    Time    time.Time `json:"time"`
}
```

### 2.2 Go 后端 - Manager 实现

```go
// internal/worktree/manager.go
package worktree

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/reference"
)

type Manager struct {
    repoPath string   // 主仓库路径
    repo     *git.Repository
}

func NewManager(repoPath string) (*Manager, error) {
    repo, err := git.PlainOpen(repoPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open repository: %w", err)
    }
    
    return &Manager{
        repoPath: repoPath,
        repo:     repo,
    }, nil
}

// List 返回所有 worktrees
func (m *Manager) List() ([]Worktree, error) {
    // 读取 .git/worktrees 目录
    worktreesPath := filepath.Join(m.repoPath, ".git", "worktrees")
    
    entries, err := os.ReadDir(worktreesPath)
    if err != nil {
        if os.IsNotExist(err) {
            // 没有额外的 worktrees，只有 main
            return m.getMainWorktree()
        }
        return nil, err
    }
    
    worktrees := []Worktree{}
    
    // 添加 main worktree
    mainWt, err := m.getMainWorktree()
    if err == nil {
        worktrees = append(worktrees, mainWt...)
    }
    
    // 添加其他 worktrees
    for _, entry := range entries {
        if entry.IsDir() {
            wtPath := filepath.Join(worktreesPath, entry.Name())
            wt, err := m.parseWorktree(wtPath, entry.Name())
            if err != nil {
                continue // 跳过无效的 worktree
            }
            worktrees = append(worktrees, *wt)
        }
    }
    
    return worktrees, nil
}

func (m *Manager) getMainWorktree() ([]Worktree, error) {
    head, err := m.repo.Head()
    if err != nil {
        return nil, err
    }
    
    branch := ""
    if head.Name().IsBranch() {
        branch = head.Name().Short()
    }
    
    status, _ := m.GetStatus(m.repoPath)
    
    wt := Worktree{
        ID:         "main",
        Name:       "main",
        Path:       m.repoPath,
        Branch:     branch,
        Head:       head.Hash().String()[:7],
        IsMain:     true,
        HasChanges: len(status.Staged) > 0 || len(status.Unstaged) > 0,
        Unpushed:   status.Ahead,
    }
    
    return []Worktree{wt}, nil
}

func (m *Manager) parseWorktree(wtPath, name string) (*Worktree, error) {
    // 读取 gitdir 文件获取实际路径
    gitdirFile := filepath.Join(wtPath, "gitdir")
    data, err := os.ReadFile(gitdirFile)
    if err != nil {
        return nil, err
    }
    
    // 解析实际路径
    gitdir := strings.TrimSpace(string(data))
    actualPath := filepath.Dir(filepath.Dir(gitdir))
    
    // 打开该 worktree 的仓库
    wtRepo, err := git.PlainOpen(actualPath)
    if err != nil {
        return nil, err
    }
    
    head, err := wtRepo.Head()
    if err != nil {
        return nil, err
    }
    
    branch := ""
    if head.Name().IsBranch() {
        branch = head.Name().Short()
    }
    
    status, _ := m.GetStatus(actualPath)
    
    return &Worktree{
        ID:         name,
        Name:       filepath.Base(actualPath),
        Path:       actualPath,
        Branch:     branch,
        Head:       head.Hash().String()[:7],
        IsMain:     false,
        HasChanges: len(status.Staged) > 0 || len(status.Unstaged) > 0,
        Unpushed:   status.Ahead,
    }, nil
}

// Create 创建新的 worktree
func (m *Manager) Create(opts CreateOptions) (*Worktree, error) {
    // 1. 验证名称
    if opts.Name == "" {
        return nil, fmt.Errorf("name is required")
    }
    
    // 2. 确定目标路径
    parentPath := filepath.Dir(m.repoPath)
    targetPath := filepath.Join(parentPath, opts.Name)
    
    // 3. 检查路径是否已存在
    if _, err := os.Stat(targetPath); err == nil {
        return nil, fmt.Errorf("path already exists: %s", targetPath)
    }
    
    // 4. 执行 git worktree add
    // 使用 git 命令（go-git 不直接支持 worktree 操作）
    args := []string{"worktree", "add"}
    
    if opts.CreateNew {
        // 创建新分支: git worktree add -b <branch> <path> <base>
        args = append(args, "-b", opts.Branch, targetPath, opts.BaseBranch)
    } else {
        // 使用现有分支: git worktree add <path> <branch>
        args = append(args, targetPath, opts.Branch)
    }
    
    cmd := exec.Command("git", args...)
    cmd.Dir = m.repoPath
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("git worktree add failed: %s", string(output))
    }
    
    // 5. 返回新创建的 worktree
    worktrees, err := m.List()
    if err != nil {
        return nil, err
    }
    
    for _, wt := range worktrees {
        if wt.Name == opts.Name {
            return &wt, nil
        }
    }
    
    return nil, fmt.Errorf("worktree created but not found")
}

// Delete 删除 worktree
func (m *Manager) Delete(name string) error {
    if name == "main" {
        return fmt.Errorf("cannot delete main worktree")
    }
    
    // 检查是否有未提交的变更
    worktrees, err := m.List()
    if err != nil {
        return err
    }
    
    for _, wt := range worktrees {
        if wt.Name == name && wt.HasChanges {
            return fmt.Errorf("worktree has uncommitted changes, please commit or stash first")
        }
    }
    
    // 执行 git worktree remove
    cmd := exec.Command("git", "worktree", "remove", name)
    cmd.Dir = m.repoPath
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        // 如果 remove 失败，尝试 force
        cmd = exec.Command("git", "worktree", "remove", "--force", name)
        cmd.Dir = m.repoPath
        output, err = cmd.CombinedOutput()
        if err != nil {
            return fmt.Errorf("git worktree remove failed: %s", string(output))
        }
    }
    
    return nil
}

// GetStatus 获取 worktree 的详细状态
func (m *Manager) GetStatus(wtPath string) (*WorktreeStatus, error) {
    repo, err := git.PlainOpen(wtPath)
    if err != nil {
        return nil, err
    }
    
    status := &WorktreeStatus{}
    
    // 获取工作区状态
    worktree, err := repo.Worktree()
    if err != nil {
        return nil, err
    }
    
    s, err := worktree.Status()
    if err != nil {
        return nil, err
    }
    
    for file, fs := range s {
        stat := FileStat{Path: file}
        
        if fs.Staging != git.Unmodified {
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
            status.Staged = append(status.Staged, stat)
        }
        
        if fs.Worktree != git.Unmodified {
            stat2 := FileStat{Path: file}
            switch fs.Worktree {
            case git.Modified:
                stat2.Status = "M"
            case git.Deleted:
                stat2.Status = "D"
            }
            status.Unstaged = append(status.Unstaged, stat2)
        }
        
        if fs.Staging == git.Untracked && fs.Worktree == git.Untracked {
            status.Untracked = append(status.Untracked, file)
        }
    }
    
    // 获取 ahead/behind
    head, _ := repo.Head()
    if head != nil {
        branch := head.Name().Short()
        remoteRef := fmt.Sprintf("refs/remotes/origin/%s", branch)
        
        remote, err := repo.Reference(plumbing.ReferenceName(remoteRef), true)
        if err == nil {
            // 计算 ahead/behind
            localHash := head.Hash()
            remoteHash := remote.Hash()
            
            localCommit, _ := repo.CommitObject(localHash)
            remoteCommit, _ := repo.CommitObject(remoteHash)
            
            if localCommit != nil && remoteCommit != nil {
                // 简化计算：只比较是否相同
                if localHash != remoteHash {
                    status.Ahead = 1 // 实际应该用 rev-list 计算
                }
            }
        }
        
        status.Branch = branch
        status.Head = head.Hash().String()[:7]
    }
    
    return status, nil
}
```

### 2.3 App.go 绑定

```go
// app.go (新增方法)
func (a *App) ListWorktrees() ([]worktree.Worktree, error) {
    return a.worktree.List()
}

func (a *App) CreateWorktree(opts worktree.CreateOptions) (*worktree.Worktree, error) {
    return a.worktree.Create(opts)
}

func (a *App) DeleteWorktree(name string) error {
    return a.worktree.Delete(name)
}

func (a *App) GetWorktreeStatus(name string) (*worktree.WorktreeStatus, error) {
    wt, err := a.worktree.GetWorktree(name)
    if err != nil {
        return nil, err
    }
    return a.worktree.GetStatus(wt.Path)
}

func (a *App) GetBranches() ([]string, error) {
    return a.worktree.ListBranches()
}
```

### 2.4 前端 - 类型定义

```typescript
// frontend/src/types/worktree.ts
export interface Worktree {
  id: string
  name: string
  path: string
  branch: string
  head: string
  isMain: boolean
  hasChanges: boolean
  unpushed: number
  lastActivity?: string
}

export interface CreateOptions {
  name: string
  branch: string
  baseBranch?: string
  createNew: boolean
}

export interface WorktreeStatus {
  worktreeId: string
  branch: string
  head: string
  ahead: number
  behind: number
  staged: FileStat[]
  unstaged: FileStat[]
  untracked: string[]
  lastCommit?: CommitInfo
}

export interface FileStat {
  path: string
  status: 'M' | 'A' | 'D' | 'R'
}

export interface CommitInfo {
  hash: string
  message: string
  author: string
  time: string
}
```

### 2.5 前端 - Store

```typescript
// frontend/src/stores/worktreeStore.ts
import { create } from 'zustand'
import { Worktree, CreateOptions, WorktreeStatus } from '../types/worktree'
import { 
  ListWorktrees, 
  CreateWorktree, 
  DeleteWorktree, 
  GetWorktreeStatus 
} from '../../wailsjs/go/main/App'

interface WorktreeState {
  worktrees: Worktree[]
  selectedId: string | null
  status: WorktreeStatus | null
  loading: boolean
  error: string | null
  
  // Actions
  fetchWorktrees: () => Promise<void>
  selectWorktree: (id: string) => Promise<void>
  createWorktree: (opts: CreateOptions) => Promise<void>
  deleteWorktree: (name: string) => Promise<void>
  clearError: () => void
}

export const useWorktreeStore = create<WorktreeState>((set, get) => ({
  worktrees: [],
  selectedId: null,
  status: null,
  loading: false,
  error: null,
  
  fetchWorktrees: async () => {
    set({ loading: true, error: null })
    try {
      const worktrees = await ListWorktrees()
      set({ worktrees, loading: false })
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },
  
  selectWorktree: async (id: string) => {
    set({ selectedId: id, loading: true })
    try {
      const status = await GetWorktreeStatus(id)
      set({ status, loading: false })
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },
  
  createWorktree: async (opts: CreateOptions) => {
    set({ loading: true, error: null })
    try {
      await CreateWorktree(opts)
      await get().fetchWorktrees()
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },
  
  deleteWorktree: async (name: string) => {
    set({ loading: true, error: null })
    try {
      await DeleteWorktree(name)
      await get().fetchWorktrees()
      // 如果删除的是当前选中的，清除选中
      if (get().selectedId === name) {
        set({ selectedId: null, status: null })
      }
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },
  
  clearError: () => set({ error: null }),
}))
```

### 2.6 前端 - 组件

```tsx
// frontend/src/components/Worktree/WorktreeList.tsx
import { useEffect } from 'react'
import { useWorktreeStore } from '../../stores/worktreeStore'
import WorktreeItem from './WorktreeItem'
import CreateModal from './CreateModal'

interface WorktreeListProps {
  onSelect: (id: string) => void
  selectedId: string | null
}

export default function WorktreeList({ onSelect, selectedId }: WorktreeListProps) {
  const { worktrees, loading, error, fetchWorktrees } = useWorktreeStore()
  const [showCreate, setShowCreate] = useState(false)
  
  useEffect(() => {
    fetchWorktrees()
  }, [])
  
  if (loading && worktrees.length === 0) {
    return <div className="p-4 text-gray-500">Loading...</div>
  }
  
  return (
    <div className="flex flex-col h-full">
      {/* 头部 */}
      <div className="p-4 border-b border-gray-700">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-400 uppercase">Worktrees</h2>
          <button
            onClick={() => setShowCreate(true)}
            className="p-1 hover:bg-gray-700 rounded"
            title="Create worktree"
          >
            <PlusIcon className="w-4 h-4" />
          </button>
        </div>
      </div>
      
      {/* 错误提示 */}
      {error && (
        <div className="p-2 bg-red-900/50 text-red-300 text-sm">
          {error}
        </div>
      )}
      
      {/* 列表 */}
      <div className="flex-1 overflow-auto">
        {worktrees.map(wt => (
          <WorktreeItem
            key={wt.id}
            worktree={wt}
            isSelected={selectedId === wt.id}
            onClick={() => onSelect(wt.id)}
          />
        ))}
      </div>
      
      {/* 创建弹窗 */}
      {showCreate && (
        <CreateModal onClose={() => setShowCreate(false)} />
      )}
    </div>
  )
}
```

```tsx
// frontend/src/components/Worktree/WorktreeItem.tsx
import { Worktree } from '../../types/worktree'

interface WorktreeItemProps {
  worktree: Worktree
  isSelected: boolean
  onClick: () => void
}

export default function WorktreeItem({ worktree, isSelected, onClick }: WorktreeItemProps) {
  return (
    <div
      className={`
        p-3 cursor-pointer border-l-2 transition-colors
        ${isSelected 
          ? 'bg-gray-800 border-blue-500' 
          : 'border-transparent hover:bg-gray-800/50'}
      `}
      onClick={onClick}
    >
      <div className="flex items-center gap-2">
        <span className="font-medium">{worktree.name}</span>
        {worktree.isMain && (
          <span className="px-1.5 py-0.5 text-xs bg-blue-900 text-blue-300 rounded">
            main
          </span>
        )}
      </div>
      
      <div className="flex items-center gap-2 mt-1 text-sm text-gray-500">
        <GitBranchIcon className="w-3 h-3" />
        <span>{worktree.branch}</span>
        <span className="text-gray-600">•</span>
        <span>{worktree.head}</span>
      </div>
      
      {/* 状态指示器 */}
      <div className="flex items-center gap-2 mt-1">
        {worktree.hasChanges && (
          <span className="flex items-center gap-1 text-yellow-500 text-xs">
            <DotIcon className="w-2 h-2 fill-current" />
            Changes
          </span>
        )}
        {worktree.unpushed > 0 && (
          <span className="flex items-center gap-1 text-blue-400 text-xs">
            <ArrowUpIcon className="w-3 h-3" />
            {worktree.unpushed}
          </span>
        )}
      </div>
    </div>
  )
}
```

```tsx
// frontend/src/components/Worktree/CreateModal.tsx
import { useState } from 'react'
import { useWorktreeStore } from '../../stores/worktreeStore'

interface CreateModalProps {
  onClose: () => void
}

export default function CreateModal({ onClose }: CreateModalProps) {
  const { createWorktree, loading, error } = useWorktreeStore()
  const [name, setName] = useState('')
  const [branch, setBranch] = useState('')
  const [createNewBranch, setCreateNewBranch] = useState(true)
  const [baseBranch, setBaseBranch] = useState('main')
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    await createWorktree({
      name,
      branch,
      baseBranch,
      createNew: createNewBranch,
    })
    
    if (!error) {
      onClose()
    }
  }
  
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-lg p-6 w-96">
        <h3 className="text-lg font-semibold mb-4">Create Worktree</h3>
        
        <form onSubmit={handleSubmit}>
          <div className="space-y-4">
            {/* 名称 */}
            <div>
              <label className="block text-sm text-gray-400 mb-1">Name</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                className="w-full px-3 py-2 bg-gray-700 rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                placeholder="feature/my-feature"
                required
              />
            </div>
            
            {/* 分支选项 */}
            <div>
              <label className="flex items-center gap-2 text-sm text-gray-400">
                <input
                  type="checkbox"
                  checked={createNewBranch}
                  onChange={e => setCreateNewBranch(e.target.checked)}
                />
                Create new branch
              </label>
            </div>
            
            {/* 分支名 */}
            <div>
              <label className="block text-sm text-gray-400 mb-1">
                {createNewBranch ? 'New Branch Name' : 'Existing Branch'}
              </label>
              <input
                type="text"
                value={branch}
                onChange={e => setBranch(e.target.value)}
                className="w-full px-3 py-2 bg-gray-700 rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                placeholder={createNewBranch ? 'feature/auth' : 'main'}
                required
              />
            </div>
            
            {/* 基础分支 */}
            {createNewBranch && (
              <div>
                <label className="block text-sm text-gray-400 mb-1">Base Branch</label>
                <select
                  value={baseBranch}
                  onChange={e => setBaseBranch(e.target.value)}
                  className="w-full px-3 py-2 bg-gray-700 rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
                >
                  <option value="main">main</option>
                  <option value="develop">develop</option>
                </select>
              </div>
            )}
          </div>
          
          {/* 错误 */}
          {error && (
            <div className="mt-4 p-2 bg-red-900/50 text-red-300 text-sm rounded">
              {error}
            </div>
          )}
          
          {/* 按钮 */}
          <div className="flex justify-end gap-2 mt-6">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-400 hover:text-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
```

---

## 3. 测试计划

### 3.1 单元测试

```go
// internal/worktree/manager_test.go
package worktree

import (
    "os"
    "path/filepath"
    "testing"
    
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
    // 创建临时仓库
    tmpDir := t.TempDir()
    
    repo, err := git.PlainInit(tmpDir, false)
    require.NoError(t, err)
    
    // 创建初始提交
    wt, err := repo.Worktree()
    require.NoError(t, err)
    
    // 创建文件
    err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644)
    require.NoError(t, err)
    
    _, err = wt.Add("README.md")
    require.NoError(t, err)
    
    _, err = wt.Commit("Initial commit", &git.CommitOptions{
        Author: &object.Signature{
            Name:  "Test",
            Email: "test@test.com",
        },
    })
    require.NoError(t, err)
    
    return tmpDir
}

func TestManager_List(t *testing.T) {
    repoPath := setupTestRepo(t)
    
    m, err := NewManager(repoPath)
    require.NoError(t, err)
    
    worktrees, err := m.List()
    require.NoError(t, err)
    
    assert.Len(t, worktrees, 1)
    assert.Equal(t, "main", worktrees[0].ID)
    assert.True(t, worktrees[0].IsMain)
}

func TestManager_Create(t *testing.T) {
    repoPath := setupTestRepo(t)
    
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
    assert.Equal(t, "feature/test", wt.Branch)
    
    // 验证目录存在
    _, err = os.Stat(wt.Path)
    assert.NoError(t, err)
    
    // 验证能列出
    worktrees, err := m.List()
    require.NoError(t, err)
    assert.Len(t, worktrees, 2)
}

func TestManager_Delete(t *testing.T) {
    repoPath := setupTestRepo(t)
    
    m, err := NewManager(repoPath)
    require.NoError(t, err)
    
    // 先创建
    opts := CreateOptions{
        Name:       "to-delete",
        Branch:     "feature/delete",
        BaseBranch: "main",
        CreateNew:  true,
    }
    _, err = m.Create(opts)
    require.NoError(t, err)
    
    // 再删除
    err = m.Delete("to-delete")
    require.NoError(t, err)
    
    // 验证只剩 main
    worktrees, err := m.List()
    require.NoError(t, err)
    assert.Len(t, worktrees, 1)
}

func TestManager_DeleteMain(t *testing.T) {
    repoPath := setupTestRepo(t)
    
    m, err := NewManager(repoPath)
    require.NoError(t, err)
    
    // 删除 main 应该失败
    err = m.Delete("main")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "cannot delete main worktree")
}

func TestManager_GetStatus(t *testing.T) {
    repoPath := setupTestRepo(t)
    
    m, err := NewManager(repoPath)
    require.NoError(t, err)
    
    status, err := m.GetStatus(repoPath)
    require.NoError(t, err)
    
    assert.Equal(t, "main", status.Branch)
    assert.NotEmpty(t, status.Head)
}
```

### 3.2 集成测试

```go
// internal/worktree/integration_test.go
// +build integration

package worktree

import (
    "os/exec"
    "testing"
    
    "github.com/stretchr/testify/require"
)

func TestManager_CreateWithRealGit(t *testing.T) {
    // 检查 git 是否可用
    _, err := exec.LookPath("git")
    if err != nil {
        t.Skip("git not available")
    }
    
    // 使用真实 git 命令测试
    // ...
}
```

### 3.3 E2E 测试清单

- [ ] 打开应用，显示 main worktree
- [ ] 点击 "Create Worktree"，填写表单，创建成功
- [ ] 新 worktree 出现在列表中
- [ ] 点击 worktree，右侧显示状态
- [ ] 创建有未提交变更的 worktree，尝试删除 → 失败提示
- [ ] 提交变更后，删除成功
- [ ] 刷新应用，worktree 列表正确

---

## 4. 验收标准

| 标准 | 描述 |
|------|------|
| 列表显示 | 能正确列出所有 worktrees（包括 main） |
| 创建功能 | 能创建新 worktree（新分支/现有分支） |
| 删除功能 | 能删除 worktree（有保护检查） |
| 状态显示 | 显示分支、commit、变更状态 |
| 错误处理 | 优雅处理各种错误（路径存在、未提交变更等） |

---

## 5. 发布检查清单

- [ ] 所有单元测试通过
- [ ] 集成测试通过（需要 git 环境）
- [ ] E2E 测试清单完成
- [ ] golangci-lint 通过
- [ ] TypeScript 检查通过
- [ ] 更新 CHANGELOG.md
- [ ] 创建 Git tag: v0.2.0-alpha
- [ ] GitHub Release 构建

---

## 6. 依赖更新

### Go 依赖

```go
// 新增
require (
    github.com/go-git/go-git/v5 v5.12.0
    github.com/go-git/go-git-fixtures/v4 v4.3.2-0.20231010084843-55aef409d66f // test
)
```

---

## 8. 技术方案与权衡

### 8.1 问题：go-git 不直接支持 worktree 操作

`go-git/v5` 库支持读取 worktree 信息，但不支持创建/删除 worktree 操作。需要通过 `exec.Command` 调用系统 `git` 命令。

### 8.2 潜在风险

| 风险类型 | 说明 | 影响程度 |
|---------|------|----------|
| **Git 版本依赖** | 不同 Git 版本命令参数可能有差异 | 中 |
| **环境依赖** | 用户可能未安装 Git 或 Git 不在 PATH | 高 |
| **跨平台差异** | Windows/macOS/Linux 上 Git 行为可能不同 | 中 |
| **命令注入** | 参数未正确转义可能导致 shell 注入 | 高 |
| **错误解析** | 需要解析命令输出字符串，容易出错 | 低 |
| **性能开销** | 每次操作都要启动新进程 (~10-50ms) | 低 |
| **原子性** | 命令执行中途失败，状态可能不一致 | 中 |
| **测试困难** | 单元测试需要真实 Git 环境 | 低 |

### 8.3 替代方案对比

#### 方案 A: exec.Command 调用 git 命令 (当前方案)

```
优点:
- 快速实现，功能完整
- 支持所有 Git 功能
- 无额外依赖

缺点:
- 依赖系统安装 Git
- 需要解析字符串输出
- 有命令注入风险
```

#### 方案 B: CGO + libgit2

```
优点:
- 无需系统 Git
- 性能更好
- 类型安全 API

缺点:
- 需要 CGO，交叉编译复杂
- libgit2 安装配置繁琐
- 仍不完全支持 worktree 操作
```

#### 方案 C: Rust Sidecar (gitoxide)

```
优点:
- 纯 Rust，无 CGO 依赖
- gitoxide 完整支持 worktree
- 高性能，内存安全
- 可独立测试

缺点:
- 架构复杂度增加
- 需要维护两套代码
- 进程间通信开销
```

**gitoxide 仓库**: https://github.com/Byron/gitoxide

```rust
// gitoxide 示例 (Rust)
use gix::Repository;

let repo = Repository::open(".")?;
let worktrees = repo.worktrees()?.iter().collect::<Vec<_>>();
```

### 8.4 方案对比总结

| 对比项 | exec.Command (git) | CGO + libgit2 | Rust Sidecar |
|--------|-------------------|---------------|--------------|
| **外部依赖** | 需要系统 Git | 需要 libgit2 | 无 |
| **性能** | 进程启动开销 | 内存调用 | 内存调用 |
| **错误处理** | 解析字符串 | 结构化错误 | 结构化错误 |
| **安全性** | 需手动转义 | 类型安全 | 类型安全 |
| **跨平台** | 依赖 Git 行为 | 需编译各平台 | 需编译各平台 |
| **可测试性** | 需要 mock 进程 | 可 mock 接口 | 可 mock 接口 |
| **功能覆盖** | 完整 | 部分 | 完整 |
| **实现复杂度** | 低 | 中 | 高 |

### 8.5 推荐策略

| 阶段 | 方案 | 理由 |
|------|------|------|
| **Phase 1 (v0.2)** | exec.Command | 快速实现，验证功能 |
| **Phase 2-3 (v0.3-0.4)** | 混合架构 | go-git 读 + exec 写，优化读取性能 |
| **v1.0+** | Rust Sidecar | 生产级可靠性，消除 Git 依赖 |

### 8.6 缓解措施 (exec.Command 方案)

```go
// 1. 启动时检查 Git 版本
func checkGitVersion() error {
    out, err := exec.Command("git", "--version").Output()
    if err != nil {
        return errors.New("git not found: please install git >= 2.17")
    }
    // 解析版本，确保 >= 2.17 (worktree 稳定支持)
    version := parseGitVersion(string(out))
    if version < "2.17" {
        return fmt.Errorf("git version too old: %s, need >= 2.17", version)
    }
    return nil
}

// 2. 安全转义参数 (防止命令注入)
import "github.com/alessio/shellescape"

func (m *Manager) Create(opts CreateOptions) (*Worktree, error) {
    // 使用 shellescape 转义用户输入
    args := []string{
        "worktree", "add",
        "-b", shellescape.Quote(opts.Branch),
        shellescape.Quote(targetPath),
        shellescape.Quote(opts.BaseBranch),
    }
    cmd := exec.Command("git", args...)
    // ...
}

// 3. 结构化错误解析
type GitError struct {
    Code    int
    Message string
    Hint    string
}

func parseGitError(output []byte, err error) *GitError {
    // 解析 git 输出的错误信息
    // 提取有用的 hint 信息
    lines := strings.Split(string(output), "\n")
    ge := &GitError{Code: 1}
    for _, line := range lines {
        if strings.HasPrefix(line, "hint:") {
            ge.Hint = strings.TrimPrefix(line, "hint: ")
        } else if strings.HasPrefix(line, "fatal:") {
            ge.Message = strings.TrimPrefix(line, "fatal: ")
        } else if strings.HasPrefix(line, "error:") {
            ge.Message = strings.TrimPrefix(line, "error: ")
        }
    }
    return ge
}

// 4. 接口抽象 (便于后续替换实现)
type WorktreeBackend interface {
    List() ([]Worktree, error)
    Create(opts CreateOptions) (*Worktree, error)
    Delete(name string) error
    GetStatus(path string) (*WorktreeStatus, error)
}

// 当前实现
type GitCommandBackend struct {
    repoPath string
}

// 未来可替换为
type GitoxideBackend struct {
    // Rust sidecar via IPC
}
```

### 8.7 混合架构示意 (Phase 2+)

```
┌─────────────────────────────────────────┐
│           Go Backend (Wails)            │
├─────────────────────────────────────────┤
│  go-git (读取操作)  │  exec (写入操作)   │
│  - List worktrees   │  - Create          │
│  - GetStatus        │  - Delete          │
│  - Read commits     │  - Branch ops      │
│  - Ahead/Behind     │                    │
└─────────────────────────────────────────┘
```

### 8.8 Rust Sidecar 架构 (v1.0+)

```
┌────────────────┐     JSON-RPC/IPC     ┌────────────────┐
│   Go Backend   │ ◄──────────────────► │  Rust gitoxide │
│   (Wails)      │                      │  (Sidecar)     │
│                │                      │                │
│  - UI/IPC      │                      │  - Worktree    │
│  - State mgmt  │                      │  - Git ops     │
│  - Terminal    │                      │  - Status      │
└────────────────┘                      └────────────────┘
```

---

## 9. 时间估算

| 任务 | 时间 |
|------|------|
| Go Manager 实现 | 3 天 |
| Go 测试编写 | 1.5 天 |
| 前端 Store + Hooks | 1 天 |
| 前端组件 | 2 天 |
| 集成测试 | 1 天 |
| 文档和发布 | 0.5 天 |
| Buffer | 1 天 |
| **总计** | **10 天 (2 周)** |
