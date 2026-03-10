
## Arbor 项目分析

### 项目概述
Arbor 是一个基于 Rust + GPUI 构建的**原生桌面应用**，用于管理 Git worktrees、嵌入式终端、diff 查看和 AI 编码代理活动。MIT 开源许可，255 stars，活跃开发中。

### 核心功能模块

1. **Git Worktree 管理**
   - 创建、列表、删除 worktrees
   - 分支安全检查、导航历史
   - 未推送提交检测
   - 最后 git 活动时间戳

2. **嵌入式终端**
   - PTY 终端（支持 truecolor 和 xterm-256color）
   - 每个 worktree 多终端标签
   - 持久化会话（支持应用重启后恢复）
   - 支持 Alacritty、Ghostty 后端

3. **Diff 和文件变更**
   - 并排 diff 显示
   - 变更文件列表
   - 文件树浏览

4. **AI 代理活动监控**
   - 检测 Claude Code、Codex、OpenCode
   - 工作/等待状态指示器
   - WebSocket 实时更新

5. **MCP 服务器**
   - 独立的 `arbor-mcp` 二进制
   - 提供 repos、worktrees、terminals 等工具
   - 支持远程认证

6. **远程 Outposts**
   - SSH 远程 worktree 管理
   - Mosh 支持
   - 主机状态追踪

7. **GitHub 集成**
   - 自动 PR 检测和链接
   - Git 操作（commit、push）

8. **UI 和配置**
   - 三栏布局（仓库、终端、变更）
   - 25+ 主题
   - TOML 配置热重载

### 技术栈

| 组件 | 技术 |
|------|------|
| UI 框架 | GPUI（Zed 团队开源的 GPU 加速 UI 框架） |
| 后端 | Rust (nightly-2025-11-30) |
| HTTP 守护进程 | arbor-httpd |
| Web UI | TypeScript + xterm.js |
| 终端 | PTY + 多后端支持 |
| 包管理 | Cargo workspace |

### 项目结构（crates/）
- `arbor-core` - 核心逻辑（worktree 原语、变更检测、agent hooks）
- `arbor-gui` - GPUI 桌面应用
- `arbor-httpd` - 远程 HTTP 守护进程
- `arbor-mcp` - MCP 服务器
- `arbor-daemon-client` - 类型化客户端和 API DTOs
- `arbor-web-ui` - TypeScript 资源

---

## 实现方案对比

### 方案 A：Fork 并基于此修改

**优点：**
- 立即可用，234 commits 的成熟代码
- GPUI 框架学习曲线高，直接复用省时
- 已有完整的基础设施（CI、打包、Homebrew）
- 活跃维护，可贡献上游

**缺点：**
- 受限于 GPUI + Rust 技术栈
- 修改需要了解 GPUI 内部机制
- 依赖 Rust nightly

**适合：** 想要快速迭代、接受 Rust 技术栈

---

### 方案 B：从头实现

**优点：**
- 可选择熟悉的技术栈（如 Electron + React、Tauri + React）
- 更灵活的架构设计
- 不依赖 GPUI 的复杂性

**缺点：**
- 工作量大（估计 2-3 个月 MVP）
- 需要自己实现 PTY 终端、Git 操作、会话持久化
- 原生性能和体验可能不如 GPUI

**需要实现的核心组件：**
1. Git worktree 管理层（可用 isomorphic-git 或 simple-git）
2. 终端模拟器（xterm.js）
3. PTY 进程管理（node-pty）
4. 文件 diff 渲染
5. 进程/agent 监控
6. WebSocket 实时通信
7. 配置系统
8. 打包和分发

---

---

## 最终实现方案：AI Agent Worktree Manager

### 项目定位

**核心差异化**：在 Arbor 基础上增加 **LLM API 管理中心** + **稳定的 TUI 终端体验**

**目标用户痛点**：
1. 多个 AI 编码工具（Claude Code、Codex、Cursor 等）需要分别配置 API
2. 切换 API Key / 中转站很麻烦（目前用 cc-switch 等工具）
3. 无法统一查看 API 使用量和成本
4. 无法在多 worktree 中并行运行不同配置的 Agent
5. 终端 Tab 切换时 TUI 应用（Claude Code、vim）显示异常

---

### 技术架构

#### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                     Wails Desktop App                           │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  React Frontend (WebView)                                 │  │
│  │  ┌─────────┐  ┌─────────────┐  ┌───────────────────────┐  │  │
│  │  │Worktree │  │ xterm.js    │  │ API Manager           │  │  │
│  │  │Manager  │  │ + Serialize │  │ (核心差异化)           │  │  │
│  │  └─────────┘  └─────────────┘  └───────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────┘  │
│                           │ IPC                                 │
│  ┌────────────────────────▼──────────────────────────────────┐  │
│  │  Go Backend (Wails 绑定)                                  │  │
│  │  - Worktree 操作 (go-git)                                 │  │
│  │  - API Proxy (HTTP 代理)                                  │  │
│  │  - 配置管理 (TOML)                                        │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            │ Unix Socket / HTTP
┌───────────────────────────▼───────────────────────────────────┐
│                   Terminal Daemon (独立进程)                   │
│  - 拥有所有 PTY 进程                                           │
│  - Headless xterm.js (Node.js) 做状态追踪                      │
│  - 会话持久化、状态序列化                                       │
│  - 应用重启后会话存活                                          │
│  - 参考 Superset 的实现                                        │
└───────────────────────────────────────────────────────────────┘
```

#### 技术栈选择

| 组件 | 技术 | 理由 |
|------|------|------|
| 桌面框架 | **Wails v2** (Go + Web) | Go 后端，轻量，打包小，学习曲线低 |
| 前端 | **React + TailwindCSS** | 生态成熟，Wails 官方推荐 |
| 终端 | **xterm.js + xterm-addon-serialize** | 业界标准，支持状态序列化 |
| Git 操作 | **go-git** | 纯 Go 实现，无 CGO 依赖 |
| PTY 管理 | **creack/pty** | Go 成熟库 |
| API 代理 | **自建 HTTP Proxy** | 拦截/转发 LLM API 请求 |
| 数据库 | **SQLite + GORM** | 轻量，单文件，易备份 |
| 终端持久化 | **tmux (MVP) → Daemon (生产)** | 分阶段实现 |

---

### UI 框架对比

```
┌─────────────────────────────────────────────────────────────────┐
│ 方案 A: WebView (Wails/Tauri) ← 选择                            │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  系统 WebView (macOS: WKWebView, Windows: Edge WebView2)│   │
│  │  ↓                                                       │   │
│  │  渲染 HTML/CSS/JS (React 应用)                           │   │
│  └─────────────────────────────────────────────────────────┘   │
│  优点：打包小 (10-30MB)，xterm.js 直接用                       │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 方案 B: Electron                                               │
│  内置 Chromium (~150MB)，打包大                                │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ 方案 C: GPUI (Zed/Arbor)                                       │
│  GPU 直接渲染，性能极佳，但需要自己实现 terminal                │
└─────────────────────────────────────────────────────────────────┘
```

| 特性 | Wails (Go) | Tauri (Rust) | Electron | GPUI |
|------|------------|--------------|----------|------|
| **打包体积** | 15-30 MB | 3-10 MB | 150+ MB | 5-10 MB |
| **xterm.js** | ✅ 直接用 | ✅ 直接用 | ✅ 直接用 | ❌ 需重写 |
| **学习曲线** | 低 | 高 | 低 | 很高 |
| **Git 库** | go-git 成熟 | git2-rs 成熟 | isomorphic-git | 需绑定 |

### Wails 的角色

**Wails 只是一个"壳" + IPC 桥梁**，不承担业务逻辑：

```
┌─────────────────────────────────────────────────────────────┐
│                    Wails 的职责                              │
│  1. 打包：把 Go + Web 资源打包成 .app/.exe                  │
│  2. IPC：提供 Go ↔ JavaScript 通信通道                       │
│  3. 窗口：创建原生窗口（基于 WebView）                        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    你需要自己实现的                           │
│  Go 后端：Git 操作、PTY 管理、API 代理、配置管理              │
│  React 前端：所有 UI 组件、xterm.js 集成、状态管理            │
└─────────────────────────────────────────────────────────────┘
```

---

### Wails + React + xterm.js 架构解释

#### 各组件职责

| 组件 | 角色 | 类比 |
|------|------|------|
| **Wails** | 桌面框架，负责把 Go + Web 打包成桌面应用 | 像一个"壳"，装着你的应用 |
| **Go 后端** | 所有业务逻辑、系统调用 | 后端 API 服务 |
| **React** | UI 渲染，处理用户交互 | 前端页面 |
| **xterm.js** | 终端模拟器，一个 React 组件 | 网页版的终端界面 |

#### 数据流

```
用户在终端输入 "ls -la"
        │
        ▼
   xterm.js 捕获输入
        │
        ▼
   通过 Wails IPC 发送给 Go
        │
        ▼
   Go 创建 PTY 进程执行命令
        │
        ▼
   Go 把输出流式返回给前端
        │
        ▼
   xterm.js 渲染输出到屏幕
```

#### xterm.js 嵌入方式

xterm.js 是一个 **npm 包**，作为 React 组件使用：

```tsx
// Terminal.tsx (React 组件)
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { useEffect, useRef } from 'react';

function TerminalComponent() {
  const terminalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // 1. 创建 xterm.js 实例
    const term = new Terminal({
      theme: { background: '#1e1e1e' }
    });

    // 2. 挂载到 DOM
    const addon = new FitAddon();
    term.loadAddon(addon);
    term.open(terminalRef.current!);
    addon.fit();

    // 3. 用户输入 → 发送给 Go 后端
    term.onData((data) => {
      // Wails 生成的绑定函数
      window.go.main.App.SendInput(data);
    });

    // 4. 接收 Go 后端的输出
    window.runtime.EventsOn('terminal-output', (output) => {
      term.write(output);
    });
  }, []);

  return <div ref={terminalRef} className="h-full" />;
}
```

---

### 终端稳定性方案

#### 问题分析

TUI 应用（Claude Code、vim、htop）在 Tab 切换时可能出现：
- 黑屏
- 输入框截断/重复
- 显示错乱

**根因**：
1. **Alternate Screen Buffer** - TUI 使用备用屏幕缓冲区，切换时需要正确恢复
2. **终端状态序列化** - 隐藏/显示 DOM 时需要保存/恢复光标、滚动区域等
3. **Resize 同步** - Tab 切换触发尺寸变化，PTY 需要 SIGWINCH 信号
4. **输入模式** - Bracketed Paste、Mouse Tracking 等需要正确恢复

#### 解决方案：分阶段实现

**Phase 1: tmux 包装（MVP，简单可靠）**

```go
// 不是直接启动 shell，而是创建/附加 tmux session
func createTerminal(worktree string) {
    sessionName := "agentflow-" + worktree

    cmd := exec.Command("tmux", "new-session",
        "-A",           // 如果不存在则创建
        "-s", sessionName,
        "-c", worktreePath,
    )

    pty.Start(cmd)
}
```

优点：
- 天然支持 alternate screen
- 会话持久化
- Tab 切换状态不丢失
- 兼容所有 TUI 应用

缺点：
- 额外依赖
- 不支持 Windows（需要 zellij 或 conpty）

**Phase 2: 独立 Daemon + Headless xterm（生产级）**

参考 Superset 的实现：
- 独立 Node.js 守护进程
- 每个 PTY 会话维护 headless xterm.js
- 支持状态序列化和恢复
- 应用重启后会话存活

#### Tab 切换处理

```typescript
// React 组件：终端容器
function TerminalPane({ sessionId }: { sessionId: string }) {
  const xtermRef = useRef<Terminal | null>(null);
  const isVisible = useIsVisible(); // 检测当前 tab 是否可见

  // Tab 切换时的处理
  useEffect(() => {
    if (xtermRef.current) {
      if (isVisible) {
        // 重新聚焦时：刷新尺寸 + 恢复光标
        fitAddon.fit();
        xtermRef.current.focus();
      }
    }
  }, [isVisible]);

  return <div ref={terminalRef} className="h-full w-full" />;
}
```

#### Resize 闪烁修复

```typescript
// 防抖 resize
const debouncedFit = debounce(() => {
  if (xtermRef.current && fitAddon) {
    fitAddon.fit();
    resizePty(sessionId, xtermRef.current.cols, xtermRef.current.rows);
  }
}, 50);

// 使用 ResizeObserver
useEffect(() => {
  const observer = new ResizeObserver(() => {
    debouncedFit();
  });

  if (containerRef.current) {
    observer.observe(containerRef.current);
  }

  return () => observer.disconnect();
}, []);
```

---

### 核心功能模块

#### 1. Git Worktree 管理
- 创建、列表、删除 worktrees
- 分支安全检查、导航历史
- 未推送提交检测
- 最后 git 活动时间戳

#### 2. 嵌入式终端
- PTY 终端（支持 truecolor 和 xterm-256color）
- 每个 worktree 多终端标签
- 持久化会话（tmux/Daemon）
- TUI 应用兼容（Claude Code、Codex、vim）

#### 3. Diff 和文件变更
- 并排 diff 显示
- 变更文件列表
- 文件树浏览

#### 4. AI Agent 监控
- 检测运行中的 Claude Code、Codex、OpenCode
- 工作/等待状态指示器
- 实时更新

#### 5. API 管理中心 ⭐ 核心差异化
- API Key 管理界面
- 多 Profile 支持（官方/中转站/自建）
- **API Proxy 代理层**
  - 拦截 Claude/GPT/Gemini API 请求
  - 自动注入/替换 API Key
  - 请求日志记录
- 环境变量自动注入
- 使用量追踪（tokens、成本估算）
- 一键切换 Profile

#### 6. GitHub 集成
- 自动 PR 检测和链接
- Git 操作（commit、push）

#### 7. UI 和配置
- 三栏布局（仓库、终端、变更）
- 多主题支持
- TOML 配置热重载

---

### 项目文件结构

```
agent-orch/
├── build/                      # 构建配置
│   ├── appicon.png            # 应用图标
│   ├── darwin/                # macOS 特定
│   │   └── Info.plist
│   └── windows/               # Windows 特定
│       └── wails.exe.manifest
│
├── frontend/                   # React 前端
│   ├── src/
│   │   ├── App.tsx            # 主布局
│   │   ├── main.tsx           # 入口
│   │   │
│   │   ├── components/        # UI 组件
│   │   │   ├── Layout/
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   └── MainPane.tsx
│   │   │   │
│   │   │   ├── Terminal/      # 终端模块
│   │   │   │   ├── Terminal.tsx       # xterm.js 封装
│   │   │   │   ├── TerminalTab.tsx    # Tab 管理
│   │   │   │   └── TerminalPane.tsx   # 终端容器
│   │   │   │
│   │   │   ├── Worktree/      # Worktree 模块
│   │   │   │   ├── WorktreeList.tsx
│   │   │   │   ├── WorktreeItem.tsx
│   │   │   │   └── CreateModal.tsx
│   │   │   │
│   │   │   ├── Diff/          # Diff 查看器
│   │   │   │   ├── DiffViewer.tsx
│   │   │   │   └── FileTree.tsx
│   │   │   │
│   │   │   ├── APIManager/    # API 管理中心 ⭐
│   │   │   │   ├── ProfileList.tsx
│   │   │   │   ├── UsageStats.tsx
│   │   │   │   └── ProfileForm.tsx
│   │   │   │
│   │   │   └── Editor/        # 代码编辑器 (Phase 6)
│   │   │       ├── CodeEditor.tsx
│   │   │       └── MarkdownPreview.tsx
│   │   │
│   │   ├── hooks/             # 自定义 Hooks
│   │   │   ├── useTerminal.ts
│   │   │   ├── useWorktree.ts
│   │   │   └── useAPIProxy.ts
│   │   │
│   │   ├── stores/            # 状态管理
│   │   │   └── appStore.ts
│   │   │
│   │   ├── lib/               # 工具函数
│   │   │   └── wails.ts
│   │   │
│   │   └── types/             # TypeScript 类型
│   │       └── index.ts
│   │
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   └── wailsjs/               # Wails 自动生成
│       └── go/
│           └── main/
│               ├── models.ts  # Go struct → TS types
│               └── App.ts     # Go 函数 → TS 函数
│
├── internal/                   # Go 业务逻辑
│   ├── worktree/
│   │   ├── manager.go         # Worktree 增删查改
│   │   ├── git.go             # Git 操作
│   │   └── types.go
│   │
│   ├── terminal/
│   │   ├── pty.go             # PTY 进程管理
│   │   ├── session.go         # 会话管理
│   │   ├── tmux.go            # tmux 集成 (macOS/Linux)
│   │   └── types.go
│   │
│   ├── proxy/
│   │   ├── server.go          # HTTP 代理服务器
│   │   ├── interceptor.go     # 请求拦截
│   │   ├── cert.go            # 证书管理
│   │   └── usage.go           # 使用量统计
│   │
│   ├── config/
│   │   ├── config.go          # 配置加载
│   │   └── profile.go         # API Profile 管理
│   │
│   └── agent/
│       ├── detector.go        # Agent 进程检测
│       └── monitor.go         # Agent 状态监控
│
├── app.go                      # Wails 应用 + 绑定函数
├── main.go                     # 程序入口
├── wails.json
├── go.mod
└── go.sum
```

---

### 核心依赖

#### Go 后端

```go
// go.mod
require (
    github.com/wailsapp/wails/v2 v2.9.0
    github.com/go-git/go-git/v5 v5.12.0
    github.com/creack/pty v1.1.21
    github.com/google/go-github/v62 v62.0.0
    github.com/spf13/viper v1.19.0      // 配置管理
    gorm.io/gorm v1.25.12               // 数据库
    gorm.io/driver/sqlite v1.5.6
)
```

#### 前端

```json
{
  "dependencies": {
    "react": "^18.3",
    "react-dom": "^18.3",
    "xterm": "^5.3",
    "xterm-addon-fit": "^0.8",
    "xterm-addon-serialize": "^0.13",
    "xterm-addon-web-links": "^0.9",
    "@codemirror/lang-javascript": "^6.2",
    "@codemirror/lang-python": "^6.1",
    "@codemirror/lang-rust": "^6.0",
    "@codemirror/lang-go": "^6.0",
    "zustand": "^4.5",
    "react-markdown": "^9.0",
    "tailwindcss": "^3.4"
  }
}
```

---

### 关键实现示例

#### app.go - Wails 绑定入口

```go
package main

import (
    "context"
    "agent-orch/internal/worktree"
    "agent-orch/internal/terminal"
    "agent-orch/internal/proxy"
)

type App struct {
    ctx       context.Context
    worktree  *worktree.Manager
    terminal  *terminal.Manager
    proxy     *proxy.Server
}

// === Worktree 管理 ===
func (a *App) CreateWorktree(name, branch string) error {
    return a.worktree.Create(name, branch)
}

func (a *App) ListWorktrees() ([]worktree.Worktree, error) {
    return a.worktree.List()
}

func (a *App) DeleteWorktree(name string) error {
    return a.worktree.Delete(name)
}

// === Terminal 管理 ===
func (a *App) CreateTerminal(id, cwd string) error {
    return a.terminal.CreateSession(id, cwd)
}

func (a *App) SendTerminalInput(id, data string) error {
    return a.terminal.SendInput(id, data)
}

func (a *App) ResizeTerminal(id string, cols, rows uint16) error {
    return a.terminal.Resize(id, cols, rows)
}

// === API Proxy ===
func (a *App) StartProxy(port int) error {
    return a.proxy.Start(port)
}

func (a *App) GetUsageStats() (*proxy.UsageStats, error) {
    return a.proxy.GetStats()
}

func main() {
    app := &App{}

    wails.Run(&options.App{
        Title:  "Agent Orchestrator",
        Width:  1200,
        Height: 800,
        Bind: []interface{}{
            app,
        },
    })
}
```

#### frontend/src/hooks/useTerminal.ts

```typescript
import { useEffect, useRef, useCallback } from 'react'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { CreateTerminal, SendTerminalInput, ResizeTerminal } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

export function useTerminal(sessionId: string, containerRef: React.RefObject<HTMLDivElement>) {
    const terminalRef = useRef<Terminal | null>(null)
    const fitAddonRef = useRef<FitAddon | null>(null)

    useEffect(() => {
        if (!containerRef.current) return

        const term = new Terminal({
            theme: { background: '#1e1e1e' },
            fontSize: 14,
            fontFamily: 'JetBrains Mono, monospace',
        })

        const fitAddon = new FitAddon()
        term.loadAddon(fitAddon)
        term.open(containerRef.current)
        fitAddon.fit()

        // 用户输入 → Go 后端
        term.onData((data) => {
            SendTerminalInput(sessionId, data)
        })

        // 接收 Go 后端输出
        EventsOn(`terminal-output-${sessionId}`, (output: string) => {
            term.write(output)
        })

        terminalRef.current = term
        fitAddonRef.current = fitAddon

        // 创建 PTY 会话
        CreateTerminal(sessionId, process.cwd())

        return () => {
            EventsOff(`terminal-output-${sessionId}`)
            term.dispose()
        }
    }, [sessionId])

    // Resize 处理
    const handleResize = useCallback(() => {
        if (terminalRef.current && fitAddonRef.current) {
            fitAddonRef.current.fit()
            ResizeTerminal(
                sessionId,
                terminalRef.current.cols,
                terminalRef.current.rows
            )
        }
    }, [sessionId])

    return { terminal: terminalRef, handleResize }
}
```

#### internal/terminal/pty.go

```go
package terminal

import (
    "os"
    "os/exec"
    "sync"

    "github.com/creack/pty"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Manager struct {
    sessions map[string]*Session
    mu       sync.RWMutex
    ctx      context.Context
}

type Session struct {
    ID      string
    PTY     *os.File
    Cmd     *exec.Cmd
    CWD     string
}

func (m *Manager) CreateSession(id, cwd string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 创建 shell 命令
    shell := os.Getenv("SHELL")
    if shell == "" {
        shell = "/bin/bash"
    }

    cmd := exec.Command(shell)
    cmd.Dir = cwd

    // 启动 PTY
    ptyFile, err := pty.Start(cmd)
    if err != nil {
        return err
    }

    session := &Session{
        ID:  id,
        PTY: ptyFile,
        Cmd: cmd,
        CWD: cwd,
    }

    m.sessions[id] = session

    // 读取输出并发送到前端
    go m.readOutput(session)

    return nil
}

func (m *Manager) readOutput(session *Session) {
    buf := make([]byte, 1024)
    for {
        n, err := session.PTY.Read(buf)
        if err != nil {
            return
        }

        // 通过 Wails 事件发送到前端
        runtime.EventsEmit(m.ctx, "terminal-output-"+session.ID, string(buf[:n]))
    }
}

func (m *Manager) SendInput(id, data string) error {
    m.mu.RLock()
    session, ok := m.sessions[id]
    m.mu.RUnlock()

    if !ok {
        return fmt.Errorf("session not found: %s", id)
    }

    _, err := session.PTY.Write([]byte(data))
    return err
}

func (m *Manager) Resize(id string, cols, rows uint16) error {
    m.mu.RLock()
    session, ok := m.sessions[id]
    m.mu.RUnlock()

    if !ok {
        return fmt.Errorf("session not found: %s", id)
    }

    return pty.Setsize(session.PTY, &pty.Winsize{
        Cols: cols,
        Rows: rows,
    })
}
```

---

### 开发路线图

#### Phase 1: 基础框架 (2-3 周)
- [ ] Wails 项目初始化
- [ ] Git worktree 增删查改
- [ ] 基础三栏 UI 布局
- [ ] 配置文件管理

#### Phase 2: 终端集成 - tmux 方案 (2 周)
- [ ] tmux session 管理
- [ ] xterm.js 基础集成
- [ ] 输入输出流
- [ ] Tab 切换基础测试

#### Phase 3: API 管理中心 (3-4 周) ⭐ 核心差异化
- [ ] API Key 管理界面
- [ ] 多 Profile 支持
- [ ] API Proxy 代理层
- [ ] 使用量追踪

#### Phase 4: 终端稳定性打磨 (2-3 周)
- [ ] TUI 应用测试（Claude Code、Codex、vim、htop）
- [ ] Tab 切换优化
- [ ] Resize 闪烁修复
- [ ] 会话持久化（可选升级到 Daemon）

#### Phase 5: Agent 监控 + 增值功能 (2-3 周)
- [ ] 进程检测
- [ ] GitHub PR 集成
- [ ] Diff 查看器
- [ ] 打包发布

#### Phase 6: 内置代码编辑器 (3-4 周)
- [ ] CodeMirror 6 集成
- [ ] 文件打开/保存
- [ ] 语法高亮（40+ 语言）
- [ ] Markdown 实时预览
- [ ] 简单代码导航（正则匹配）
- [ ] Tree-sitter 后台索引（可选）
- [ ] Stack Graphs 跨文件跳转（可选）

**预计总时间：4-5 个月完成完整功能**

---

### 内置代码编辑器方案

#### 需求
1. 打开文件（不依赖外部编辑器如 Zed/JetBrains）
2. Markdown 预览支持
3. 编程语言语法高亮（Node.js、TypeScript、Go、Rust 等）
4. 代码导航（跳转到定义）- **不使用 LSP**

#### 为什么不用 LSP

**LSP 痛点**：
- 启动需要 1-3 秒加载索引
- 每个项目首次打开都要等待
- 内存占用高（每个语言服务器 100MB+）
- 配置复杂，需要用户安装

**目标**：开箱即用，打开文件 <10ms 即可编辑和跳转

#### 技术选型

| 组件 | 选择 | 理由 |
|------|------|------|
| 代码编辑器 | **CodeMirror 6** | 轻量（~100KB）、可扩展、无 LSP 依赖 |
| Markdown 预览 | **react-markdown** + remark-gfm | 成熟、支持 GFM |
| 语法高亮 | CodeMirror 内置 | 40+ 语言，0 延迟 |
| 代码导航 | **正则 + Tree-sitter** | 分层实现，立即可用 |

#### 代码导航分层方案

```
┌─────────────────────────────────────────────────────────────┐
│                    Code Navigation Layers                   │
├─────────────────────────────────────────────────────────────┤
│ Layer 1: 即时响应（打开即用）<10ms                          │
│  - 正则符号匹配                                             │
│  - 当前文件跳转                                             │
│  - 无需任何初始化                                           │
├─────────────────────────────────────────────────────────────┤
│ Layer 2: 后台索引（异步）1-5秒后可用                        │
│  - Tree-sitter 解析：10-50ms/文件                           │
│  - 后台构建符号索引                                         │
│  - 不阻塞 UI                                                │
├─────────────────────────────────────────────────────────────┤
│ Layer 3: 高级导航（可选）索引完成后可用                     │
│  - Stack Graphs：跨文件引用追踪                             │
│  - GitHub 开源技术                                          │
└─────────────────────────────────────────────────────────────┘
```

#### 性能对比

| 方案 | 首次打开文件 | 后续打开 | 增量编辑 | 跨文件跳转 |
|------|------------|---------|---------|-----------|
| **LSP** | 1-3 秒 | 100-300ms | 快 | ✅ 支持 |
| **正则匹配** | <5ms | <5ms | N/A | ❌ 不支持 |
| **Tree-sitter** | 10-50ms | 10-50ms | <5ms | ❌ 单文件 |
| **Stack Graphs** | 1-5秒（后台） | 快 | 快 | ✅ 支持 |

#### 实现架构

```
┌─────────────────────────────────────────────────────────────────┐
│                     Wails Desktop App                           │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  React Frontend (WebView)                                 │  │
│  │  ┌─────────┐  ┌─────────────┐  ┌───────────────────────┐  │  │
│  │  │Worktree │  │ xterm.js    │  │ API Manager           │  │  │
│  │  │Manager  │  │ Terminal    │  │                       │  │  │
│  │  └─────────┘  └─────────────┘  └───────────────────────┘  │  │
│  │  ┌─────────────────────────────────────────────────────┐  │  │
│  │  │ Code Editor (CodeMirror 6)                          │  │  │
│  │  │ • Markdown 实时预览                                  │  │  │
│  │  │ • 语法高亮 (40+ 语言，0 延迟)                        │  │  │
│  │  │ • 简单跳转 (正则，<5ms)                              │  │  │
│  │  │ • 精确跳转 (Tree-sitter，后台)                       │  │  │
│  │  └─────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────┘  │
│                           │ IPC                                 │
│  ┌────────────────────────▼──────────────────────────────────┐  │
│  │  Go Backend                                              │  │
│  │  - gotreesitter (代码解析，纯 Go，无 CGO)                │  │
│  │  - Symbol Index (符号索引)                               │  │
│  │  - Stack Graphs (可选，跨文件引用)                       │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

#### MVP 实现（正则匹配）

```go
type QuickNavigator struct {
    patterns map[Language]*LanguagePatterns
}

type LanguagePatterns struct {
    FunctionDef *regexp.Regexp // Go: func (\w+)\(
    VarDef      *regexp.Regexp // Go: (var|const) (\w+)
    TypeDef     *regexp.Regexp // Go: type (\w+)
    Import      *regexp.Regexp // Go: import.*"(.+)"
}

func (n *QuickNavigator) FindDefinition(symbol string, content string, lang Language) []Location {
    patterns := n.patterns[lang]
    // 正则匹配，<5ms 返回
    matches := patterns.FindAll(symbol, content)
    return matches
}
```

#### 增强实现（Tree-sitter）

```go
// 使用 gotreesitter（纯 Go，比 CGo 快 90 倍）
import "github.com/odvcencio/gotreesitter"

func (e *Editor) startIndexing(file string) {
    go func() {
        // 后台异步解析
        parser := gotreesitter.NewParser()
        parser.SetLanguage(gotreesitter.Go)

        tree := parser.Parse([]byte(content))

        // 提取符号定义
        symbols := e.extractSymbols(tree)

        // 更新索引（不阻塞 UI）
        e.symbolIndex.Update(file, symbols)
    }()
}
```

#### 支持的语言

**Phase 6 MVP**（正则匹配）：
- Go
- TypeScript/JavaScript
- Rust
- Python

**Phase 6 增强**（Tree-sitter）：
- 20+ 语言（使用现有 grammar）

#### Markdown 支持

```tsx
// Markdown 编辑器组件
function MarkdownEditor({ file }: { file: string }) {
  const [content, setContent] = useState('');
  const [preview, setPreview] = useState(false);

  return (
    <div className="flex h-full">
      {/* 编辑区 */}
      <CodeMirror
        value={content}
        extensions={[markdown(), gfm()]}
        onChange={setContent}
      />

      {/* 预览区 */}
      {preview && (
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          components={{
            code: SyntaxHighlighter  // 代码块高亮
          }}
        >
          {content}
        </ReactMarkdown>
      )}
    </div>
  );
}
```

---

### 简历亮点设计

#### 技术深度展示
1. **系统级编程**：PTY 进程管理、信号处理
2. **网络代理**：HTTP/HTTPS 代理实现、证书管理
3. **并发模型**：Go goroutine + channel 管理多个终端会话
4. **跨平台**：macOS/Windows/Linux 打包和适配
5. **编译器/解析器**：AST 解析、Tree-sitter 集成、符号索引
6. **编辑器架构**：CodeMirror 6 扩展开发、实时预览

---

### 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| tmux 不支持 Windows | 中 | Windows 用户使用 zellij 或 conpty |
| TUI 应用兼容性问题 | 高 | 充分测试主流 CLI agent |
| Daemon 复杂度高 | 中 | Phase 1 用 tmux，Phase 2 再实现 |
| 前端经验不足 | 中 | 使用成熟 UI 库（shadcn/ui） |

---

### 获取用户策略

1. **解决真实痛点**：API 管理是当前空白，cc-switch 用户是天然目标
2. **开源 + 文档**：详细的 setup guide，支持常见中转站配置
3. **社区运营**：
   - 在 V2EX、Twitter、Reddit 发布
   - 写技术博客介绍实现细节
   - 录制 demo 视频
4. **持续迭代**：根据用户反馈快速更新

---

### 参考资料

- [Arbor GitHub](https://github.com/penso/arbor) - 原始项目参考
- [Superset Terminal Daemon Deep Dive](https://superset.sh/blog/terminal-daemon-deep-dive) - 终端持久化方案
- [Wails Documentation](https://wails.io/docs/introduction) - 桌面框架文档
- [xterm.js](https://xtermjs.org/) - 终端模拟器
- [go-git](https://github.com/go-git/go-git) - Git 操作库
- [gotreesitter](https://github.com/odvcencio/gotreesitter) - 纯 Go Tree-sitter，无 CGO
- [CodeMirror 6](https://codemirror.net/) - 代码编辑器
- [react-markdown](https://github.com/remarkjs/react-markdown) - Markdown 预览