# Phase 2: Terminal MVP

> **周期**：2 周
> **目标**：实现基础终端功能（tmux 方案）
> **依赖**：Phase 0 (Foundation), Phase 1 (Worktree)
> **交付**：v0.3.0-alpha

---

## 1. Feature List

### 1.1 Go 后端 - PTY 管理

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F2.1 | terminal.Manager 结构 | P0 |
| F2.2 | CreateSession(id, cwd string) | P0 |
| F2.3 | SendInput(id, data string) | P0 |
| F2.4 | Resize(id, cols, rows) | P0 |
| F2.5 | DetachSession(id string) | P0 |
| F2.6 | 输出流事件发送（统一事件协议） | P0 |

### 1.2 Go 后端 - tmux 集成

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F2.7 | tmux 可用性检测 | P0 |
| F2.8 | tmux session 创建/附加 | P0 |
| F2.9 | tmux session 列表 | P1 |
| F2.10 | tmux session 销毁（显式 Destroy） | P0 |
| F2.20 | tmux 保活策略（UI 关闭仅 detach） | P0 |

### 1.3 前端 - xterm.js 集成

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F2.11 | xterm.js 初始化 | P0 |
| F2.12 | FitAddon 集成 | P0 |
| F2.13 | 输入事件处理 | P0 |
| F2.14 | 输出渲染 | P0 |
| F2.15 | 基础主题配置 | P0 |

### 1.4 前端 - 终端组件

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F2.16 | Terminal.tsx 组件 | P0 |
| F2.17 | useTerminal hook | P0 |
| F2.18 | terminalStore | P0 |
| F2.19 | 单终端 Tab UI | P0 |
| F2.21 | 会话重连与状态同步 UI | P1 |

### 1.5 会话状态与协议

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F2.22 | 会话状态机（creating/running/detached/exited/destroyed） | P0 |
| F2.23 | 统一事件协议（terminal:output/state/error/exit） | P0 |

---

## 2. 实现细节

### 2.1 Go 后端 - 数据结构

```go
// internal/terminal/types.go
package terminal

import (
    "os"
    "os/exec"
    "sync"
    "time"
)

type Session struct {
    ID          string
    WorktreeID  string
    CWD         string
    PTY         *os.File
    Cmd         *exec.Cmd
    TmuxSession string    // tmux session 名称
    CreatedAt   time.Time
}

type Manager struct {
    sessions map[string]*Session
    mu       sync.RWMutex
    ctx      context.Context
    tmuxPath string  // tmux 可执行文件路径
    hasTmux  bool    // tmux 是否可用
}

type TerminalConfig struct {
    Shell      string
    FontFamily string
    FontSize   int
    Theme      TerminalTheme
}

type TerminalTheme struct {
    Background string
    Foreground string
    Cursor     string
    Selection  string
}
```

### 2.2 Go 后端 - Manager 实现

```go
// internal/terminal/pty.go
package terminal

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "sync"
    
    "github.com/creack/pty"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

func NewManager(ctx context.Context) *Manager {
    tmuxPath, hasTmux := exec.LookPath("tmux")
    
    return &Manager{
        sessions: make(map[string]*Session),
        ctx:      ctx,
        tmuxPath: tmuxPath,
        hasTmux:  hasTmux,
    }
}

// CreateSession 创建新的终端会话
func (m *Manager) CreateSession(id, worktreeID, cwd string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // 检查是否已存在
    if _, exists := m.sessions[id]; exists {
        return fmt.Errorf("session already exists: %s", id)
    }
    
    var ptyFile *os.File
    var cmd *exec.Cmd
    var tmuxSession string
    
    if m.hasTmux {
        // 使用 tmux
        tmuxSession = fmt.Sprintf("agent-orch-%s", id)
        cmd = exec.Command(m.tmuxPath, "new-session",
            "-A",              // 如果不存在则创建
            "-s", tmuxSession, // session 名称
            "-c", cwd,         // 工作目录
        )
    } else {
        // 直接使用 shell
        shell := os.Getenv("SHELL")
        if shell == "" {
            shell = "/bin/bash"
        }
        cmd = exec.Command(shell)
        cmd.Dir = cwd
    }
    
    // 启动 PTY
    var err error
    ptyFile, err = pty.Start(cmd)
    if err != nil {
        return fmt.Errorf("failed to start pty: %w", err)
    }
    
    session := &Session{
        ID:          id,
        WorktreeID:  worktreeID,
        CWD:         cwd,
        PTY:         ptyFile,
        Cmd:         cmd,
        TmuxSession: tmuxSession,
        CreatedAt:   time.Now(),
    }
    
    m.sessions[id] = session
    
    // 启动输出读取协程
    go m.readOutput(session)
    
    return nil
}

// readOutput 读取 PTY 输出并发送到前端
func (m *Manager) readOutput(session *Session) {
    buf := make([]byte, 4096)
    
    for {
        n, err := session.PTY.Read(buf)
        if err != nil {
            // PTY 关闭或出错
            m.Close(session.ID)
            return
        }
        
        if n > 0 {
            // 通过 Wails 事件发送到前端
            runtime.EventsEmit(m.ctx, fmt.Sprintf("terminal-output-%s", session.ID), string(buf[:n]))
        }
    }
}

// SendInput 发送输入到 PTY
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

// Resize 调整 PTY 尺寸
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

// Close 关闭终端会话
func (m *Manager) Close(id string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    session, ok := m.sessions[id]
    if !ok {
        return nil // 已经关闭
    }
    
    // 关闭 PTY
    if session.PTY != nil {
        session.PTY.Close()
    }
    
    // 如果使用 tmux，杀死 session
    if session.TmuxSession != "" && m.hasTmux {
        exec.Command(m.tmuxPath, "kill-session", "-t", session.TmuxSession).Run()
    }
    
    // 等待进程结束
    if session.Cmd != nil && session.Cmd.Process != nil {
        session.Cmd.Process.Kill()
    }
    
    delete(m.sessions, id)
    return nil
}

// CloseAll 关闭所有会话
func (m *Manager) CloseAll() {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    for id := range m.sessions {
        m.Close(id)
    }
}

// ListSessions 列出所有会话
func (m *Manager) ListSessions() []SessionInfo {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    var infos []SessionInfo
    for _, session := range m.sessions {
        infos = append(infos, SessionInfo{
            ID:         session.ID,
            WorktreeID: session.WorktreeID,
            CWD:        session.CWD,
            CreatedAt:  session.CreatedAt,
        })
    }
    return infos
}

type SessionInfo struct {
    ID         string    `json:"id"`
    WorktreeID string    `json:"worktreeId"`
    CWD        string    `json:"cwd"`
    CreatedAt  time.Time `json:"createdAt"`
}
```

### 2.3 App.go 绑定

```go
// app.go (新增)
func (a *App) CreateTerminal(id, worktreeId string) error {
    // 获取 worktree 路径
    worktrees, err := a.worktree.List()
    if err != nil {
        return err
    }
    
    var cwd string
    for _, wt := range worktrees {
        if wt.ID == worktreeId {
            cwd = wt.Path
            break
        }
    }
    
    if cwd == "" {
        return fmt.Errorf("worktree not found: %s", worktreeId)
    }
    
    return a.terminal.CreateSession(id, worktreeId, cwd)
}

func (a *App) SendTerminalInput(id, data string) error {
    return a.terminal.SendInput(id, data)
}

func (a *App) ResizeTerminal(id string, cols, rows uint16) error {
    return a.terminal.Resize(id, cols, rows)
}

func (a *App) CloseTerminal(id string) error {
    return a.terminal.Close(id)
}

func (a *App) ListTerminalSessions() []terminal.SessionInfo {
    return a.terminal.ListSessions()
}

func (a *App) HasTmux() bool {
    return a.terminal.HasTmux()
}
```

### 2.4 前端 - 类型定义

```typescript
// frontend/src/types/terminal.ts
export interface SessionInfo {
  id: string
  worktreeId: string
  cwd: string
  createdAt: string
}

export interface TerminalConfig {
  shell: string
  fontFamily: string
  fontSize: number
  theme: TerminalTheme
}

export interface TerminalTheme {
  background: string
  foreground: string
  cursor: string
  selection: string
}
```

### 2.5 前端 - Store

```typescript
// frontend/src/stores/terminalStore.ts
import { create } from 'zustand'
import { SessionInfo } from '../types/terminal'
import {
  CreateTerminal,
  CloseTerminal,
  ListTerminalSessions,
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

interface TerminalState {
  sessions: SessionInfo[]
  activeSessionId: string | null
  loading: boolean
  error: string | null
  
  // Actions
  fetchSessions: () => Promise<void>
  createSession: (id: string, worktreeId: string) => Promise<void>
  closeSession: (id: string) => Promise<void>
  setActiveSession: (id: string | null) => void
}

export const useTerminalStore = create<TerminalState>((set, get) => ({
  sessions: [],
  activeSessionId: null,
  loading: false,
  error: null,
  
  fetchSessions: async () => {
    try {
      const sessions = await ListTerminalSessions()
      set({ sessions })
    } catch (err) {
      set({ error: String(err) })
    }
  },
  
  createSession: async (id: string, worktreeId: string) => {
    set({ loading: true, error: null })
    try {
      await CreateTerminal(id, worktreeId)
      await get().fetchSessions()
      set({ activeSessionId: id, loading: false })
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },
  
  closeSession: async (id: string) => {
    try {
      await CloseTerminal(id)
      await get().fetchSessions()
      // 如果关闭的是当前活跃的，清除选中
      if (get().activeSessionId === id) {
        const remaining = get().sessions.filter(s => s.id !== id)
        set({ 
          activeSessionId: remaining.length > 0 ? remaining[0].id : null 
        })
      }
    } catch (err) {
      set({ error: String(err) })
    }
  },
  
  setActiveSession: (id) => set({ activeSessionId: id }),
}))
```

### 2.6 前端 - useTerminal Hook

```typescript
// frontend/src/hooks/useTerminal.ts
import { useEffect, useRef, useCallback } from 'react'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import {
  SendTerminalInput,
  ResizeTerminal,
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useTerminalStore } from '../stores/terminalStore'
import { useConfigStore } from '../stores/configStore'

export function useTerminal(sessionId: string) {
  const terminalRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const containerRef = useRef<HTMLDivElement | null>(null)
  
  const { config } = useConfigStore()
  
  // 初始化终端
  useEffect(() => {
    if (!containerRef.current) return
    
    // 创建终端实例
    const term = new Terminal({
      fontFamily: config?.terminal?.fontFamily || 'JetBrains Mono, monospace',
      fontSize: config?.terminal?.fontSize || 14,
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#ffffff',
        selection: 'rgba(255, 255, 255, 0.3)',
      },
      cursorBlink: true,
      cursorStyle: 'block',
      allowTransparency: true,
    })
    
    // 加载插件
    const fitAddon = new FitAddon()
    const webLinksAddon = new WebLinksAddon()
    
    term.loadAddon(fitAddon)
    term.loadAddon(webLinksAddon)
    term.open(containerRef.current)
    
    // 首次适配
    setTimeout(() => {
      fitAddon.fit()
      ResizeTerminal(sessionId, term.cols, term.rows)
    }, 0)
    
    // 输入处理
    term.onData((data) => {
      SendTerminalInput(sessionId, data)
    })
    
    // 监听输出
    EventsOn(`terminal-output-${sessionId}`, (output: string) => {
      term.write(output)
    })
    
    terminalRef.current = term
    fitAddonRef.current = fitAddon
    
    // 清理
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
  
  // 聚焦
  const focus = useCallback(() => {
    terminalRef.current?.focus()
  }, [])
  
  return {
    containerRef,
    terminal: terminalRef,
    handleResize,
    focus,
  }
}
```

### 2.7 前端 - Terminal 组件

```tsx
// frontend/src/components/Terminal/Terminal.tsx
import { useEffect, useRef } from 'react'
import { useTerminal } from '../../hooks/useTerminal'
import 'xterm/css/xterm.css'

interface TerminalProps {
  sessionId: string
  className?: string
}

export default function Terminal({ sessionId, className }: TerminalProps) {
  const { containerRef, handleResize } = useTerminal(sessionId)
  
  // 监听容器尺寸变化
  useEffect(() => {
    if (!containerRef.current) return
    
    const observer = new ResizeObserver(() => {
      handleResize()
    })
    
    observer.observe(containerRef.current)
    
    return () => observer.disconnect()
  }, [handleResize])
  
  return (
    <div 
      ref={containerRef} 
      className={`w-full h-full ${className || ''}`}
    />
  )
}
```

```tsx
// frontend/src/components/Terminal/TerminalPane.tsx
import { useEffect } from 'react'
import Terminal from './Terminal'
import { useTerminalStore } from '../../stores/terminalStore'
import { useWorktreeStore } from '../../stores/worktreeStore'

export default function TerminalPane() {
  const { selectedId: worktreeId } = useWorktreeStore()
  const { 
    sessions, 
    activeSessionId, 
    createSession, 
    setActiveSession,
    loading 
  } = useTerminalStore()
  
  // 当选中 worktree 时，创建或激活对应的终端
  useEffect(() => {
    if (!worktreeId) return
    
    // 查找该 worktree 的现有 session
    const existingSession = sessions.find(s => s.worktreeId === worktreeId)
    
    if (existingSession) {
      setActiveSession(existingSession.id)
    } else {
      // 创建新 session
      const sessionId = `terminal-${Date.now()}`
      createSession(sessionId, worktreeId)
    }
  }, [worktreeId])
  
  if (!activeSessionId) {
    return (
      <div className="flex items-center justify-center h-full text-gray-500">
        Select a worktree to start terminal
      </div>
    )
  }
  
  if (loading) {
    return (
      <div className="flex items-center justify-center h-full text-gray-500">
        Starting terminal...
      </div>
    )
  }
  
  return (
    <div className="h-full flex flex-col">
      {/* Tab 栏 */}
      <div className="flex items-center border-b border-gray-700 bg-gray-800">
        {sessions.map(session => (
          <button
            key={session.id}
            className={`
              px-4 py-2 text-sm border-b-2 transition-colors
              ${activeSessionId === session.id 
                ? 'border-blue-500 text-white' 
                : 'border-transparent text-gray-400 hover:text-white'}
            `}
            onClick={() => setActiveSession(session.id)}
          >
            {session.worktreeId}
          </button>
        ))}
        
        {/* 新建 Tab 按钮 */}
        {worktreeId && (
          <button
            className="px-3 py-2 text-gray-400 hover:text-white"
            onClick={() => {
              const sessionId = `terminal-${Date.now()}`
              createSession(sessionId, worktreeId)
            }}
          >
            +
          </button>
        )}
      </div>
      
      {/* 终端区域 */}
      <div className="flex-1 overflow-hidden">
        {sessions.map(session => (
          <div
            key={session.id}
            className={`h-full ${activeSessionId === session.id ? '' : 'hidden'}`}
          >
            <Terminal sessionId={session.id} />
          </div>
        ))}
      </div>
    </div>
  )
}
```

---

## 3. 测试计划

### 3.1 单元测试

```go
// internal/terminal/pty_test.go
package terminal

import (
    "context"
    "os"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestManager_CreateSession(t *testing.T) {
    ctx := context.Background()
    m := NewManager(ctx)
    
    // 创建临时目录
    tmpDir := t.TempDir()
    
    err := m.CreateSession("test-1", "worktree-1", tmpDir)
    require.NoError(t, err)
    
    // 验证 session 存在
    m.mu.RLock()
    _, exists := m.sessions["test-1"]
    m.mu.RUnlock()
    
    assert.True(t, exists)
    
    // 清理
    m.Close("test-1")
}

func TestManager_SendInput(t *testing.T) {
    ctx := context.Background()
    m := NewManager(ctx)
    
    tmpDir := t.TempDir()
    
    err := m.CreateSession("test-2", "worktree-1", tmpDir)
    require.NoError(t, err)
    defer m.Close("test-2")
    
    // 发送输入
    err = m.SendInput("test-2", "echo hello\n")
    assert.NoError(t, err)
    
    // 给进程时间执行
    time.Sleep(100 * time.Millisecond)
}

func TestManager_Resize(t *testing.T) {
    ctx := context.Background()
    m := NewManager(ctx)
    
    tmpDir := t.TempDir()
    
    err := m.CreateSession("test-3", "worktree-1", tmpDir)
    require.NoError(t, err)
    defer m.Close("test-3")
    
    // 调整大小
    err = m.Resize("test-3", 120, 40)
    assert.NoError(t, err)
}

func TestManager_Close(t *testing.T) {
    ctx := context.Background()
    m := NewManager(ctx)
    
    tmpDir := t.TempDir()
    
    err := m.CreateSession("test-4", "worktree-1", tmpDir)
    require.NoError(t, err)
    
    // 关闭
    err = m.Close("test-4")
    assert.NoError(t, err)
    
    // 验证 session 已删除
    m.mu.RLock()
    _, exists := m.sessions["test-4"]
    m.mu.RUnlock()
    
    assert.False(t, exists)
}

func TestManager_CloseNonExistent(t *testing.T) {
    ctx := context.Background()
    m := NewManager(ctx)
    
    // 关闭不存在的 session 不应该报错
    err := m.Close("non-existent")
    assert.NoError(t, err)
}
```

### 3.2 集成测试

```go
// internal/terminal/tmux_test.go
// +build integration

package terminal

import (
    "context"
    "os/exec"
    "testing"
    
    "github.com/stretchr/testify/require"
)

func TestTmuxIntegration(t *testing.T) {
    // 检查 tmux 是否可用
    _, err := exec.LookPath("tmux")
    if err != nil {
        t.Skip("tmux not available")
    }
    
    ctx := context.Background()
    m := NewManager(ctx)
    
    require.True(t, m.hasTmux)
    
    tmpDir := t.TempDir()
    
    err = m.CreateSession("tmux-test", "wt-1", tmpDir)
    require.NoError(t, err)
    defer m.Close("tmux-test")
    
    // 验证 tmux session 创建
    session := m.sessions["tmux-test"]
    require.NotEmpty(t, session.TmuxSession)
}
```

### 3.3 手动测试清单

- [ ] 打开应用，选择 worktree，终端自动创建
- [ ] 终端显示 shell 提示符
- [ ] 能输入命令并看到输出
- [ ] 能运行 `ls -la`、`pwd` 等基本命令
- [ ] 能运行 `vim` 并正常退出
- [ ] 能运行 `htop` 并正常退出
- [ ] Tab 切换后终端保持状态
- [ ] 调整窗口大小，终端自适应
- [ ] 关闭应用，终端进程正确终止

---

## 4. 验收标准

| 标准 | 描述 |
|------|------|
| 终端创建 | 选择 worktree 后能自动创建终端 |
| 输入输出 | 能输入命令并看到正确输出 |
| 尺寸适配 | 窗口调整时终端正确 resize |
| tmux 支持 | 如果 tmux 可用，使用 tmux session |
| 进程清理 | 关闭应用时所有 PTY 进程终止 |

---

## 5. 发布检查清单

- [ ] 所有单元测试通过
- [ ] tmux 集成测试通过（如有 tmux）
- [ ] 手动测试清单完成
- [ ] golangci-lint 通过
- [ ] TypeScript 检查通过
- [ ] 更新 CHANGELOG.md
- [ ] 创建 Git tag: v0.3.0-alpha
- [ ] GitHub Release 构建

---

## 6. 依赖更新

### Go 依赖

```go
// 新增
require (
    github.com/creack/pty v1.1.21
)
```

### 前端依赖

```json
{
  "dependencies": {
    "xterm": "^5.3.0",
    "xterm-addon-fit": "^0.8.0",
    "xterm-addon-web-links": "^0.9.0"
  }
}
```

---

## 7. 时间估算

| 任务 | 时间 |
|------|------|
| Go PTY Manager 实现 | 2 天 |
| tmux 集成 | 1 天 |
| Go 测试编写 | 1 天 |
| 前端 xterm.js 集成 | 1.5 天 |
| 前端组件和 Hook | 1.5 天 |
| 集成测试和调试 | 2 天 |
| 文档和发布 | 1 天 |
| **总计** | **10 天 (2 周)** |

---

## 8. 已知限制

1. **Windows 不支持 tmux**：Windows 用户将直接使用 shell
2. **无持久化**：应用重启后终端会话丢失（Phase 3 解决）
3. **单终端**：每个 worktree 暂时只支持一个终端（Phase 3 支持多 Tab）

---

## 9. 方案对比（当前 vs 推荐）

### 9.1 架构层面对比

| 维度 | 当前方案（tmux+PTY直连） | 推荐方案（tmux + SessionBridge） |
|------|---------------------------|----------------------------------|
| 会话模型 | `Session` 同时承载进程、PTY、UI关系 | 分层：`TmuxSession` / `BridgeAttach` / `UITab` |
| 关闭 Tab 语义 | `Close` 常等价 kill | **Close=Detach**，Destroy 才 kill（保活优先） |
| 事件协议 | `terminal-output-${id}` 动态事件名 | 统一事件：`terminal:output/state/error/exit` |
| 重连能力 | 弱（前端重建可能丢上下文） | 强（UI可重连既有 tmux 会话） |
| 并发安全 | `CloseAll -> Close` 有锁重入风险 | 统一 teardown 路径，规避死锁 |
| 可观测性 | 错误以字符串为主 | 错误码 + 状态机 + 可追踪事件 |
| 扩展性（Phase 3+） | 一般 | 高（多窗口/多订阅/恢复自然） |

### 9.2 API 语义对比

| 能力 | 当前文档接口 | 推荐接口语义 |
|------|--------------|-------------|
| 创建 | `CreateSession` | `CreateOrAttachSession` |
| 输入 | `SendInput` | `SendInput`（不变） |
| Resize | `Resize` | `Resize`（不变） |
| 关闭 UI | `Close` | `DetachSession` |
| 销毁会话 | `Close` 内隐含 kill | `DestroySession`（显式） |
| 列表 | `ListSessions` | `ListSessions` + `ListDetachedSessions`（可选） |

---

## 10. 简化数据流图

### 10.1 当前方案（直连）

```text
[React xterm]
   │ onData
   ▼
[Wails IPC: SendInput(sessionId,data)]
   ▼
[terminal.Manager]
   ▼
[PTY <-> tmux/shell process]
   │ stdout/stderr
   ▼
[readOutput goroutine]
   ▼
[Wails EventsEmit("terminal-output-${id}", chunk)]
   ▼
[React EventsOn -> term.write()]
```

### 10.2 推荐方案（SessionBridge）

```text
[React xterm tab]
   │ onData
   ▼
[IPC: SendInput(sessionId,data)]
   ▼
[SessionBridge]  (I/O路由, 状态管理, 重连)
   ▼
[tmux session]  (长期存活)
   │ output stream
   ▼
[SessionBridge normalize event]
   ▼
[EventsEmit("terminal:output", {sessionId,chunk,ts})]
   ▼
[React route by sessionId -> term.write()]
```

---

## 11. 生命周期流（保活优先）

```text
CreateOrAttach
  └─> running
       ├─ UI tab close -> detached (tmux仍在)
       ├─ UI reopen    -> running (reattach)
       ├─ process exit -> exited
       └─ user destroy -> destroyed (kill tmux)
```

---

## 12. 设计建议（落地优先）

1. **默认保活语义**：关闭 Tab 仅 detach，不 kill tmux  
2. **统一事件协议**：避免动态事件名膨胀，前端按 `sessionId` 路由  
3. **统一 teardown**：避免 `CloseAll` 持锁调用 `Close` 造成死锁  
4. **状态驱动 UI**：tab 状态由 `terminal:state` 事件驱动，不靠前端猜测  
5. **先聚焦终端 MVP**：Diff 弹出页在 Phase 2 仅保留接口占位，不进入实现
