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
| F0.4 | TailwindCSS 配置 | P0 | ✅ 完成 |
| F0.5 | TypeScript 配置 | P0 | ✅ 完成 |

### 1.2 基础 UI 布局

| Feature | 描述 | 优先级 | 状态 |
|---------|------|--------|------|
| F0.6 | 上下分区布局 (可折叠上方 + Terminal) | P0 | ✅ 完成 |
| F0.7 | 上方面板 (Projects + Diff) | P0 | ✅ 完成 |
| F0.8 | 暗色主题基础样式 (GitHub Dark Dimmed) | P0 | ✅ 完成 |
| F0.9 | Diff 侧边滑出面板 | P1 | ✅ 完成 |

#### UI 布局设计（已确认 2026-03-11）

**整体结构**：上下分区，上方可折叠

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Agent Orchestrator                         [API: ● claude-opus] [⚙️] [▼]  │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─ Projects ─────────────┐  ┌─ Diff ───────────────────────────────────┐   │
│  │                        │  │                                          │   │
│  │  ▼ agent-orch          │  │  ▸ src/app.tsx      +45 -12  [点击展开]  │   │
│  │    ● main              │  │  ▸ lib/auth.ts      +120 -3              │   │
│  │    ○ feature-auth      │  │  ▸ types/user.ts    +28                  │   │
│  │    ○ fix-bug           │  │                                          │   │
│  │    [+ New Worktree]    │  │  点击文件 → 弹出独立 Diff 页面            │   │
│  │                        │  │                                          │   │
│  │  ▶ another-project     │  │  [Accept All] [Reject All]               │   │
│  │    (点击切换)          │  │                                          │   │
│  │                        │  │                                          │   │
│  │  ────────────────────  │  └──────────────────────────────────────────┘   │
│  │  [可折叠] [▲]          │                                                  │
│  └────────────────────────┘                                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─ Terminal ───────────────────────────────────────────────────────────┐   │
│  │  [Tab: main ●] [Tab: feature-auth ○] [Tab: fix-bug ○] [+]            │   │
│  ├──────────────────────────────────────────────────────────────────────┤   │
│  │                                                                      │   │
│  │  $ claude                                                            │   │
│  │  > Working on authentication...                                      │   │
│  │  > [▓▓▓▓▓▓░░░░] 60%                                                 │   │
│  │  >                                                                   │   │
│  │                                                                      │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  底部状态栏: main @ agent-orch | ● 3 changes | API: $2.34 today            │
└─────────────────────────────────────────────────────────────────────────────┘
        上方面板 (可折叠, ~180px)                      下方 Terminal (flex-1)
```

#### 布局优化思考（可选方案）

当前方案以“上下分区 + 可折叠”优先保证终端空间，但在高频切换 Project / Worktree / Diff 时，用户需要反复折叠与视线跳转；在小屏下，纵向分区也容易压缩 Terminal 的可用高度。为提升可用性与降低操作成本，可作为后续迭代考虑以下布局变体（不与已确认方案冲突）：

**方案 A（推荐）：左侧导航 + 主工作区 + 右侧 Inspector**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Agent Orchestrator  [Project ▾] [Worktree ▾] [API ● claude-opus] [⚙️]       │
├─────────── Projects / Worktrees ───────────┬────────── Terminal ────────────┤
│ 🔎 Search                                  │  [Tab: main ●] [Tab: ...]  [+] │
│ ▼ agent-orch                               │  $ claude                      │
│   ● main                                   │  > Working...                  │
│   ○ feature-auth                           │                               │
│   ○ fix-bug                                │                               │
│                                              ├──────── Inspector ─────────┤
│ ▶ another-project                          │  Diff List → File Diff        │
│   (点击切换)                                │  [Accept] [Reject] [Close]    │
├────────────────────────────────────────────┴───────────────────────────────┤
│ Status: main @ agent-orch | ● 3 changes | Agent: claude (running)           │
└─────────────────────────────────────────────────────────────────────────────┘
```

- **好处**：导航位置固定、视线移动更少，减少“折叠-展开”的动作成本；符合 IDE 习惯。
- **交互**：右侧 Inspector 可折叠/拖拽宽度；双击文件进入全屏 Diff；Esc 返回。
- **适配**：当窗口变窄时，Inspector 自动变为底部抽屉（见方案 B）。

**方案 B：主工作区 + 底部抽屉（小屏 / 专注模式）**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Agent Orchestrator  [Project ▾] [Worktree ▾] [API ● claude-opus] [⚙️]       │
├───────── Terminal (flex-1) ────────────────────────────────────────────────┤
│  [Tab: main ●] [Tab: ...] [+]                                               │
│  $ claude                                                                   │
│  > Working...                                                               │
├────────── Diff Drawer (40% height, 可拖拽) ─────────────────────────────────┤
│  ▸ src/app.tsx     +45 -12  [Open]                                          │
│  ▸ lib/auth.ts     +120 -3                                                  │
│  [Accept All] [Reject All] [Close]                                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

- **好处**：保留最大终端高度；小屏下操作更集中。
- **交互**：抽屉高度可拖拽；支持 `Cmd/Ctrl + D` 呼出/关闭。

**响应式规则建议**

1. 宽度 < 1100px：右侧 Inspector 自动变为底部抽屉。
2. 宽度 < 900px：Projects 列表收为图标轨道，提供搜索与最近访问入口。
3. 任何模式：保持 `Project/Worktree` 选择在顶部栏，避免入口被折叠隐藏。

#### 核心设计决策

| 决策 | 说明 |
|------|------|
| **上下分区** | 上方 Projects+Diff，下方 Terminal |
| **上方可折叠** | 点击 [▲] 折叠为 40px，Terminal 全屏 |
| **Terminal Tab 归属 Worktree** | 每个 worktree 有独立的 terminal 上下文 |
| **Diff 侧边滑出** | 点击文件后从右侧滑出面板，不完全遮挡 Terminal |

#### 核心交互逻辑

| 操作 | 效果 |
|------|------|
| 点击 worktree | 切换当前 worktree，Diff 更新，Terminal Tab 切换 |
| 点击 project | 切换 project，展开其 worktree 列表 |
| 点击 Diff 文件 | **侧边滑出** Diff 详情面板 |
| 点击 [▲] | 折叠上方面板，Terminal 全屏 |
| 点击 Terminal Tab | 切换 terminal session |

#### Diff 侧边滑出面板

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Projects │ Diff (折叠) │ Terminal                                     │
│                        │                                                │
│                        │          $ claude                              │
│                        │          > Working...                          │
│                        │                                                │
│                        ├────────────────────────────────────────────────┤
│                        │  ← Diff: src/app.tsx                          ││
│                        │  ┌──────────────────┬──────────────────────┐  ││
│                        │  │ Original         │ Modified            │  ││
│                        │  │ 1 function App() │ 1 function App(...)  │  ││
│                        │  └──────────────────┴──────────────────────┘  ││
│                        │  [Accept] [Reject]                    [×]    ││
│                        └────────────────────────────────────────────────┤
└─────────────────────────────────────────────────────────────────────────────┘
                                  ↑ Diff 从右侧滑出，覆盖部分 Terminal
```

#### Terminal Tab 归属 Worktree

```
当前 worktree: main
  → Tab 1: claude (running)
  → Tab 2: codex (idle)
  → Tab 3: shell

切换到 worktree: feature-auth
  → Tab 1: claude (running)
  → Tab 2: shell
```

**优点**：每个 worktree 有独立的 terminal 上下文，切换时状态保持。

#### 组件结构

```
frontend/src/
├── App.tsx                    # 主容器 (上下分区布局)
├── components/
│   └── Layout/
│       ├── TopPanel.tsx       # 上方面板 (可折叠)
│       ├── ProjectList.tsx    # 左侧 - 项目+worktree 层级列表
│       ├── DiffList.tsx       # 右侧 - 文件变更列表
│       ├── TerminalPane.tsx   # 下方 - 终端区域
│       ├── TerminalTabs.tsx   # Terminal Tab 管理
│       ├── DiffPanel.tsx      # 侧边滑出 Diff 详情
│       └── StatusBar.tsx      # 底部状态栏
├── stores/
│   └── appStore.ts           # projects, selectedWorktree, terminalTabs 状态
└── types/
    └── index.ts              # Project, Worktree, TerminalTab 类型定义
```

#### 状态管理

```typescript
interface AppState {
  projects: Project[]           // 所有项目
  selectedWorktreeId: string    // 当前选中的 worktree
  expandedProjectIds: string[]  // 展开的项目
  topPanelCollapsed: boolean    // 上方面板是否折叠
  diffPanelOpen: boolean        // Diff 侧边面板是否打开
  selectedDiffFile: string      // 选中的 diff 文件
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
  terminalTabs: TerminalTab[]  // 该 worktree 的 terminal tabs
}

interface TerminalTab {
  id: string
  name: string
  status: 'running' | 'idle'
  agentType?: 'claude' | 'codex' | 'cursor' | 'shell'
}
```

#### 键盘快捷键

| 快捷键 | 功能 |
|--------|------|
| `Cmd/Ctrl + T` | 新建 Terminal Tab |
| `Cmd/Ctrl + W` | 关闭当前 Tab |
| `Cmd/Ctrl + 1-9` | 切换 Tab |
| `Cmd/Ctrl + P` | 切换 Project |
| `Cmd/Ctrl + D` | 切换 Diff 面板 |
| `Cmd/Ctrl + B` | 折叠/展开上方面板 |

#### 配色方案（GitHub Dark Dimmed）

```css
:root {
  --bg-base: #22272e;        /* 主背景 */
  --bg-surface: #2d333b;     /* 面板背景 */
  --border: #373e47;         /* 边框 */
  --text: #adbac7;           /* 主文字 */
  --text-muted: #768390;     /* 次要文字 */
  --accent: #6cb6ff;         /* 强调色（蓝） */
  --success: #57ab5a;        /* 新增（绿） */
  --warning: #c69026;        /* 修改（黄） */
  --error: #f47067;          /* 删除（红） */
}
```

#### 动画与过渡

| 交互 | 动画 |
|------|------|
| 折叠上方面板 | `height: 180px → 40px` (200ms ease) |
| Diff 侧边滑出 | `transform: translateX(100% → 0)` (150ms) |
| 切换 worktree | 内容淡入淡出 (100ms) |

---

#### UI 实现 (2026-03-15)

**技术栈**：
- React 18 + TypeScript
- TailwindCSS 3.4 (自定义 GitHub Dark Dimmed 主题)
- Zustand 5 (状态管理)
- Wails v2 (Go 后端 + WebView)

**已实现组件**：

| 组件 | 文件 | 功能 |
|------|------|------|
| TopPanel | `components/Layout/TopPanel.tsx` | 上方面板，包含 Projects + Diff 列表，可折叠 |
| TerminalPane | `components/Layout/TerminalPane.tsx` | 终端区域，Tab 管理，状态指示 |
| DiffPanel | `components/Layout/DiffPanel.tsx` | 侧边滑出 Diff 详情面板 |
| StatusBar | `components/Layout/StatusBar.tsx` | 底部状态栏 |

**CSS 增强**：

```css
/* 自定义动画 */
@keyframes fadeIn { ... }
@keyframes slideInRight { ... }
@keyframes pulseSoft { ... }
@keyframes statusPulse { ... }

/* 工具类 */
.animate-fade-in { ... }
.animate-slide-in-right { ... }
.animate-pulse-soft { ... }
.glass-panel { ... }
.hover-glow { ... }
.active-scale { ... }
```

**交互细节**：

1. **Projects 列表**
   - 树形展开/折叠动画
   - 选中态左边框高亮
   - 变更状态脉冲指示器 (hasChanges)
   - hover 背景变化

2. **Diff 列表**
   - 文件类型图标
   - +/- 行数统计
   - 点击触发侧边面板
   - 动态文件数量徽章

3. **Terminal Tab**
   - 运行状态脉冲指示器 (animate-ping)
   - Agent 类型标签
   - 进度条渐变效果
   - 光标闪烁动画

4. **DiffPanel**
   - 背景毛玻璃遮罩 (backdrop-blur)
   - 滑入动画 (slideInRight)
   - 统一 Diff 视图（并排显示）
   - 键盘快捷键提示

5. **StatusBar**
   - 分支图标
   - 变更状态徽章
   - Agent 运行状态
   - API 费用显示
   - 实时时钟

**状态管理**：

```typescript
// stores/appStore.ts
interface AppState {
  projects: Project[]
  selectedWorktreeId: string | null
  expandedProjectIds: string[]
  topPanelCollapsed: boolean
  diffPanelOpen: boolean
  selectedDiffFile: string | null
  activeTerminalTabId: string | null
}

// Actions
selectWorktree(id: string): void
toggleProjectExpand(id: string): void
toggleTopPanel(): void
selectDiffFile(path: string | null): void
selectTerminalTab(id: string): void
```

**文件结构**：

```
frontend/src/
├── App.tsx                    # 主容器
├── main.tsx                   # 入口
├── index.css                  # 全局样式 + 动画
├── components/
│   └── Layout/
│       ├── TopPanel.tsx       # 上方面板 (Projects + Diff)
│       ├── TerminalPane.tsx   # 终端区域
│       ├── DiffPanel.tsx      # Diff 侧边面板
│       └── StatusBar.tsx      # 底部状态栏
├── stores/
│   └── appStore.ts            # Zustand store
└── types/
    └── index.ts               # TypeScript 类型定义
```

---

#### 改进建议 (未来迭代参考)

以下是基于 UI/UX 分析的改进建议，供未来迭代参考：

**1. 顶部栏增强**
- 增加快速切换 Tab: `[▶ Projects] [▶ Diff]`
- 便于用户快速定位当前视图

**2. 状态栏信息增强**
- 增加 Agent 状态显示: `Agent: claude (running)`
- 让用户了解当前 AI Agent 工作状态

**3. Diff 面板交互优化**
- 面板宽度可拖拽调整
- 双击文件 → 全屏 Diff 视图
- 键盘导航: `j/k` 上下, `Enter` 打开, `Esc` 关闭
- Accept/Reject 后自动跳到下一个文件

**4. Projects 列表增强**
- 搜索过滤框 (当项目多时)
- 按最近使用排序
- 收藏/置顶常用项目
- 右键菜单 (New Worktree, Open in Finder, Settings)

**5. 上方面板比例**
- 支持拖拽调整 Projects/Diff 比例
- 记住用户偏好

**6. 动画细节**
- Diff 侧边滑出时添加背景遮罩 (opacity 0→0.5)
- 终端 Tab 切换添加滑动指示器动画

---

#### 历史方案 (已否决)

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
| F0.14 | components/Layout/ | P0 | ✅ 完成 |
| F0.15 | hooks/ 目录 | P0 | ⏸️ 推迟 |
| F0.16 | stores/appStore.ts | P0 | ✅ 完成 |
| F0.17 | types/index.ts | P0 | ✅ 完成 |

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
- [x] 上下分区布局显示正常
- [x] 暗色主题渲染正确 (GitHub Dark Dimmed)
- [x] 上方面板可折叠
- [x] Projects 列表展开/折叠
- [x] Worktree 选中态
- [x] Diff 列表显示
- [x] Diff 侧边面板滑出动画
- [x] Terminal Tab 切换
- [x] StatusBar 信息显示
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
| TailwindCSS | 样式框架配置 | ✅ 通过 |
| 上下分区布局 | 基础 UI 布局 | ✅ 通过 |
| 暗色主题 | GitHub Dark Dimmed 配色 | ✅ 通过 |
| Diff 侧边面板 | 滑出动画、交互 | ✅ 通过 |
| 状态管理 | Zustand store + 类型 | ✅ 通过 |

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
