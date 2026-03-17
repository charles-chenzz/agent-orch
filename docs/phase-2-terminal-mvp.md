# Phase 2: Terminal MVP

> **周期**：2 周
> **目标**：实现基础终端功能（tmux 方案）
> **依赖**：Phase 0 (Foundation), Phase 1 (Worktree)
> **交付**：v0.3.0-alpha

---

## 0. 实现进度

> **最后更新**：2026-03-16

| 模块 | 完成度 | 状态 |
|------|--------|------|
| Go 后端 - PTY 管理 | 100% | ✅ 完成 |
| Go 后端 - tmux 集成 | 80% | ✅ 核心完成 |
| 前端 - xterm.js 集成 | 0% | ❌ 未开始 |
| 前端 - 终端组件 | 0% | ❌ 未开始 |
| 会话状态与协议 | 100% | ✅ 完成 |
| **整体进度** | **~45%** | 🚧 进行中 |

### 已完成文件

- `internal/terminal/types.go` - 数据结构定义
- `internal/terminal/manager.go` - PTY 管理器实现
- `app.go` - Wails API 绑定

### 待完成

- [ ] 安装 xterm.js 依赖
- [ ] 创建 `frontend/src/stores/terminalStore.ts`
- [ ] 创建 `frontend/src/hooks/useTerminal.ts`
- [ ] 创建 `frontend/src/components/Terminal/Terminal.tsx`
- [ ] 更新 `TerminalPane.tsx` 替换 Mock UI

---

## 1. Feature List

### 1.1 Go 后端 - PTY 管理 ✅

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F2.1 | terminal.Manager 结构（含 sessions 状态管理） | P0 | ✅ |
| F2.2 | CreateOrAttachSession(id, cwd string) - 创建或复用 tmux session | P0 | ✅ |
| F2.3 | SendInput(id, data string) | P0 | ✅ |
| F2.4 | Resize(id, cols, rows) | P0 | ✅ |
| F2.5 | DetachSession(id string) - 仅断开 UI 连接，tmux 保活 | P0 | ✅ |
| F2.5a | DestroySession(id string) - 彻底销毁 tmux session | P0 | ✅ |
| F2.6 | 统一事件协议（terminal:output/state/error/exit） | P0 | ✅ |

### 1.2 Go 后端 - tmux 集成 ✅

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F2.7 | tmux 可用性检测 | P0 | ✅ |
| F2.8 | tmux session 创建/附加（-A 参数） | P0 | ✅ |
| F2.9 | tmux session 列表（含 detached 状态） | P1 | ✅ |
| F2.10 | tmux session 销毁（显式 kill-session） | P0 | ✅ |
| F2.20 | tmux 保活策略（UI 关闭仅 detach，不 kill） | P0 | ✅ |

### 1.3 前端 - xterm.js 集成 ❌

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F2.11 | xterm.js 初始化 | P0 | ❌ |
| F2.12 | FitAddon 集成 | P0 | ❌ |
| F2.13 | 输入事件处理 | P0 | ❌ |
| F2.14 | 输出渲染（统一事件路由） | P0 | ❌ |
| F2.15 | 基础主题配置 | P0 | ❌ |

### 1.4 前端 - 终端组件 ❌

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F2.16 | Terminal.tsx 组件 | P0 | ❌ Mock UI |
| F2.17 | useTerminal hook | P0 | ❌ |
| F2.18 | terminalStore（含 session 状态） | P0 | ❌ |
| F2.19 | 单终端 Tab UI | P0 | ❌ Mock UI |
| F2.21 | 会话重连与状态同步 UI | P1 | ❌ |

### 1.5 会话状态与协议 ✅

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F2.22 | Session.State 字段 + 状态机（creating/running/detached/exited/destroyed） | P0 | ✅ |
| F2.23 | StateChange 事件发送（terminal:state） | P0 | ✅ |
| F2.24 | 统一事件协议（terminal:output/state/error/exit） | P0 | ✅ |

---

## 2. 实现细节

### 2.1 Go 后端 - 数据结构

```go
// internal/terminal/types.go
package terminal

import (
    "context"
    "os"
    "os/exec"
    "sync"
    "time"
)

// SessionState 会话状态
type SessionState string

const (
    StateCreating  SessionState = "creating"
    StateRunning   SessionState = "running"
    StateDetached  SessionState = "detached"
    StateExited    SessionState = "exited"
    StateDestroyed SessionState = "destroyed"
)

type Session struct {
    ID          string
    WorktreeID  string
    CWD         string
    State       SessionState    // 当前状态
    PTY         *os.File
    Cmd         *exec.Cmd
    TmuxSession string          // tmux session 名称
    CreatedAt   time.Time
    LastActive  time.Time       // 最后活动时间
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

// EventPayload 统一事件负载
type EventPayload struct {
    SessionID string `json:"sessionId"`
    Type      string `json:"type"`       // output/state/error/exit
    Data      string `json:"data,omitempty"`
    State     string `json:"state,omitempty"`
    Error     string `json:"error,omitempty"`
    Timestamp int64  `json:"ts"`
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
    "time"

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

// CreateOrAttachSession 创建或附加到现有会话
func (m *Manager) CreateOrAttachSession(id, worktreeID, cwd string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 检查是否已存在 detached 的会话
    if session, exists := m.sessions[id]; exists {
        if session.State == StateDetached {
            // 重新附加
            return m.attachSession(session)
        }
        return fmt.Errorf("session already running: %s", id)
    }

    // 创建新会话
    session := &Session{
        ID:          id,
        WorktreeID:  worktreeID,
        CWD:         cwd,
        State:       StateCreating,
        CreatedAt:   time.Now(),
        LastActive:  time.Now(),
    }
    m.sessions[id] = session

    var ptyFile *os.File
    var cmd *exec.Cmd
    var tmuxSession string

    if m.hasTmux {
        // 使用 tmux（-A 如果不存在则创建）
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
        session.State = StateDestroyed
        m.emitState(session)
        return fmt.Errorf("failed to start pty: %w", err)
    }

    session.PTY = ptyFile
    session.Cmd = cmd
    session.TmuxSession = tmuxSession
    session.State = StateRunning
    m.emitState(session)

    // 启动输出读取协程
    go m.readOutput(session)

    return nil
}

// attachSession 重新附加到 detached 的会话
func (m *Manager) attachSession(session *Session) error {
    if session.TmuxSession == "" || !m.hasTmux {
        return fmt.Errorf("cannot reattach to non-tmux session")
    }

    // tmux session 仍在运行，只需重新附加
    session.State = StateRunning
    session.LastActive = time.Now()
    m.emitState(session)

    // 重新启动输出读取
    go m.readOutput(session)

    return nil
}

// readOutput 读取 PTY 输出并发送到前端（统一事件协议）
func (m *Manager) readOutput(session *Session) {
    buf := make([]byte, 4096)

    for {
        n, err := session.PTY.Read(buf)
        if err != nil {
            // PTY 关闭或出错
            m.handleSessionExit(session.ID)
            return
        }

        if n > 0 {
            // 统一事件协议
            m.emitOutput(session.ID, string(buf[:n]))
        }
    }
}

// emitOutput 发送输出事件（统一协议）
func (m *Manager) emitOutput(sessionID, data string) {
    runtime.EventsEmit(m.ctx, "terminal:output", EventPayload{
        SessionID: sessionID,
        Type:      "output",
        Data:      data,
        Timestamp: time.Now().UnixMilli(),
    })
}

// emitState 发送状态变更事件
func (m *Manager) emitState(session *Session) {
    runtime.EventsEmit(m.ctx, "terminal:state", EventPayload{
        SessionID: session.ID,
        Type:      "state",
        State:     string(session.State),
        Timestamp: time.Now().UnixMilli(),
    })
}

// emitError 发送错误事件
func (m *Manager) emitError(sessionID, errMsg string) {
    runtime.EventsEmit(m.ctx, "terminal:error", EventPayload{
        SessionID: sessionID,
        Type:      "error",
        Error:     errMsg,
        Timestamp: time.Now().UnixMilli(),
    })
}

// SendInput 发送输入到 PTY
func (m *Manager) SendInput(id, data string) error {
    m.mu.RLock()
    session, ok := m.sessions[id]
    m.mu.RUnlock()

    if !ok {
        return fmt.Errorf("session not found: %s", id)
    }

    if session.State != StateRunning {
        return fmt.Errorf("session not running: %s (state: %s)", id, session.State)
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

// DetachSession 断开会话（tmux 保活）
func (m *Manager) DetachSession(id string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    session, ok := m.sessions[id]
    if !ok {
        return nil // 已经不存在
    }

    // 关闭 PTY 文件句柄（但不杀死进程）
    if session.PTY != nil {
        session.PTY.Close()
        session.PTY = nil
    }

    // 更新状态
    session.State = StateDetached
    m.emitState(session)

    // 从内存中移除（tmux session 仍在后台运行）
    delete(m.sessions, id)

    return nil
}

// DestroySession 彻底销毁会话
func (m *Manager) DestroySession(id string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    session, ok := m.sessions[id]
    if !ok {
        return nil // 已经不存在
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

    // 发送销毁事件
    session.State = StateDestroyed
    m.emitState(session)

    delete(m.sessions, id)
    return nil
}

// handleSessionExit 处理会话退出（进程自然结束）
func (m *Manager) handleSessionExit(id string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    session, ok := m.sessions[id]
    if !ok {
        return
    }

    session.State = StateExited
    m.emitState(session)

    // 清理资源
    if session.PTY != nil {
        session.PTY.Close()
    }

    delete(m.sessions, id)
}

// CloseAll 销毁所有会话
func (m *Manager) CloseAll() {
    m.mu.Lock()
    // 先收集所有 ID，避免遍历时修改
    ids := make([]string, 0, len(m.sessions))
    for id := range m.sessions {
        ids = append(ids, id)
    }
    m.mu.Unlock()

    // 逐个销毁（不持锁调用 DestroySession）
    for _, id := range ids {
        m.DestroySession(id)
    }
}

// ListSessions 列出所有活跃会话
func (m *Manager) ListSessions() []SessionInfo {
    m.mu.RLock()
    defer m.mu.RUnlock()

    var infos []SessionInfo
    for _, session := range m.sessions {
        infos = append(infos, SessionInfo{
            ID:         session.ID,
            WorktreeID: session.WorktreeID,
            CWD:        session.CWD,
            State:      string(session.State),
            CreatedAt:  session.CreatedAt,
            LastActive: session.LastActive,
        })
    }
    return infos
}

// GetSessionState 获取会话状态
func (m *Manager) GetSessionState(id string) (SessionState, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    session, ok := m.sessions[id]
    if !ok {
        return "", fmt.Errorf("session not found: %s", id)
    }
    return session.State, nil
}

// HasTmux 返回 tmux 是否可用
func (m *Manager) HasTmux() bool {
    return m.hasTmux
}

type SessionInfo struct {
    ID         string    `json:"id"`
    WorktreeID string    `json:"worktreeId"`
    CWD        string    `json:"cwd"`
    State      string    `json:"state"`
    CreatedAt  time.Time `json:"createdAt"`
    LastActive time.Time `json:"lastActive"`
}
```

### 2.3 App.go 绑定

```go
// app.go (新增)
func (a *App) CreateOrAttachTerminal(id, worktreeId string) error {
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

    return a.terminal.CreateOrAttachSession(id, worktreeId, cwd)
}

func (a *App) SendTerminalInput(id, data string) error {
    return a.terminal.SendInput(id, data)
}

func (a *App) ResizeTerminal(id string, cols, rows uint16) error {
    return a.terminal.Resize(id, cols, rows)
}

// DetachTerminal 断开终端（保活，可重连）
func (a *App) DetachTerminal(id string) error {
    return a.terminal.DetachSession(id)
}

// DestroyTerminal 彻底销毁终端
func (a *App) DestroyTerminal(id string) error {
    return a.terminal.DestroySession(id)
}

func (a *App) ListTerminalSessions() []terminal.SessionInfo {
    return a.terminal.ListSessions()
}

func (a *App) GetTerminalState(id string) (string, error) {
    state, err := a.terminal.GetSessionState(id)
    if err != nil {
        return "", err
    }
    return string(state), nil
}

func (a *App) HasTmux() bool {
    return a.terminal.HasTmux()
}
```

### 2.4 前端 - 类型定义

```typescript
// frontend/src/types/terminal.ts

export type SessionState = 'creating' | 'running' | 'detached' | 'exited' | 'destroyed'

export interface SessionInfo {
  id: string
  worktreeId: string
  cwd: string
  state: SessionState
  createdAt: string
  lastActive: string
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

// 统一事件协议
export interface TerminalEvent {
  sessionId: string
  type: 'output' | 'state' | 'error' | 'exit'
  data?: string
  state?: SessionState
  error?: string
  ts: number  // timestamp in ms
}

// 事件类型常量
export const TERMINAL_EVENTS = {
  OUTPUT: 'terminal:output',
  STATE: 'terminal:state',
  ERROR: 'terminal:error',
  EXIT: 'terminal:exit',
} as const
```

### 2.5 前端 - Store

```typescript
// frontend/src/stores/terminalStore.ts
import { create } from 'zustand'
import { SessionInfo, SessionState, TerminalEvent, TERMINAL_EVENTS } from '../types/terminal'
import {
  CreateOrAttachTerminal,
  DetachTerminal,
  DestroyTerminal,
  ListTerminalSessions,
  GetTerminalState,
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

interface TerminalState {
  sessions: SessionInfo[]
  activeSessionId: string | null
  loading: boolean
  error: string | null

  // Actions
  fetchSessions: () => Promise<void>
  createOrAttachSession: (id: string, worktreeId: string) => Promise<void>
  detachSession: (id: string) => Promise<void>
  destroySession: (id: string) => Promise<void>
  setActiveSession: (id: string | null) => void
  updateSessionState: (id: string, state: SessionState) => void

  // Event handling
  subscribeToEvents: () => void
  unsubscribeFromEvents: () => void
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

  createOrAttachSession: async (id: string, worktreeId: string) => {
    set({ loading: true, error: null })
    try {
      await CreateOrAttachTerminal(id, worktreeId)
      await get().fetchSessions()
      set({ activeSessionId: id, loading: false })
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },

  detachSession: async (id: string) => {
    try {
      await DetachTerminal(id)
      await get().fetchSessions()
      // 如果断开的是当前活跃的，清除选中
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

  destroySession: async (id: string) => {
    try {
      await DestroyTerminal(id)
      await get().fetchSessions()
      // 如果销毁的是当前活跃的，清除选中
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

  updateSessionState: (id: string, state: SessionState) => {
    set((s) => ({
      sessions: s.sessions.map((sess) =>
        sess.id === id ? { ...sess, state } : sess
      ),
    }))
  },

  subscribeToEvents: () => {
    // 订阅统一事件协议
    EventsOn(TERMINAL_EVENTS.STATE, (event: TerminalEvent) => {
      get().updateSessionState(event.sessionId, event.state!)
    })

    EventsOn(TERMINAL_EVENTS.ERROR, (event: TerminalEvent) => {
      set({ error: event.error })
    })

    EventsOn(TERMINAL_EVENTS.EXIT, (event: TerminalEvent) => {
      get().updateSessionState(event.sessionId, 'exited')
    })
  },

  unsubscribeFromEvents: () => {
    EventsOff(TERMINAL_EVENTS.STATE)
    EventsOff(TERMINAL_EVENTS.ERROR)
    EventsOff(TERMINAL_EVENTS.EXIT)
  },
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
import { TerminalEvent, TERMINAL_EVENTS } from '../types/terminal'

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

    // 监听统一输出事件（按 sessionId 路由）
    EventsOn(TERMINAL_EVENTS.OUTPUT, (event: TerminalEvent) => {
      if (event.sessionId === sessionId && event.data) {
        term.write(event.data)
      }
    })

    terminalRef.current = term
    fitAddonRef.current = fitAddon

    // 清理（仅取消订阅，不销毁会话）
    return () => {
      EventsOff(TERMINAL_EVENTS.OUTPUT)
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
    createOrAttachSession,
    detachSession,
    destroySession,
    setActiveSession,
    subscribeToEvents,
    unsubscribeFromEvents,
    loading
  } = useTerminalStore()

  // 订阅终端事件
  useEffect(() => {
    subscribeToEvents()
    return () => unsubscribeFromEvents()
  }, [])

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
      createOrAttachSession(sessionId, worktreeId)
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

  const activeSession = sessions.find(s => s.id === activeSessionId)

  return (
    <div className="h-full flex flex-col">
      {/* Tab 栏 */}
      <div className="flex items-center border-b border-gray-700 bg-gray-800">
        {sessions.map(session => (
          <button
            key={session.id}
            className={`
              px-4 py-2 text-sm border-b-2 transition-colors flex items-center gap-2
              ${activeSessionId === session.id
                ? 'border-blue-500 text-white'
                : 'border-transparent text-gray-400 hover:text-white'}
            `}
            onClick={() => setActiveSession(session.id)}
          >
            <span>{session.worktreeId}</span>
            {/* 状态指示器 */}
            <span className={`
              w-2 h-2 rounded-full
              ${session.state === 'running' ? 'bg-green-500' : ''}
              ${session.state === 'detached' ? 'bg-yellow-500' : ''}
              ${session.state === 'exited' ? 'bg-gray-500' : ''}
            `} />
            {/* 关闭按钮（detach） */}
            <span
              className="ml-1 text-gray-500 hover:text-white"
              onClick={(e) => {
                e.stopPropagation()
                detachSession(session.id)
              }}
            >
              ×
            </span>
          </button>
        ))}

        {/* 新建 Tab 按钮 */}
        {worktreeId && (
          <button
            className="px-3 py-2 text-gray-400 hover:text-white"
            onClick={() => {
              const sessionId = `terminal-${Date.now()}`
              createOrAttachSession(sessionId, worktreeId)
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

      {/* 状态栏 */}
      {activeSession && (
        <div className="px-4 py-1 text-xs text-gray-500 border-t border-gray-700 bg-gray-800 flex justify-between">
          <span>State: {activeSession.state}</span>
          <span>CWD: {activeSession.cwd}</span>
        </div>
      )}
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

## 10. 数据流图（统一事件协议）

```text
[React xterm tab]
   │ onData
   ▼
[Wails IPC: SendInput(sessionId, data)]
   ▼
[terminal.Manager]
   ▼
[PTY <-> tmux/shell process]
   │ stdout/stderr
   ▼
[readOutput goroutine]
   ▼
[EventsEmit("terminal:output", {sessionId, chunk, ts})]
   ▼
[React EventsOn -> 按 sessionId 路由 -> term.write()]
```

---

## 11. 生命周期流（保活优先）

```text
CreateOrAttachSession
  └─> creating -> running
       ├─ UI tab close -> DetachSession -> detached (tmux仍在)
       ├─ UI reopen    -> CreateOrAttach -> running (reattach)
       ├─ process exit -> exited (自然结束)
       └─ user destroy -> DestroySession -> destroyed (kill tmux)
```

---

## 12. 设计要点总结

| 要点 | 说明 |
|------|------|
| **保活语义** | 关闭 Tab 仅 `DetachSession`，tmux 会话保活；显式 `DestroySession` 才 kill |
| **统一事件协议** | `terminal:output/state/error/exit`，前端按 `sessionId` 路由，避免动态事件名膨胀 |
| **状态机** | `Session.State` 字段驱动 UI，通过 `terminal:state` 事件同步 |
| **死锁规避** | `CloseAll` 先收集 ID 再逐个销毁，不持锁调用 `DestroySession` |
| **重连能力** | tmux 会话可被 `CreateOrAttachSession` 重新附加 |

---

## 13. 已知限制

1. **Windows 不支持 tmux**：Windows 用户将直接使用 shell（无保活）
2. **无持久化**：应用重启后终端会话丢失（Phase 3 解决）
3. **单终端**：每个 worktree 暂时只支持一个终端（Phase 3 支持多 Tab）
