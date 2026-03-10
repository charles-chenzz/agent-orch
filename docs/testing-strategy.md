# Testing Strategy

## 1. 测试金字塔

```
                    ▲
                   /│\
                  / │ \
                 /  │  \        E2E Tests
                /   │   \       (Playwright)
               /    │    \      - 用户流程测试
              /     │     \     - 跨平台测试
             /      │      \
            /───────┼───────\
           /        │        \
          /         │         \   Integration Tests
         /          │          \  - Go 后端集成测试
        /           │           \ - 前端组件测试
       /            │            \
      /─────────────┼─────────────\
     /              │              \
    /               │               \ Unit Tests
   /                │                \ (Go: testify, 前端: Vitest)
  /                 │                 \ - 函数级别测试
 /                  │                  \ - Mock 依赖
/___________________│___________________\
        70%                    20%           10%
```

## 2. 测试工具选型

### 2.1 Go 后端

| 工具 | 用途 |
|------|------|
| **testify** | 断言和 Mock |
| **gomock** | 接口 Mock 生成 |
| **testcontainers** | 集成测试容器 |
| **gotestsum** | 测试运行器（美化输出） |
| **golangci-lint** | 静态分析 |

### 2.2 前端

| 工具 | 用途 |
|------|------|
| **Vitest** | 单元测试 |
| **@testing-library/react** | 组件测试 |
| **Playwright** | E2E 测试 |
| **msw** | API Mock |

---

## 3. Go 测试规范

### 3.1 测试文件命名

```
internal/
├── worktree/
│   ├── manager.go
│   ├── manager_test.go        # 单元测试
│   ├── manager_integration_test.go  # 集成测试
│   └── manager_mock_test.go   # Mock 文件
```

### 3.2 测试函数命名

```go
// 功能_场景_期望结果
func TestCreateWorktree_ValidInput_ReturnsWorktree(t *testing.T) {}
func TestCreateWorktree_EmptyName_ReturnsError(t *testing.T) {}
func TestCreateWorktree_DuplicateName_ReturnsError(t *testing.T) {}

// 表驱动测试
func TestDetectAgentType(t *testing.T) {
    tests := []struct {
        name     string
        cmdline  string
        expected agent.AgentType
    }{
        {
            name:     "Claude Code",
            cmdline:  "node /path/to/@anthropic-ai/claude-code/cli.js",
            expected: agent.AgentClaudeCode,
        },
        {
            name:     "Codex",
            cmdline:  "python -m openai.codex",
            expected: agent.AgentCodex,
        },
        {
            name:     "Unknown",
            cmdline:  "some other process",
            expected: "",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := detectAgentType(tt.cmdline)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 3.3 Mock 示例

```go
//go:generate mockgen -source=types.go -destination=mock_test.go -package=worktree

type GitOperations interface {
    Open(path string) (*git.Repository, error)
    Worktree(repo *git.Repository) (*git.Worktree, error)
}

func TestManager_WithMockGit(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockGit := NewMockGitOperations(ctrl)
    mockGit.EXPECT().
        Open("/path/to/repo").
        Return(nil, errors.New("not a git repo"))
    
    m := &Manager{git: mockGit}
    
    _, err := m.List()
    assert.Error(t, err)
}
```

### 3.4 集成测试

```go
// +build integration

package worktree

import (
    "os"
    "os/exec"
    "testing"
    
    "github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
    // 检查 git 是否可用
    if _, err := exec.LookPath("git"); err != nil {
        fmt.Println("Skipping integration tests: git not found")
        os.Exit(0)
    }
    
    os.Exit(m.Run())
}

func TestManager_CreateRealWorktree(t *testing.T) {
    // 创建临时仓库
    tmpDir := t.TempDir()
    
    cmd := exec.Command("git", "init")
    cmd.Dir = tmpDir
    require.NoError(t, cmd.Run())
    
    // 创建初始提交
    // ...
    
    m, err := NewManager(tmpDir)
    require.NoError(t, err)
    
    // 测试真实操作
    _, err = m.Create(CreateOptions{
        Name:       "test-wt",
        Branch:     "test-branch",
        CreateNew:  true,
        BaseBranch: "main",
    })
    
    require.NoError(t, err)
}
```

### 3.5 测试覆盖率

```bash
# 运行测试并生成覆盖率报告
go test ./... -coverprofile=coverage.out

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html

# 目标：单元测试覆盖率 > 80%
```

---

## 4. 前端测试规范

### 4.1 Vitest 配置

```typescript
// vite.config.ts
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      reporter: ['text', 'html'],
      exclude: ['node_modules/', 'src/test/'],
    },
  },
})
```

### 4.2 组件测试示例

```typescript
// src/components/Worktree/WorktreeItem.test.tsx
import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import WorktreeItem from './WorktreeItem'

describe('WorktreeItem', () => {
  const mockWorktree = {
    id: 'test-1',
    name: 'feature/auth',
    branch: 'feature/auth',
    path: '/path/to/auth',
    head: 'abc1234',
    isMain: false,
    hasChanges: true,
    unpushed: 2,
  }
  
  it('renders worktree name', () => {
    render(
      <WorktreeItem
        worktree={mockWorktree}
        isSelected={false}
        onClick={() => {}}
      />
    )
    
    expect(screen.getByText('feature/auth')).toBeInTheDocument()
  })
  
  it('shows change indicator when hasChanges is true', () => {
    render(
      <WorktreeItem
        worktree={mockWorktree}
        isSelected={false}
        onClick={() => {}}
      />
    )
    
    expect(screen.getByText('Changes')).toBeInTheDocument()
  })
  
  it('calls onClick when clicked', () => {
    const handleClick = vi.fn()
    
    render(
      <WorktreeItem
        worktree={mockWorktree}
        isSelected={false}
        onClick={handleClick}
      />
    )
    
    fireEvent.click(screen.getByText('feature/auth'))
    
    expect(handleClick).toHaveBeenCalledTimes(1)
  })
  
  it('applies selected styles when selected', () => {
    const { container } = render(
      <WorktreeItem
        worktree={mockWorktree}
        isSelected={true}
        onClick={() => {}}
      />
    )
    
    expect(container.firstChild).toHaveClass('bg-gray-800')
  })
})
```

### 4.3 Hook 测试示例

```typescript
// src/hooks/useWorktree.test.ts
import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { useWorktreeStore } from '../stores/worktreeStore'

// Mock Wails 绑定
vi.mock('../../wailsjs/go/main/App', () => ({
  ListWorktrees: vi.fn().mockResolvedValue([
    { id: 'main', name: 'main', branch: 'main' }
  ]),
}))

describe('useWorktreeStore', () => {
  it('fetches worktrees', async () => {
    const { result } = renderHook(() => useWorktreeStore())
    
    await act(async () => {
      await result.current.fetchWorktrees()
    })
    
    expect(result.current.worktrees).toHaveLength(1)
    expect(result.current.worktrees[0].name).toBe('main')
  })
})
```

---

## 5. E2E 测试

### 5.1 Playwright 配置

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  
  use: {
    baseURL: 'http://localhost:34115',
    trace: 'on-first-retry',
  },
  
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  
  webServer: {
    command: 'wails dev',
    url: 'http://localhost:34115',
    reuseExistingServer: !process.env.CI,
  },
})
```

### 5.2 E2E 测试示例

```typescript
// e2e/worktree.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Worktree Management', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })
  
  test('displays main worktree on load', async ({ page }) => {
    // 等待 worktree 列表加载
    await expect(page.locator('[data-testid="worktree-list"]')).toBeVisible()
    
    // 应该显示 main worktree
    await expect(page.locator('text=main')).toBeVisible()
  })
  
  test('creates new worktree', async ({ page }) => {
    // 点击创建按钮
    await page.click('button:has-text("+ New Worktree")')
    
    // 填写表单
    await page.fill('input[placeholder="feature/my-feature"]', 'test-feature')
    await page.fill('input[placeholder="feature/auth"]', 'test-branch')
    
    // 提交
    await page.click('button:has-text("Create")')
    
    // 验证新 worktree 出现
    await expect(page.locator('text=test-feature')).toBeVisible()
  })
  
  test('shows error when creating duplicate worktree', async ({ page }) => {
    // 尝试创建已存在的 worktree
    await page.click('button:has-text("+ New Worktree")')
    await page.fill('input[placeholder="feature/my-feature"]', 'main')
    await page.click('button:has-text("Create")')
    
    // 应该显示错误
    await expect(page.locator('text=already exists')).toBeVisible()
  })
})
```

---

## 6. CI/CD 测试流程

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Run unit tests
        run: go test ./... -v -coverprofile=coverage.out
      
      - name: Run integration tests
        run: go test ./... -tags=integration -v
        if: github.event_name == 'push'
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: coverage.out
  
  test-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: Install dependencies
        working-directory: frontend
        run: npm ci
      
      - name: Run tests
        working-directory: frontend
        run: npm test
      
      - name: Type check
        working-directory: frontend
        run: npm run typecheck
  
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        
      - name: ESLint
        working-directory: frontend
        run: npm run lint
```

---

## 7. 测试覆盖率目标

| 类型 | 目标 |
|------|------|
| Go 单元测试 | > 80% |
| Go 集成测试 | 核心路径覆盖 |
| 前端组件测试 | > 70% |
| E2E 测试 | 关键用户流程 |

---

## 8. 测试最佳实践

### 8.1 Go 测试

1. **使用 t.Parallel()** 并行运行独立测试
2. **使用 t.Cleanup()** 清理资源
3. **使用 testify/assert** 提高可读性
4. **避免全局状态** 使用依赖注入
5. **Mock 外部依赖** 如文件系统、网络

### 8.2 前端测试

1. **测试用户行为** 而非实现细节
2. **使用 data-testid** 选择元素
3. **Mock Wails 绑定** 隔离前后端
4. **避免快照测试** 除非 UI 稳定
5. **使用 MSW Mock API**

---

## 9. 测试命令速查

```bash
# Go
go test ./...                    # 运行所有测试
go test ./... -v                 # 详细输出
go test ./... -run TestName      # 运行特定测试
go test ./... -cover             # 显示覆盖率
go test ./... -tags=integration  # 运行集成测试

# 前端
npm test                         # 运行所有测试
npm run test:watch               # 监听模式
npm run test:coverage            # 覆盖率报告

# E2E
npx playwright test              # 运行 E2E 测试
npx playwright test --ui         # UI 模式
npx playwright test --debug      # 调试模式
```
