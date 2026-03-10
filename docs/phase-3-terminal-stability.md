# Phase 3: Terminal Stability

> **周期**：2 周
> **目标**：TUI 应用兼容 + 多终端 Tab + 会话持久化
> **依赖**：Phase 2 (Terminal MVP)
> **交付**：v0.4.0-alpha

---

## 1. Feature List

### 1.1 终端稳定性

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F3.1 | Tab 切换状态恢复 | P0 |
| F3.2 | ResizeObserver + 防抖 | P0 |
| F3.3 | Resize 闪烁修复 | P0 |
| F3.4 | 光标位置恢复 | P1 |
| F3.5 | 滚动位置保持 | P1 |

### 1.2 多终端 Tab

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F3.6 | 多 Tab 创建 | P0 |
| F3.7 | Tab 关闭 | P0 |
| F3.8 | Tab 切换 | P0 |
| F3.9 | Tab 重命名 | P2 |
| F3.10 | Tab 拖拽排序 | P2 |

### 1.3 会话持久化

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F3.11 | 会话状态保存到 DB | P1 |
| F3.12 | 应用重启后恢复 | P1 |
| F3.13 | tmux session 附加 | P0 |

### 1.4 TUI 兼容性

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F3.14 | vim 兼容测试 | P0 |
| F3.15 | htop 兼容测试 | P0 |
| F3.16 | Claude Code 兼容测试 | P0 |
| F3.17 | Codex 兼容测试 | P1 |

---

## 2. 问题分析与解决方案

### 2.1 Tab 切换问题根因

```
问题 1: Alternate Screen Buffer
┌─────────────────────────────────────┐
│ 正常模式                            │
│ $ command 1                         │
│ $ command 2                         │
│ $ _                                 │
└─────────────────────────────────────┘
        │ vim 启动
        ▼
┌─────────────────────────────────────┐
│ Alternate Screen Buffer             │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ 1 import React from 'react'     │ │
│ │ 2                               │ │
│ │ 3 function App() {              │ │
│ │ 4   return <div>Hello</div>     │ │
│ │ 5 }                             │ │
│ │ ~                               │ │
│ │ ~                               │ │
│ └─────────────────────────────────┘ │
│ :_                                  │
└─────────────────────────────────────┘
```

当 Tab 切换时，如果处理不当：
- xterm.js DOM 被 hidden，失去焦点
- Alternate screen 状态丢失
- PTY 不知道窗口尺寸变化

### 2.2 解决方案

```typescript
// 核心策略：不销毁终端，只隐藏 DOM
// 使用 CSS display: none 而不是卸载组件

// 错误做法
{activeSessionId === session.id && (
  <Terminal sessionId={session.id} />
)}

// 正确做法
<div className={activeSessionId === session.id ? '' : 'hidden'}>
  <Terminal sessionId={session.id} />
</div>
```

---

## 3. 实现细节

### 3.1 改进的 TerminalPane

```tsx
// frontend/src/components/Terminal/TerminalPane.tsx
import { useEffect, useRef, useCallback } from 'react'
import Terminal from './Terminal'
import TerminalTab from './TerminalTab'
import { useTerminalStore } from '../../stores/terminalStore'
import { useWorktreeStore } from '../../stores/worktreeStore'

export default function TerminalPane() {
  const { selectedId: worktreeId } = useWorktreeStore()
  const {
    sessions,
    activeSessionId,
    createSession,
    setActiveSession,
    closeSession,
  } = useTerminalStore()
  
  const containerRefs = useRef<Map<string, HTMLDivElement>>(new Map())
  
  // 初始化或恢复会话
  useEffect(() => {
    if (!worktreeId) return
    
    const existingSessions = sessions.filter(s => s.worktreeId === worktreeId)
    
    if (existingSessions.length === 0) {
      // 创建第一个终端
      createSession(`terminal-${Date.now()}`, worktreeId)
    } else if (!activeSessionId) {
      // 恢复选中的会话
      setActiveSession(existingSessions[0].id)
    }
  }, [worktreeId])
  
  // 处理 Tab 切换焦点
  useEffect(() => {
    if (!activeSessionId) return
    
    // 延迟聚焦，确保 DOM 已渲染
    requestAnimationFrame(() => {
      const container = containerRefs.current.get(activeSessionId)
      if (container) {
        const terminalEl = container.querySelector('.xterm')
        if (terminalEl) {
          ;(terminalEl as HTMLElement).focus()
        }
      }
    })
  }, [activeSessionId])
  
  // 新建 Tab
  const handleNewTab = useCallback(() => {
    if (!worktreeId) return
    createSession(`terminal-${Date.now()}`, worktreeId)
  }, [worktreeId, createSession])
  
  // 关闭 Tab
  const handleCloseTab = useCallback((sessionId: string) => {
    closeSession(sessionId)
  }, [closeSession])
  
  return (
    <div className="h-full flex flex-col bg-gray-900">
      {/* Tab 栏 */}
      <div className="flex items-center border-b border-gray-700 bg-gray-800 overflow-x-auto">
        {sessions.map(session => (
          <TerminalTab
            key={session.id}
            session={session}
            isActive={activeSessionId === session.id}
            onClick={() => setActiveSession(session.id)}
            onClose={() => handleCloseTab(session.id)}
          />
        ))}
        
        {/* 新建按钮 */}
        <button
          className="px-3 py-2 text-gray-400 hover:text-white hover:bg-gray-700 shrink-0"
          onClick={handleNewTab}
          title="New terminal"
        >
          <PlusIcon className="w-4 h-4" />
        </button>
      </div>
      
      {/* 终端容器 - 关键：使用 hidden 而不是条件渲染 */}
      <div className="flex-1 relative overflow-hidden">
        {sessions.map(session => (
          <div
            key={session.id}
            ref={el => {
              if (el) containerRefs.current.set(session.id, el)
            }}
            className={`absolute inset-0 ${activeSessionId === session.id ? '' : 'hidden'}`}
          >
            <Terminal sessionId={session.id} />
          </div>
        ))}
      </div>
    </div>
  )
}
```

### 3.2 改进的 Terminal 组件

```tsx
// frontend/src/components/Terminal/Terminal.tsx
import { useEffect, useRef, useCallback, useState } from 'react'
import { Terminal as XTerm } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { SerializeAddon } from 'xterm-addon-serialize'
import { WebLinksAddon } from 'xterm-addon-web-links'
import {
  SendTerminalInput,
  ResizeTerminal,
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useTerminalStore } from '../../stores/terminalStore'
import { useConfigStore } from '../stores/configStore'
import 'xterm/css/xterm.css'

interface TerminalProps {
  sessionId: string
}

export default function Terminal({ sessionId }: TerminalProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const terminalRef = useRef<XTerm | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const serializeAddonRef = useRef<SerializeAddon | null>(null)
  const resizeObserverRef = useRef<ResizeObserver | null>(null)
  const [isReady, setIsReady] = useState(false)
  
  const { config } = useConfigStore()
  
  // 初始化终端
  useEffect(() => {
    if (!containerRef.current || terminalRef.current) return
    
    const term = new XTerm({
      fontFamily: config?.terminal?.fontFamily || 'JetBrains Mono, Consolas, monospace',
      fontSize: config?.terminal?.fontSize || 14,
      lineHeight: 1.2,
      theme: {
        background: '#0d1117',
        foreground: '#c9d1d9',
        cursor: '#c9d1d9',
        cursorAccent: '#0d1117',
        selection: 'rgba(56, 139, 253, 0.4)',
        black: '#484f58',
        red: '#f97583',
        green: '#85e89d',
        yellow: '#ffea7f',
        blue: '#79b8ff',
        magenta: '#b392f0',
        cyan: '#56d4dd',
        white: '#f0f6fc',
        brightBlack: '#6a737d',
        brightRed: '#fdaeb7',
        brightGreen: '#bef5cb',
        brightYellow: '#fff5b1',
        brightBlue: '#c8e1ff',
        brightMagenta: '#d2b3ff',
        brightCyan: '#a4f4fd',
        brightWhite: '#fafbfc',
      },
      cursorBlink: true,
      cursorStyle: 'block',
      allowTransparency: true,
      scrollback: 10000,
    })
    
    // 加载插件
    const fitAddon = new FitAddon()
    const serializeAddon = new SerializeAddon()
    const webLinksAddon = new WebLinksAddon()
    
    term.loadAddon(fitAddon)
    term.loadAddon(serializeAddon)
    term.loadAddon(webLinksAddon)
    
    term.open(containerRef.current)
    
    terminalRef.current = term
    fitAddonRef.current = fitAddon
    serializeAddonRef.current = serializeAddon
    
    // 首次适配
    requestAnimationFrame(() => {
      fitAddon.fit()
      ResizeTerminal(sessionId, term.cols, term.rows)
      setIsReady(true)
    })
    
    // 输入处理
    term.onData((data) => {
      SendTerminalInput(sessionId, data)
    })
    
    // 监听输出
    EventsOn(`terminal-output-${sessionId}`, (output: string) => {
      term.write(output)
    })
    
    // 清理
    return () => {
      EventsOff(`terminal-output-${sessionId}`)
      // 注意：不 dispose 终端，保持状态
    }
  }, [sessionId])
  
  // Resize 处理 - 带防抖
  const handleResize = useCallback(
    debounce(() => {
      if (terminalRef.current && fitAddonRef.current) {
        try {
          fitAddonRef.current.fit()
          ResizeTerminal(
            sessionId,
            terminalRef.current.cols,
            terminalRef.current.rows
          )
        } catch (e) {
          // 忽略 resize 错误
        }
      }
    }, 50),
    [sessionId]
  )
  
  // 监听容器尺寸变化
  useEffect(() => {
    if (!containerRef.current) return
    
    const observer = new ResizeObserver(() => {
      handleResize()
    })
    
    resizeObserverRef.current = observer
    observer.observe(containerRef.current)
    
    return () => {
      observer.disconnect()
    }
  }, [handleResize])
  
  // 聚焦
  const focus = useCallback(() => {
    if (terminalRef.current) {
      terminalRef.current.focus()
    }
  }, [])
  
  // 序列化状态（用于持久化）
  const serialize = useCallback(() => {
    if (serializeAddonRef.current) {
      return serializeAddonRef.current.serialize()
    }
    return ''
  }, [])
  
  return (
    <div 
      ref={containerRef} 
      className="w-full h-full"
      onClick={focus}
    />
  )
}

// 防抖函数
function debounce<T extends (...args: any[]) => any>(
  fn: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timeoutId: ReturnType<typeof setTimeout> | null = null
  
  return (...args: Parameters<T>) => {
    if (timeoutId) clearTimeout(timeoutId)
    timeoutId = setTimeout(() => fn(...args), delay)
  }
}
```

### 3.3 TerminalTab 组件

```tsx
// frontend/src/components/Terminal/TerminalTab.tsx
import { useState } from 'react'
import { SessionInfo } from '../../types/terminal'

interface TerminalTabProps {
  session: SessionInfo
  isActive: boolean
  onClick: () => void
  onClose: () => void
}

export default function TerminalTab({ 
  session, 
  isActive, 
  onClick, 
  onClose 
}: TerminalTabProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [name, setName] = useState(session.worktreeId)
  
  // 只有一个 session 时不显示关闭按钮
  const canClose = true // 实际应该检查 session 数量
  
  return (
    <div
      className={`
        group flex items-center gap-2 px-3 py-2 border-b-2 cursor-pointer
        min-w-[100px] max-w-[200px]
        ${isActive 
          ? 'border-blue-500 bg-gray-700' 
          : 'border-transparent text-gray-400 hover:text-white hover:bg-gray-700/50'}
      `}
      onClick={onClick}
      onDoubleClick={() => setIsEditing(true)}
    >
      {/* 图标 */}
      <TerminalIcon className="w-4 h-4 shrink-0" />
      
      {/* 名称 */}
      {isEditing ? (
        <input
          type="text"
          value={name}
          onChange={e => setName(e.target.value)}
          onBlur={() => setIsEditing(false)}
          onKeyDown={e => {
            if (e.key === 'Enter') setIsEditing(false)
          }}
          className="flex-1 bg-transparent border-none outline-none text-sm"
          autoFocus
        />
      ) : (
        <span className="flex-1 text-sm truncate">{name}</span>
      )}
      
      {/* 关闭按钮 */}
      {canClose && (
        <button
          className={`
            p-0.5 rounded hover:bg-gray-600
            ${isActive ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'}
          `}
          onClick={(e) => {
            e.stopPropagation()
            onClose()
          }}
        >
          <XIcon className="w-3 h-3" />
        </button>
      )}
    </div>
  )
}
```

### 3.4 会话持久化

```go
// internal/terminal/session.go
package terminal

import (
    "gorm.io/gorm"
)

// 会话记录 - 存储到数据库
type SessionRecord struct {
    gorm.Model
    ID          string `gorm:"uniqueIndex;size:64"`
    WorktreeID  string `gorm:"index;size:64"`
    CWD         string `gorm:"size:512"`
    TmuxSession string `gorm:"size:128"`
    Cols        uint16
    Rows        uint16
    Active      bool
}

// SaveSession 保存会话状态
func (m *Manager) SaveSession(id string) error {
    m.mu.RLock()
    session, ok := m.sessions[id]
    m.mu.RUnlock()
    
    if !ok {
        return fmt.Errorf("session not found: %s", id)
    }
    
    // 获取当前 PTY 尺寸
    ws, _ := pty.GetsizeFull(session.PTY)
    
    record := SessionRecord{
        ID:          id,
        WorktreeID:  session.WorktreeID,
        CWD:         session.CWD,
        TmuxSession: session.TmuxSession,
        Cols:        ws.Cols,
        Rows:        ws.Rows,
        Active:      true,
    }
    
    return m.db.Save(&record).Error
}

// RestoreSessions 恢复所有会话
func (m *Manager) RestoreSessions() error {
    var records []SessionRecord
    if err := m.db.Where("active = ?", true).Find(&records).Error; err != nil {
        return err
    }
    
    for _, record := range records {
        // 如果使用 tmux，尝试附加到现有 session
        if record.TmuxSession != "" && m.hasTmux {
            // 检查 tmux session 是否还存在
            if m.tmuxSessionExists(record.TmuxSession) {
                // 附加到现有 session
                m.attachTmuxSession(record.ID, record.WorktreeID, record.TmuxSession)
                continue
            }
        }
        
        // 创建新 session
        m.CreateSession(record.ID, record.WorktreeID, record.CWD)
    }
    
    return nil
}

// tmuxSessionExists 检查 tmux session 是否存在
func (m *Manager) tmuxSessionExists(name string) bool {
    cmd := exec.Command(m.tmuxPath, "has-session", "-t", name)
    return cmd.Run() == nil
}

// attachTmuxSession 附加到现有 tmux session
func (m *Manager) attachTmuxSession(id, worktreeID, tmuxSession string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    cmd := exec.Command(m.tmuxPath, "attach-session", "-t", tmuxSession)
    
    ptyFile, err := pty.Start(cmd)
    if err != nil {
        return err
    }
    
    session := &Session{
        ID:          id,
        WorktreeID:  worktreeID,
        PTY:         ptyFile,
        Cmd:         cmd,
        TmuxSession: tmuxSession,
        CreatedAt:   time.Now(),
    }
    
    m.sessions[id] = session
    go m.readOutput(session)
    
    return nil
}
```

### 3.5 应用启动时恢复

```go
// app.go
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    
    // ... 其他初始化 ...
    
    // 初始化终端管理器
    a.terminal = terminal.NewManager(ctx, a.db)
    
    // 恢复之前的会话
    if a.terminal.HasTmux() {
        a.terminal.RestoreSessions()
    }
}

func (a *App) shutdown(ctx context.Context) {
    // 保存所有活跃会话
    for _, info := range a.terminal.ListSessions() {
        a.terminal.SaveSession(info.ID)
    }
    
    // 不关闭 PTY，让 tmux session 保持运行
}
```

---

## 4. TUI 兼容性测试

### 4.1 测试矩阵

| 应用 | 测试项 | 预期结果 |
|------|--------|----------|
| **vim** | 打开文件 | 正常显示 |
| | 编辑内容 | 输入正常 |
| | Tab 切换后返回 | 状态恢复，光标位置正确 |
| | 退出 | 返回 shell |
| **htop** | 启动 | 正常显示进程列表 |
| | 滚动 | 正常滚动 |
| | Tab 切换后返回 | 显示恢复 |
| | 退出 (q) | 返回 shell |
| **Claude Code** | 启动 | 正常显示界面 |
| | 输入提示词 | 输入正常 |
| | Tab 切换后返回 | 界面恢复，无截断 |
| | 长时间运行 | 稳定，无内存泄漏 |
| **less** | 打开大文件 | 正常显示 |
| | 滚动 | 正常滚动 |
| | 搜索 | 功能正常 |
| **tmux** (嵌套) | 启动 | 正常显示 |
| | 创建窗口 | 功能正常 |

### 4.2 自动化测试脚本

```bash
#!/bin/bash
# scripts/test-tui.sh

echo "Testing TUI compatibility..."

# 测试 vim
echo "1. Testing vim..."
timeout 5 vim -c "echo 'Hello from vim'" -c "q" 2>&1 && echo "vim: OK" || echo "vim: FAILED"

# 测试 htop
echo "2. Testing htop..."
timeout 2 htop 2>&1 && echo "htop: OK" || echo "htop: FAILED"

# 测试 less
echo "3. Testing less..."
echo "Test content" | timeout 2 less 2>&1 && echo "less: OK" || echo "less: FAILED"

echo "TUI tests completed."
```

---

## 5. 性能优化

### 5.1 内存优化

```typescript
// 限制滚动缓冲区大小
const term = new XTerm({
  scrollback: 10000,  // 限制为 10000 行
})

// 定期清理不活跃的终端缓冲区
useEffect(() => {
  const cleanup = setInterval(() => {
    // 检查不活跃的终端
    const inactiveSessions = sessions.filter(s => s.id !== activeSessionId)
    
    // 如果内存占用过高，清理最老的不活跃终端
    if (performance.memory && performance.memory.usedJSHeapSize > 500 * 1024 * 1024) {
      // 清理逻辑
    }
  }, 60000) // 每分钟检查一次
  
  return () => clearInterval(cleanup)
}, [sessions, activeSessionId])
```

### 5.2 渲染优化

```typescript
// 使用 requestAnimationFrame 优化 resize
const handleResize = useCallback(() => {
  if (rafRef.current) {
    cancelAnimationFrame(rafRef.current)
  }
  
  rafRef.current = requestAnimationFrame(() => {
    if (terminalRef.current && fitAddonRef.current) {
      fitAddonRef.current.fit()
      ResizeTerminal(sessionId, terminalRef.current.cols, terminalRef.current.rows)
    }
  })
}, [sessionId])
```

---

## 6. 测试计划

### 6.1 单元测试

- [ ] SessionRecord CRUD
- [ ] SaveSession/RestoreSessions
- [ ] tmuxSessionExists
- [ ] attachTmuxSession

### 6.2 手动测试清单

- [ ] 创建 5 个终端 Tab
- [ ] 在每个 Tab 中运行不同的 TUI 应用
- [ ] 快速切换 Tab，验证状态恢复
- [ ] 调整窗口大小，验证 resize 正确
- [ ] 关闭应用，重新打开，验证 tmux session 恢复
- [ ] 长时间运行（1小时+），验证无内存泄漏

---

## 7. 验收标准

| 标准 | 描述 |
|------|------|
| TUI 兼容 | vim/htop/Claude Code 运行正常 |
| Tab 切换 | 切换后状态正确恢复 |
| 会话持久化 | 应用重启后 tmux session 恢复 |
| 性能 | 10 个终端 Tab 内存 < 500MB |

---

## 8. 发布检查清单

- [ ] 所有 TUI 兼容性测试通过
- [ ] 手动测试清单完成
- [ ] 内存测试通过
- [ ] 更新 CHANGELOG.md
- [ ] 创建 Git tag: v0.4.0-alpha
- [ ] GitHub Release 构建

---

## 9. 依赖更新

### 前端依赖

```json
{
  "dependencies": {
    "xterm-addon-serialize": "^0.13.0"
  }
}
```

---

## 10. 时间估算

| 任务 | 时间 |
|------|------|
| Tab 切换优化 | 2 天 |
| 会话持久化 | 2 天 |
| TUI 兼容性测试 | 2 天 |
| 性能优化 | 1 天 |
| Bug 修复 | 2 天 |
| 文档和发布 | 1 天 |
| **总计** | **10 天 (2 周)** |
