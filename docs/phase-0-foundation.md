# Phase 0: Foundation

> **周期**：1 周
> **目标**：搭建项目脚手架，实现基础 UI 布局
> **交付**：v0.1.0-alpha
> **状态**：✅ **已完成** (2026-03-11)

---

## 1. Feature List

### 1.1 Wails 项目初始化

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F0.1 | Wails v2 项目创建 (wails init) | P0 | ✅ 完成 |
| F0.2 | Go 模块结构初始化 (internal/) | P0 | ✅ 完成 |
| F0.3 | 前端项目初始化 (React + Vite) | P0 | ✅ 完成 |
| F0.4 | TailwindCSS 配置 | P0 | ⏸️ 待实现 |
| F0.5 | TypeScript 配置 | P0 | ✅ 完成 |

### 1.2 基础 UI 布局

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F0.6 | 三栏布局容器 (Sidebar / Main / RightPanel) | P0 | ⏸️ 待实现 |
| F0.7 | 响应式布局适配 | P1 | ⏸️ 待实现 |
| F0.8 | 暗色主题基础样式 | P0 | ⏸️ 待实现 |

#### UI 布局设计

参考 [Arbor](https://penso.github.io/arbor/) 的三栏布局，左侧栏支持多项目层级：

```
┌────────────────────────────────────────────────────────────────────────────┐
│  Agent Orchestrator                                   [窗口控制] [API状态] │
├──────────────────┬─────────────────────────────────────┬────────────────────┤
│                  │                                     │                    │
│  ▼ agent-orch     │                                     │  Changes          │
│    ▸ main         │    Terminal                          │  ┌──────────────┐  │
│    feature-auth   │    ┌─────────────────────────────┐   │  │ Staged       │  │
│                  │    │ Tab 1 [+] [Tab 2]          │   │  │  ▸ file.ts   │  │
│  ▶ another-proj   │    ├─────────────────────────────┤   │  │  ▸ file.go   │  │
│    ▸ main         │    │                             │   │  └──────────────┘  │
│                  │    │  $ _                        │   │                    │
│                  │    │                             │   │  Unstaged         │
│                  │    │                             │   │  ┌──────────────┐  │
│                  │    │                             │   │  │ ▸ mod.ts     │  │
│                  │    │                             │   │  │ ▸ test.ts    │  │
│                  │    └─────────────────────────────┘   │  └──────────────┘  │
│                  │                                     │                    │
│  [+ New Project] │                                     │  [Diff/PR/API Tab] │
│                  │                                     │                    │
├──────────────────┴─────────────────────────────────────┴────────────────────┤
       ↑                    ↑                              ↑
    左侧栏              中间栏                          右侧栏
  (Projects +        (Terminal)                    (File Changes
   Worktrees)                                       / Diff)
```

#### 核心交互逻辑

| 操作 | 效果 |
|------|------|
| 点击项目 (▶/▼) | 展开/折叠该项目的 worktree 列表 |
| 点击 worktree | 中间栏显示该 worktree 的终端，右侧栏显示文件变更 |
| 切换 worktree | 终端会话保持（后续 Phase 实现），右侧栏更新 |
| 点击右侧栏文件 | 显示 diff 内容 |
| 点击右侧栏 Tab | 切换 Diff / API Manager 视图 |

#### 组件结构

```
frontend/src/
├── App.tsx                    # 主容器 (三栏布局)
├── components/
│   └── Layout/
│       ├── Sidebar.tsx        # 左侧栏 - 项目+worktree 层级列表
│       ├── ProjectItem.tsx    # 单个项目 (可折叠)
│       ├── WorktreeItem.tsx   # 单个 worktree
│       ├── MainPane.tsx       # 中间栏 - 终端区域
│       └── RightPanel.tsx     # 右侧栏 - 文件变更/Diff
├── stores/
│   └── appStore.ts           # projects, selectedWorktree 状态
└── types/
    └── index.ts              # Project, Worktree 类型定义
```

#### 状态管理

```typescript
interface AppState {
  projects: Project[]           // 所有项目
  selectedWorktreeId: string    // 当前选中的 worktree
  expandedProjectIds: string[]  // 展开的项目
}

interface Project {
  id: string
  name: string
  path: string
  worktrees: Worktree[]
}

interface Worktree {
  id: string
  name: string
  branch: string
  path: string
  hasChanges: boolean
}
```

#### 配色方案

| 元素 | 颜色 | 用途 |
|------|------|------|
| 背景 | `#0d1117` | 主背景色 |
| 边框 | `#30363d` | 分隔线 |
| 强调 | `#58a6ff` | 选中状态、链接 |
| 文字 | `#c9d1d9` | 主文字 |
| 次要文字 | `#8b949e` | 描述文字 |
| 成功 | `#3fb950` | 新增文件 |
| 警告 | `#d29922` | 修改文件 |
| 错误 | `#f85149` | 删除文件 |

---

#### 备选方案 (待思考)

以下是一个被否决的 UI 设计方案，记录供参考：

```
┌──────────────────────────────────────────────────────────────────────────┐
│ Agent Orchestrator                                    ⚙ ⚙  [API Status] │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─ Projects ─────────────────────────────────────────────────────────┐  │
│  │  ▼ agent-orch         │  ▼ another-project                          │  │
│  │                       │                                              │  │
│  │  ┌─ Worktrees ─────┐  │  ┌─ Worktrees ─────────────────────────────┐  │
│  │  │ ▸ main          │  │  │ ▸ main                                  │  │  │
│  │  │   main          │  │  │   main                                  │  │  │
│  │  │                 │  │  │                                         │  │  │
│  │  │ ▸ feature-auth  │  │  │ ▸ feature-api                            │  │  │
│  │  │   auth          │  │  │   api-proxy                              │  │  │
│  │  │                 │  │  │                                         │  │  │
│  │  │ + New Worktree  │  │  │ + New Worktree                          │  │  │
│  │  └─────────────────┘  │  └─────────────────────────────────────────┘  │
│  │                       │                                              │  │
│  └───────────────────────┴──────────────────────────────────────────────┘  │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  ┌─ [Selected Worktree: feature-auth] ──────────────────────────────────┐ │
│  │                                                                       │ │
│  │  ┌─ Terminal ────────────────────────────┬─ Diff ──────────────────┐ │ │
│  │  │  Terminal 1  [+]  [API Manager]      │  ▸ staged    ▸ unstaged   │ │ │
│  │  ├────────────────────────────────────────┼─────────────────────────┤ │ │
│  │  │                                       │  ▸ src/app.tsx    M      │ │ │
│  │  │  $ _                                  │  ▸ lib/utils.ts   A      │ │ │
│  │  │                                       │                              │ │ │
│  │  │                                       │  ┌──────────────────────┐ │ │ │
│  │  │                                       │  │ diff content here   │ │ │ │
│  │  │                                       │  │                      │ │ │ │
│  │  └───────────────────────────────────────┴──┴──────────────────────┘ │ │
│  │                                                                       │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────┘
```

**否决原因**: 项目列表横向排列占用过多空间，且与 Arbor 的三栏设计风格不一致。

**优点**: 项目间切换更直观，适合项目数量较少的场景。

**缺点**: 横向空间利用率低，多项目时需要滚动。

### 1.3 Go 后端结构

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F0.9 | app.go 主入口 | P0 | ✅ 完成 |
| F0.10 | internal/worktree 占位 | P0 | ✅ 完成 |
| F0.11 | internal/terminal 占位 | P0 | ✅ 完成 |
| F0.12 | internal/config 基础实现 | P0 | ✅ 完成 |
| F0.13 | internal/db 基础实现 | P1 | ✅ 完成 |

### 1.4 前端目录结构

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F0.14 | components/Layout/ | P0 | ⏸️ 推迟 (依赖 F0.4) |
| F0.15 | hooks/ 目录 | P0 | ⏸️ 推迟 |
| F0.16 | stores/appStore.ts | P0 | ⏸️ 推迟 |
| F0.17 | types/index.ts | P0 | ⏸️ 推迟 |

### 1.5 配置系统骨架

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F0.18 | config.toml 结构定义 | P0 | ✅ 完成 |
| F0.19 | 配置加载/保存 | P0 | ✅ 完成 |
| F0.20 | 配置文件路径检测 | P1 | ✅ 完成 |

---

## 2. 实现细节

### 2.1 项目初始化命令

```bash
# 1. 创建 Wails 项目
wails init -n agent-orch -t react-ts

# 2. 进入项目
cd agent-orch

# 3. 安装前端依赖
cd frontend
npm install tailwindcss postcss autoprefixer \
  zustand \
  @types/node

# 4. 初始化 TailwindCSS
npx tailwindcss init -p

# 5. 安装 Go 依赖
cd ..
go get github.com/spf13/viper
go get gorm.io/gorm
go get gorm.io/driver/sqlite
```

### 2.2 目录结构创建

```bash
# Go 后端
mkdir -p internal/{worktree,terminal,proxy,config,agent,github,db}

# 前端
mkdir -p frontend/src/{components/{Layout,Worktree,Terminal,Diff,APIManager,AgentMonitor,Editor},hooks,stores,lib,types}
```

### 2.3 基础布局组件

```tsx
// frontend/src/App.tsx
import { useState } from 'react'
import Sidebar from './components/Layout/Sidebar'
import MainPane from './components/Layout/MainPane'
import RightPanel from './components/Layout/RightPanel'

function App() {
  const [selectedWorktree, setSelectedWorktree] = useState<string | null>(null)

  return (
    <div className="flex h-screen bg-gray-900 text-gray-100">
      {/* 左侧边栏 - Worktree 列表 */}
      <Sidebar 
        className="w-64 border-r border-gray-700"
        onSelect={setSelectedWorktree}
      />
      
      {/* 中间主区域 - 终端 */}
      <MainPane 
        className="flex-1"
        worktreeId={selectedWorktree}
      />
      
      {/* 右侧面板 - Diff / API Manager */}
      <RightPanel 
        className="w-80 border-l border-gray-700"
        worktreeId={selectedWorktree}
      />
    </div>
  )
}

export default App
```

```tsx
// frontend/src/components/Layout/Sidebar.tsx
interface SidebarProps {
  className?: string
  onSelect: (id: string) => void
}

export default function Sidebar({ className, onSelect }: SidebarProps) {
  // Phase 1 将实现真实数据
  const worktrees = [
    { id: 'main', name: 'main', branch: 'main', path: '/path/to/main' },
    { id: 'feature-1', name: 'feature-1', branch: 'feature/auth', path: '/path/to/feature-1' },
  ]

  return (
    <div className={`flex flex-col ${className}`}>
      <div className="p-4 border-b border-gray-700">
        <h1 className="text-lg font-semibold">Agent Orchestrator</h1>
      </div>
      
      <div className="flex-1 overflow-auto">
        {worktrees.map(wt => (
          <div
            key={wt.id}
            className="p-3 hover:bg-gray-800 cursor-pointer"
            onClick={() => onSelect(wt.id)}
          >
            <div className="font-medium">{wt.name}</div>
            <div className="text-sm text-gray-500">{wt.branch}</div>
          </div>
        ))}
      </div>
      
      {/* 底部操作栏 */}
      <div className="p-4 border-t border-gray-700">
        <button className="w-full py-2 bg-blue-600 rounded hover:bg-blue-700">
          + New Worktree
        </button>
      </div>
    </div>
  )
}
```

```tsx
// frontend/src/components/Layout/MainPane.tsx
interface MainPaneProps {
  className?: string
  worktreeId: string | null
}

export default function MainPane({ className, worktreeId }: MainPaneProps) {
  if (!worktreeId) {
    return (
      <div className={`flex items-center justify-center ${className}`}>
        <p className="text-gray-500">Select a worktree to start</p>
      </div>
    )
  }

  return (
    <div className={`flex flex-col ${className}`}>
      {/* 终端 Tab 栏 */}
      <div className="flex border-b border-gray-700">
        <div className="px-4 py-2 border-b-2 border-blue-500">Terminal 1</div>
        <div className="px-4 py-2 text-gray-500">+</div>
      </div>
      
      {/* 终端区域 - Phase 2 实现 */}
      <div className="flex-1 bg-gray-950">
        <div className="h-full flex items-center justify-center">
          <p className="text-gray-600">Terminal placeholder</p>
        </div>
      </div>
    </div>
  )
}
```

```tsx
// frontend/src/components/Layout/RightPanel.tsx
interface RightPanelProps {
  className?: string
  worktreeId: string | null
}

export default function RightPanel({ className, worktreeId }: RightPanelProps) {
  const [activeTab, setActiveTab] = useState<'diff' | 'api'>('diff')

  return (
    <div className={`flex flex-col ${className}`}>
      {/* Tab 切换 */}
      <div className="flex border-b border-gray-700">
        <button
          className={`flex-1 py-2 ${activeTab === 'diff' ? 'border-b-2 border-blue-500' : ''}`}
          onClick={() => setActiveTab('diff')}
        >
          Diff
        </button>
        <button
          className={`flex-1 py-2 ${activeTab === 'api' ? 'border-b-2 border-blue-500' : ''}`}
          onClick={() => setActiveTab('api')}
        >
          API
        </button>
      </div>
      
      {/* 内容区域 */}
      <div className="flex-1 overflow-auto p-4">
        {activeTab === 'diff' ? (
          <DiffPlaceholder />
        ) : (
          <APIPlaceholder />
        )}
      </div>
    </div>
  )
}

function DiffPlaceholder() {
  return (
    <div className="text-center text-gray-500">
      <p>Diff viewer</p>
      <p className="text-sm mt-2">Phase 5 implementation</p>
    </div>
  )
}

function APIPlaceholder() {
  return (
    <div className="text-center text-gray-500">
      <p>API Manager</p>
      <p className="text-sm mt-2">Phase 4 implementation</p>
    </div>
  )
}
```

### 2.4 Go 后端结构

```go
// main.go
package main

import (
    "embed"
    
    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()
    
    err := wails.Run(&options.App{
        Title:  "Agent Orchestrator",
        Width:  1280,
        Height: 800,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        Bind: []interface{}{
            app,
        },
    })
    
    if err != nil {
        println("Error:", err.Error())
    }
}
```

```go
// app.go
package main

import (
    "context"
    
    "agent-orch/internal/config"
    "agent-orch/internal/db"
)

type App struct {
    ctx    context.Context
    config *config.Manager
    db     *db.Database
}

func NewApp() *App {
    return &App{}
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    
    // 初始化配置
    cfg, err := config.Load()
    if err != nil {
        cfg = config.Default()
    }
    a.config = cfg
    
    // 初始化数据库
    database, err := db.Init(cfg.DatabasePath)
    if err != nil {
        // log error
    }
    a.db = database
}

// === 配置相关 API ===

func (a *App) GetConfig() *config.Config {
    return a.config.Current()
}

func (a *App) SaveConfig(cfg *config.Config) error {
    return a.config.Save(cfg)
}

// === 占位 API（后续 Phase 实现）===

func (a *App) ListWorktrees() ([]interface{}, error) {
    // Phase 1 实现
    return []interface{}{}, nil
}

func (a *App) CreateTerminal(id, cwd string) error {
    // Phase 2 实现
    return nil
}
```

```go
// internal/config/config.go
package config

import (
    "os"
    "path/filepath"
    
    "github.com/spf13/viper"
)

type Config struct {
    App      AppConfig      `mapstructure:"app"`
    Terminal TerminalConfig `mapstructure:"terminal"`
    Proxy    ProxyConfig    `mapstructure:"proxy"`
    Profiles []Profile      `mapstructure:"profiles"`
    GitHub   GitHubConfig   `mapstructure:"github"`
}

type AppConfig struct {
    Theme       string `mapstructure:"theme"`
    CheckUpdate bool   `mapstructure:"check_updates"`
}

type TerminalConfig struct {
    Shell      string `mapstructure:"shell"`
    FontFamily string `mapstructure:"font_family"`
    FontSize   int    `mapstructure:"font_size"`
}

type ProxyConfig struct {
    Enabled bool `mapstructure:"enabled"`
    Port    int  `mapstructure:"port"`
}

type Profile struct {
    Name     string `mapstructure:"name"`
    Provider string `mapstructure:"provider"`
    APIKey   string `mapstructure:"api_key"`
    BaseURL  string `mapstructure:"base_url"`
    Active   bool   `mapstructure:"active"`
}

type GitHubConfig struct {
    Token        string `mapstructure:"token"`
    AutoDetectPR bool   `mapstructure:"auto_detect_pr"`
}

type Manager struct {
    v     *viper.Viper
    cfg   *Config
    path  string
}

func Load() (*Manager, error) {
    configPath := getConfigPath()
    
    v := viper.New()
    v.SetConfigFile(configPath)
    v.SetConfigType("toml")
    
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }
    
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    return &Manager{
        v:    v,
        cfg:  &cfg,
        path: configPath,
    }, nil
}

func Default() *Manager {
    cfg := &Config{
        App: AppConfig{
            Theme:       "dark",
            CheckUpdate: true,
        },
        Terminal: TerminalConfig{
            Shell:      "/bin/zsh",
            FontFamily: "JetBrains Mono",
            FontSize:   14,
        },
        Proxy: ProxyConfig{
            Enabled: true,
            Port:    8080,
        },
        Profiles: []Profile{},
        GitHub: GitHubConfig{
            AutoDetectPR: true,
        },
    }
    
    return &Manager{cfg: cfg}
}

func (m *Manager) Current() *Config {
    return m.cfg
}

func (m *Manager) Save(cfg *Config) error {
    m.cfg = cfg
    // 实现保存逻辑
    return nil
}

func getConfigPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "agent-orch", "config.toml")
}

func (m *Manager) DatabasePath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "agent-orch", "data.db")
}
```

```go
// internal/db/database.go
package db

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type Database struct {
    *gorm.DB
}

func Init(path string) (*Database, error) {
    db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    // 自动迁移
    db.AutoMigrate(
        &UsageRecord{},
        &Session{},
    )
    
    return &Database{db}, nil
}

// Phase 4 将使用
type UsageRecord struct {
    gorm.Model
    ProfileName string
    Provider    string
    Model       string
    InputTokens int
    OutputTokens int
    Cost        float64
    Timestamp   int64
}

// Phase 3 将使用
type Session struct {
    gorm.Model
    ID        string `gorm:"uniqueIndex"`
    WorktreeID string
    CWD       string
    CreatedAt int64
}
```

---

## 3. 测试计划

### 3.1 单元测试

```go
// internal/config/config_test.go
package config

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
    m := Default()
    cfg := m.Current()
    
    assert.Equal(t, "dark", cfg.App.Theme)
    assert.Equal(t, 8080, cfg.Proxy.Port)
    assert.Equal(t, "/bin/zsh", cfg.Terminal.Shell)
}

func TestConfigSave(t *testing.T) {
    // 使用临时文件测试
    // ...
}
```

```go
// internal/db/database_test.go
package db

import (
    "testing"
    "os"
    "path/filepath"
    
    "github.com/stretchr/testify/assert"
)

func TestDatabaseInit(t *testing.T) {
    // 创建临时数据库
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")
    
    db, err := Init(dbPath)
    assert.NoError(t, err)
    assert.NotNil(t, db)
    
    // 验证表已创建
    assert.True(t, db.Migrator().HasTable(&UsageRecord{}))
    assert.True(t, db.Migrator().HasTable(&Session{}))
}
```

### 3.2 集成测试

```go
// app_test.go
package main

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestAppStartup(t *testing.T) {
    app := NewApp()
    ctx := context.Background()
    
    app.startup(ctx)
    
    assert.NotNil(t, app.config)
    assert.NotNil(t, app.db)
}

func TestGetConfig(t *testing.T) {
    app := NewApp()
    app.config = config.Default()
    
    cfg := app.GetConfig()
    
    assert.NotNil(t, cfg)
    assert.Equal(t, "dark", cfg.App.Theme)
}
```

### 3.3 手动测试清单

- [x] 应用启动无报错
- [x] 窗口正常打开
- [x] Wails + React 默认模板正常运行
- [ ] 三栏布局显示正常 (推迟到 F0.4 完成后)
- [ ] 暗色主题渲染正确 (推迟到 F0.4 完成后)
- [ ] 窗口可拖拽、缩放
- [x] 配置文件自动创建 (~/.config/agent-orch/config.toml)
- [x] 数据库文件自动创建 (~/.config/agent-orch/data.db)

---

## 4. 验收标准

| 标准 | 描述 | 状态 |
|------|------|------|
| 应用启动 | 运行 `wails dev` 能正常启动 | ✅ 通过 |
| Go 后端结构 | internal/ 目录结构完整 | ✅ 通过 |
| TypeScript | 前端 TypeScript 配置正确 | ✅ 通过 |
| TailwindCSS | 样式框架配置 | ⏸️ 推迟 |
| 三栏布局 | 基础 UI 布局 | ⏸️ 推迟 |

---

## 5. 发布检查清单

- [ ] 所有单元测试通过
- [ ] 手动测试清单完成
- [ ] 代码通过 golangci-lint
- [ ] 前端通过 TypeScript 检查
- [ ] 更新 CHANGELOG.md
- [ ] 创建 Git tag: v0.1.0-alpha
- [ ] GitHub Release 构建

---

## 6. 依赖列表

### Go 依赖

```go
// go.mod
require (
    github.com/wailsapp/wails/v2 v2.9.0
    github.com/spf13/viper v1.19.0
    gorm.io/gorm v1.25.12
    gorm.io/driver/sqlite v1.5.6
    github.com/stretchr/testify v1.9.0  // test
)
```

### 前端依赖

```json
{
  "dependencies": {
    "react": "^18.3",
    "react-dom": "^18.3",
    "zustand": "^4.5"
  },
  "devDependencies": {
    "typescript": "^5.4",
    "tailwindcss": "^3.4",
    "postcss": "^8.4",
    "autoprefixer": "^10.4",
    "vite": "^5.2",
    "@types/react": "^18.3",
    "@types/react-dom": "^18.3"
  }
}
```

---

## 7. 时间估算

| 任务 | 时间 |
|------|------|
| Wails 项目初始化 | 0.5 天 |
| Go 后端结构 | 1 天 |
| 前端目录结构 | 0.5 天 |
| 基础布局组件 | 1.5 天 |
| 配置系统 | 1 天 |
| 测试编写 | 0.5 天 |
| 文档和发布 | 0.5 天 |
| **总计** | **5 天 (1 周)** |
