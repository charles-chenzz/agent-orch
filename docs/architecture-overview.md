# Agent-Orch 架构总览

## 1. 系统架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Agent-Orch Desktop App                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                    React Frontend (WebView)                       │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌───────────────────────────┐  │  │
│  │  │  Worktree   │  │  Terminal   │  │     API Manager           │  │  │
│  │  │  Manager    │  │  (xterm.js) │  │     (核心差异化)           │  │  │
│  │  └─────────────┘  └─────────────┘  └───────────────────────────┘  │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌───────────────────────────┐  │  │
│  │  │   Diff      │  │   Agent     │  │    Code Editor            │  │  │
│  │  │  Viewer     │  │  Monitor    │  │   (CodeMirror 6)          │  │  │
│  │  └─────────────┘  └─────────────┘  └───────────────────────────┘  │  │
│  │                                                                   │  │
│  │  State: Zustand    Styling: TailwindCSS    Router: -             │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                              │ Wails IPC                                │
│                              ▼                                          │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                      Go Backend (Wails Bindings)                  │  │
│  │                                                                   │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌───────────────────────────┐  │  │
│  │  │  Worktree   │  │  Terminal   │  │      API Proxy            │  │  │
│  │  │  Manager    │  │   Manager   │  │      Server               │  │  │
│  │  │  (go-git)   │  │  (pty/tmux) │  │                           │  │  │
│  │  └─────────────┘  └─────────────┘  └───────────────────────────┘  │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌───────────────────────────┐  │  │
│  │  │   Config    │  │   Agent     │  │      Database             │  │  │
│  │  │   Manager   │  │  Detector   │  │     (SQLite/GORM)         │  │  │
│  │  └─────────────┘  └─────────────┘  └───────────────────────────┘  │  │
│  │                                                                   │  │
│  │  ┌─────────────────────────────────────────────────────────────┐  │  │
│  │  │                    GitHub Integration                       │  │  │
│  │  └─────────────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              │ Unix Socket / HTTP (可选)
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    Terminal Daemon (独立进程 - Phase 4+)                │
│                                                                         │
│  - 拥有所有 PTY 进程                                                    │
│  - Headless xterm.js (Node.js) 做状态追踪                               │
│  - 会话持久化、状态序列化                                                │
│  - 应用重启后会话存活                                                   │
└─────────────────────────────────────────────────────────────────────────┘
```

## 2. 技术栈选型

### 2.1 桌面框架

| 选择 | Wails v2 (Go + Web) |
|------|---------------------|
| 理由 | Go 后端易于编写，打包体积小(15-30MB)，学习曲线低 |
| 替代方案 | Tauri (Rust) - 学习曲线高；Electron - 打包过大(150MB+) |

### 2.2 前端技术栈

| 组件 | 选择 | 理由 |
|------|------|------|
| UI 框架 | React 18 | 生态成熟，Wails 官方推荐 |
| 状态管理 | Zustand | 轻量，API 简洁，TypeScript 友好 |
| 样式方案 | TailwindCSS 3 | 原子化 CSS，快速开发 |
| 终端模拟 | xterm.js 5 | 业界标准，支持 truecolor |
| 代码编辑 | CodeMirror 6 | 轻量可扩展，无 LSP 依赖 |
| Markdown | react-markdown | 成熟，支持 GFM |

### 2.3 后端技术栈

| 组件 | 选择 | 理由 |
|------|------|------|
| Git 操作 | go-git v5 | 纯 Go 实现，无 CGO 依赖 |
| PTY 管理 | creack/pty | Go 成熟库 |
| 数据库 | SQLite + GORM | 轻量，单文件，易备份 |
| 配置管理 | Viper | 支持多格式，热重载 |
| HTTP 代理 | net/http (标准库) | 无额外依赖 |

### 2.4 终端持久化方案

| 阶段 | 方案 | 理由 |
|------|------|------|
| MVP | tmux | 简单可靠，天然支持 alternate screen |
| 生产 | 独立 Daemon + Headless xterm | 应用重启后会话存活 |

## 3. 模块边界与依赖

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend Modules                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐    │
│  │  Worktree    │     │  Terminal    │     │  APIManager  │    │
│  │  Components  │     │  Components  │     │  Components  │    │
│  └──────┬───────┘     └──────┬───────┘     └──────┬───────┘    │
│         │                    │                    │            │
│         └────────────────────┼────────────────────┘            │
│                              │                                 │
│                    ┌─────────▼─────────┐                       │
│                    │    App Store      │                       │
│                    │    (Zustand)      │                       │
│                    └─────────┬─────────┘                       │
│                              │                                 │
│                    ┌─────────▼─────────┐                       │
│                    │   Wails Bindings  │                       │
│                    │   (Auto-generated)│                       │
│                    └─────────┬─────────┘                       │
│                              │ IPC                             │
└──────────────────────────────┼────────────────────────────────┘
                               │
┌──────────────────────────────┼────────────────────────────────┐
│                         Go Backend                             │
├──────────────────────────────┼────────────────────────────────┤
│                    ┌─────────▼─────────┐                       │
│                    │       App         │                       │
│                    │   (Wails Entry)   │                       │
│                    └─────────┬─────────┘                       │
│                              │                                 │
│  ┌───────────────┬───────────┼───────────┬───────────────┐    │
│  │               │           │           │               │    │
│  ▼               ▼           ▼           ▼               ▼    │
│ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐        │
│ │worktree│ │terminal│ │ config │ │ proxy  │ │ agent  │        │
│ │        │ │        │ │        │ │        │ │        │        │
│ └────┬───┘ └────┬───┘ └────┬───┘ └────┬───┘ └────┬───┘        │
│      │          │          │          │          │            │
│      └──────────┴──────────┼──────────┴──────────┘            │
│                            │                                   │
│                  ┌─────────▼─────────┐                         │
│                  │     Database      │                         │
│                  │     (SQLite)      │                         │
│                  └───────────────────┘                         │
└─────────────────────────────────────────────────────────────────┘
```

### 3.1 模块依赖规则

1. **前端模块间**：通过 App Store 通信，不直接依赖
2. **前后端通信**：仅通过 Wails IPC，不直接调用
3. **后端模块间**：通过接口解耦，依赖注入

### 3.2 接口占位示例

```go
// internal/worktree/types.go

// TerminalExecutor 接口 - 由 terminal 模块实现
// Phase 0 使用 Mock，Phase 2 注入真实实现
type TerminalExecutor interface {
    CreateSession(id, cwd string) error
    SendInput(id, data string) error
    Resize(id string, cols, rows uint16) error
    Close(id string) error
}

// Manager 依赖 TerminalExecutor
type Manager struct {
    executor TerminalExecutor // 接口注入
}

// Phase 0: Mock 实现
type MockTerminalExecutor struct{}

func (m *MockTerminalExecutor) CreateSession(id, cwd string) error {
    return nil // 占位
}
```

## 4. 数据流图

### 4.1 终端输入输出流

```
用户输入 "ls -la"
       │
       ▼
┌──────────────┐
│  xterm.js    │  捕获键盘事件
│  onData()    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Wails IPC    │  SendTerminalInput(id, data)
│ Binding      │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Go Backend   │  terminal.Manager.SendInput()
│ PTY Manager  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  PTY/Shell   │  执行命令
│  Process     │
└──────┬───────┘
       │
       ▼ (输出流)
┌──────────────┐
│ Go Backend   │  runtime.EventsEmit(ctx, "output", data)
│ PTY Reader   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Wails Event  │  EventsOn("terminal-output", cb)
│ System       │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  xterm.js    │  term.write(data)
│  render      │
└──────────────┘
```

### 4.2 Worktree 操作流

```
用户点击 "Create Worktree"
       │
       ▼
┌──────────────┐
│ React Modal  │  收集 name, branch
│ Form         │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Wails IPC    │  CreateWorktree(name, branch)
│ Binding      │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Go Backend   │  worktree.Manager.Create()
│              │
└──────┬───────┘
       │
       ├────────────────────┐
       ▼                    ▼
┌──────────────┐    ┌──────────────┐
│   go-git     │    │   Config     │
│   操作       │    │   更新       │
└──────┬───────┘    └──────────────┘
       │
       ▼
┌──────────────┐
│ 返回结果     │  success/error
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ React UI     │  更新列表 / 显示通知
│ 更新         │
└──────────────┘
```

### 4.3 API Proxy 请求流

```
终端中运行 Claude Code
       │
       ▼
┌──────────────┐
│ Claude Code  │  请求 https://api.anthropic.com/v1/messages
│              │
└──────┬───────┘
       │
       ▼ (环境变量 HTTPS_PROXY)
┌──────────────┐
│ API Proxy    │  localhost:8080
│ Server       │
└──────┬───────┘
       │
       ├────────────────────┐
       ▼                    ▼
┌──────────────┐    ┌──────────────┐
│ 请求拦截     │    │  API Key     │
│ / 替换       │    │  注入        │
└──────┬───────┘    └──────────────┘
       │
       ▼
┌──────────────┐
│ 真实 API     │  转发到目标服务器
│ 服务器       │
└──────┬───────┘
       │
       ▼ (响应)
┌──────────────┐
│ 响应记录     │  tokens 使用量 / 成本
│              │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ SQLite       │  存储使用记录
│ Database     │
└──────────────┘
```

## 5. 目录结构

```
agent-orch/
├── build/                          # 构建配置
│   ├── appicon.png                # 应用图标
│   ├── darwin/
│   │   └── Info.plist
│   └── windows/
│       └── wails.exe.manifest
│
├── frontend/                       # React 前端
│   ├── src/
│   │   ├── App.tsx                # 主布局
│   │   ├── main.tsx               # 入口
│   │   │
│   │   ├── components/            # UI 组件
│   │   │   ├── Layout/
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   ├── MainPane.tsx
│   │   │   │   └── RightPanel.tsx
│   │   │   │
│   │   │   ├── Worktree/
│   │   │   │   ├── WorktreeList.tsx
│   │   │   │   ├── WorktreeItem.tsx
│   │   │   │   └── CreateModal.tsx
│   │   │   │
│   │   │   ├── Terminal/
│   │   │   │   ├── Terminal.tsx
│   │   │   │   ├── TerminalTab.tsx
│   │   │   │   └── TerminalPane.tsx
│   │   │   │
│   │   │   ├── Diff/
│   │   │   │   ├── DiffViewer.tsx
│   │   │   │   └── FileTree.tsx
│   │   │   │
│   │   │   ├── APIManager/
│   │   │   │   ├── ProfileList.tsx
│   │   │   │   ├── UsageStats.tsx
│   │   │   │   └── ProfileForm.tsx
│   │   │   │
│   │   │   ├── AgentMonitor/
│   │   │   │   ├── StatusIndicator.tsx
│   │   │   │   └── PRCards.tsx
│   │   │   │
│   │   │   └── Editor/
│   │   │       ├── CodeEditor.tsx
│   │   │       └── MarkdownPreview.tsx
│   │   │
│   │   ├── hooks/
│   │   │   ├── useTerminal.ts
│   │   │   ├── useWorktree.ts
│   │   │   ├── useAPIProxy.ts
│   │   │   └── useConfig.ts
│   │   │
│   │   ├── stores/
│   │   │   ├── worktreeStore.ts
│   │   │   ├── terminalStore.ts
│   │   │   ├── apiProxyStore.ts
│   │   │   └── appStore.ts
│   │   │
│   │   ├── lib/
│   │   │   └── wails.ts           # Wails 工具函数
│   │   │
│   │   └── types/
│   │       └── index.ts
│   │
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   │
│   └── wailsjs/                   # Wails 自动生成
│       └── go/
│           └── main/
│               ├── models.ts
│               └── App.ts
│
├── internal/                       # Go 业务逻辑
│   ├── worktree/
│   │   ├── manager.go
│   │   ├── git.go
│   │   └── types.go
│   │
│   ├── terminal/
│   │   ├── pty.go
│   │   ├── session.go
│   │   ├── tmux.go
│   │   └── types.go
│   │
│   ├── proxy/
│   │   ├── server.go
│   │   ├── interceptor.go
│   │   ├── cert.go
│   │   └── usage.go
│   │
│   ├── config/
│   │   ├── config.go
│   │   └── profile.go
│   │
│   ├── agent/
│   │   ├── detector.go
│   │   └── monitor.go
│   │
│   ├── github/
│   │   └── client.go
│   │
│   └── db/
│       ├── database.go
│       └── models.go
│
├── app.go                          # Wails 应用入口
├── main.go                         # 程序入口
├── wails.json
├── go.mod
└── go.sum
```

## 6. 配置文件格式

```toml
# ~/.config/agent-orch/config.toml

[app]
theme = "dark"
check_updates = true

[terminal]
shell = "/bin/zsh"
font_family = "JetBrains Mono"
font_size = 14

[proxy]
enabled = true
port = 8080

[[profiles]]
name = "official"
provider = "anthropic"
api_key = "sk-ant-..."
base_url = "https://api.anthropic.com"
active = true

[[profiles]]
name = "proxy-cn"
provider = "anthropic"
api_key = "sk-ant-..."
base_url = "https://api.example.com"
active = false

[github]
token = "ghp_..."
auto_detect_pr = true
```

## 7. 开发阶段总览

| Phase | 名称 | 周期 | 核心产出 |
|-------|------|------|----------|
| 0 | Foundation | 1 周 | 项目脚手架 + 基础布局 |
| 1 | Worktree | 2 周 | Worktree CRUD |
| 2 | Terminal MVP | 2 周 | 基础终端 (tmux) |
| 3 | Terminal Stability | 2 周 | TUI 兼容 + 多 Tab |
| 4 | API Manager | 3-4 周 | API 代理 + 使用量追踪 |
| 5 | Agent Monitor | 2 周 | 进程监控 + GitHub 集成 |
| 6 | Code Editor | 3-4 周 | 内置编辑器 |

**总计：约 4-5 个月**
