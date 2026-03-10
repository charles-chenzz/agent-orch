# Phase 5: Agent Monitor + GitHub Integration

> **周期**：2 周
> **目标**：Agent 进程监控 + GitHub PR 集成 + Diff 查看器
> **依赖**：Phase 0-4
> **交付**：v1.0.0-rc1

---

## 1. Feature List

### 1.1 Agent 进程检测

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F5.1 | Claude Code 进程检测 | P0 |
| F5.2 | Codex 进程检测 | P0 |
| F5.3 | Cursor 进程检测 | P1 |
| F5.4 | 进程状态轮询 | P0 |
| F5.5 | 状态变化事件 | P0 |

### 1.2 Agent 状态监控

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F5.6 | 工作/等待/空闲状态 | P0 |
| F5.7 | 状态指示器 UI | P0 |
| F5.8 | 活动历史记录 | P2 |
| F5.9 | 多 Agent 同时运行 | P1 |

### 1.3 GitHub 集成

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F5.10 | GitHub Token 配置 | P0 |
| F5.11 | 当前分支 PR 检测 | P0 |
| F5.12 | PR 列表展示 | P0 |
| F5.13 | PR 状态（open/merged/draft） | P1 |
| F5.14 | PR 链接跳转 | P0 |

### 1.4 Diff 查看器

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F5.15 | 变更文件列表 | P0 |
| F5.16 | 并排 Diff 显示 | P0 |
| F5.17 | 语法高亮 | P1 |
| F5.18 | 统一 Diff 格式 | P2 |

---

## 2. 实现细节

### 2.1 Go 后端 - Agent 检测

```go
// internal/agent/types.go
package agent

import "time"

type AgentType string

const (
    AgentClaudeCode AgentType = "claude-code"
    AgentCodex      AgentType = "codex"
    AgentCursor     AgentType = "cursor"
    AgentOpenCode   AgentType = "opencode"
)

type AgentStatus string

const (
    StatusIdle    AgentStatus = "idle"
    StatusWorking AgentStatus = "working"
    StatusWaiting AgentStatus = "waiting"
)

type AgentInfo struct {
    Type       AgentType   `json:"type"`
    PID        int         `json:"pid"`
    WorktreeID string      `json:"worktreeId"`
    Status     AgentStatus `json:"status"`
    StartTime  time.Time   `json:"startTime"`
    LastActive time.Time   `json:"lastActive"`
    Command    string      `json:"command"`
}

type AgentEvent struct {
    Type      string     `json:"type"`      // started, stopped, status_changed
    Agent     AgentInfo  `json:"agent"`
    Timestamp time.Time  `json:"timestamp"`
}
```

```go
// internal/agent/detector.go
package agent

import (
    "bytes"
    "context"
    "os"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "time"
    
    "github.com/shirou/gopsutil/v3/process"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Detector struct {
    ctx       context.Context
    agents    map[int]*AgentInfo
    mu        sync.RWMutex
    interval  time.Duration
    stopChan  chan struct{}
}

func NewDetector(ctx context.Context) *Detector {
    return &Detector{
        ctx:      ctx,
        agents:   make(map[int]*AgentInfo),
        interval: 2 * time.Second,
        stopChan: make(chan struct{}),
    }
}

// Start 开始检测
func (d *Detector) Start() {
    ticker := time.NewTicker(d.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            d.scan()
        case <-d.stopChan:
            return
        case <-d.ctx.Done():
            return
        }
    }
}

// Stop 停止检测
func (d *Detector) Stop() {
    close(d.stopChan)
}

// scan 扫描所有进程
func (d *Detector) scan() {
    procs, err := process.Processes()
    if err != nil {
        return
    }
    
    currentPID := os.Getpid()
    foundPIDs := make(map[int]bool)
    
    for _, proc := range procs {
        pid := int(proc.Pid)
        if pid == currentPID {
            continue
        }
        
        // 获取进程命令行
        cmdline, err := proc.Cmdline()
        if err != nil {
            continue
        }
        
        // 检测是否是 AI Agent
        agentType := d.detectAgentType(cmdline)
        if agentType == "" {
            continue
        }
        
        foundPIDs[pid] = true
        
        // 获取工作目录
        cwd, _ := proc.Cwd()
        
        d.mu.Lock()
        existing, exists := d.agents[pid]
        
        if !exists {
            // 新发现的 Agent
            info := &AgentInfo{
                Type:       agentType,
                PID:        pid,
                WorktreeID: cwd, // 简化：使用 cwd 作为 worktree 标识
                Status:     StatusWorking,
                StartTime:  time.Now(),
                LastActive: time.Now(),
                Command:    cmdline,
            }
            d.agents[pid] = info
            
            d.mu.Unlock()
            
            // 发送事件
            d.emitEvent("started", *info)
        } else {
            // 更新已有 Agent
            existing.LastActive = time.Now()
            existing.Command = cmdline
            d.mu.Unlock()
        }
    }
    
    // 检查已停止的 Agent
    d.mu.Lock()
    for pid, agent := range d.agents {
        if !foundPIDs[pid] {
            delete(d.agents, pid)
            d.mu.Unlock()
            d.emitEvent("stopped", *agent)
            d.mu.Lock()
        }
    }
    d.mu.Unlock()
}

// detectAgentType 检测 Agent 类型
func (d *Detector) detectAgentType(cmdline string) AgentType {
    cmdlineLower := strings.ToLower(cmdline)
    
    // Claude Code
    if strings.Contains(cmdlineLower, "claude") && 
       (strings.Contains(cmdlineLower, "code") || strings.Contains(cmdline, "@anthropic-ai/claude-code")) {
        return AgentClaudeCode
    }
    
    // Codex (OpenAI)
    if strings.Contains(cmdlineLower, "codex") && 
       strings.Contains(cmdlineLower, "openai") {
        return AgentCodex
    }
    
    // Cursor
    if strings.Contains(cmdlineLower, "cursor") && 
       strings.Contains(cmdlineLower, "electron") {
        return AgentCursor
    }
    
    // OpenCode
    if strings.Contains(cmdlineLower, "opencode") {
        return AgentOpenCode
    }
    
    return ""
}

// emitEvent 发送事件到前端
func (d *Detector) emitEvent(eventType string, agent AgentInfo) {
    event := AgentEvent{
        Type:      eventType,
        Agent:     agent,
        Timestamp: time.Now(),
    }
    runtime.EventsEmit(d.ctx, "agent-event", event)
}

// GetAgents 获取当前检测到的所有 Agent
func (d *Detector) GetAgents() []AgentInfo {
    d.mu.RLock()
    defer d.mu.RUnlock()
    
    result := make([]AgentInfo, 0, len(d.agents))
    for _, agent := range d.agents {
        result = append(result, *agent)
    }
    return result
}

// GetAgentsByWorktree 获取指定 worktree 的 Agent
func (d *Detector) GetAgentsByWorktree(worktreeID string) []AgentInfo {
    d.mu.RLock()
    defer d.mu.RUnlock()
    
    var result []AgentInfo
    for _, agent := range d.agents {
        if strings.HasPrefix(agent.WorktreeID, worktreeID) {
            result = append(result, *agent)
        }
    }
    return result
}

// UpdateStatus 更新 Agent 状态
func (d *Detector) UpdateStatus(pid int, status AgentStatus) {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    if agent, ok := d.agents[pid]; ok {
        oldStatus := agent.Status
        agent.Status = status
        agent.LastActive = time.Now()
        
        if oldStatus != status {
            d.emitEvent("status_changed", *agent)
        }
    }
}
```

### 2.2 Go 后端 - GitHub 集成

```go
// internal/github/client.go
package github

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/google/go-github/v62/github"
    "golang.org/x/oauth2"
)

type Client struct {
    client *github.Client
    ctx    context.Context
}

type PullRequest struct {
    Number    int    `json:"number"`
    Title     string `json:"title"`
    State     string `json:"state"`
    Draft     bool   `json:"draft"`
    URL       string `json:"url"`
    Author    string `json:"author"`
    Branch    string `json:"branch"`
    BaseRef   string `json:"baseRef"`
    CreatedAt string `json:"createdAt"`
    UpdatedAt string `json:"updatedAt"`
}

func NewClient(token string) *Client {
    ctx := context.Background()
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: token},
    )
    tc := oauth2.NewClient(ctx, ts)
    
    return &Client{
        client: github.NewClient(tc),
        ctx:    ctx,
    }
}

// GetPullRequestForBranch 获取指定分支的 PR
func (c *Client) GetPullRequestForBranch(owner, repo, branch string) (*PullRequest, error) {
    // 列出所有打开的 PR
    prs, _, err := c.client.PullRequests.List(c.ctx, owner, repo, &github.PullRequestListOptions{
        State: "all",
        Head:  fmt.Sprintf("%s:%s", owner, branch),
        ListOptions: github.ListOptions{
            PerPage: 10,
        },
    })
    if err != nil {
        return nil, err
    }
    
    for _, pr := range prs {
        if pr.GetHead().GetRef() == branch {
            return &PullRequest{
                Number:    pr.GetNumber(),
                Title:     pr.GetTitle(),
                State:     pr.GetState(),
                Draft:     pr.GetDraft(),
                URL:       pr.GetHTMLURL(),
                Author:    pr.GetUser().GetLogin(),
                Branch:    pr.GetHead().GetRef(),
                BaseRef:   pr.GetBase().GetRef(),
                CreatedAt: pr.GetCreatedAt().Format("2006-01-02"),
                UpdatedAt: pr.GetUpdatedAt().Format("2006-01-02"),
            }, nil
        }
    }
    
    return nil, nil
}

// ListPullRequests 列出所有 PR
func (c *Client) ListPullRequests(owner, repo string) ([]PullRequest, error) {
    prs, _, err := c.client.PullRequests.List(c.ctx, owner, repo, &github.PullRequestListOptions{
        State: "open",
        ListOptions: github.ListOptions{
            PerPage: 50,
        },
    })
    if err != nil {
        return nil, err
    }
    
    var result []PullRequest
    for _, pr := range prs {
        result = append(result, PullRequest{
            Number:    pr.GetNumber(),
            Title:     pr.GetTitle(),
            State:     pr.GetState(),
            Draft:     pr.GetDraft(),
            URL:       pr.GetHTMLURL(),
            Author:    pr.GetUser().GetLogin(),
            Branch:    pr.GetHead().GetRef(),
            BaseRef:   pr.GetBase().GetRef(),
            CreatedAt: pr.GetCreatedAt().Format("2006-01-02"),
            UpdatedAt: pr.GetUpdatedAt().Format("2006-01-02"),
        })
    }
    
    return result, nil
}

// ParseRemoteURL 解析 remote URL 获取 owner/repo
func ParseRemoteURL(url string) (owner, repo string, err error) {
    // https://github.com/owner/repo.git
    // git@github.com:owner/repo.git
    url = strings.TrimSuffix(url, ".git")
    
    if strings.HasPrefix(url, "https://github.com/") {
        parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
        if len(parts) >= 2 {
            return parts[0], parts[1], nil
        }
    }
    
    if strings.HasPrefix(url, "git@github.com:") {
        parts := strings.Split(strings.TrimPrefix(url, "git@github.com:"), "/")
        if len(parts) >= 2 {
            return parts[0], parts[1], nil
        }
    }
    
    return "", "", fmt.Errorf("invalid GitHub URL: %s", url)
}
```

### 2.3 Go 后端 - Diff 查看器

```go
// internal/diff/viewer.go
package diff

import (
    "bufio"
    "bytes"
    "fmt"
    "os/exec"
    "strings"
)

type FileDiff struct {
    Path     string     `json:"path"`
    OldPath  string     `json:"oldPath"`  // 重命名时
    NewPath  string     `json:"newPath"`
    Status   string     `json:"status"`   // A, M, D, R
    Hunks    []DiffHunk `json:"hunks"`
    Additions int       `json:"additions"`
    Deletions int       `json:"deletions"`
}

type DiffHunk struct {
    OldStart int          `json:"oldStart"`
    OldLines int          `json:"oldLines"`
    NewStart int          `json:"newStart"`
    NewLines int          `json:"newLines"`
    Header   string       `json:"header"`
    Lines    []DiffLine   `json:"lines"`
}

type DiffLine struct {
    Type     string `json:"type"` // context, add, delete
    Content  string `json:"content"`
    OldNum   int    `json:"oldNum,omitempty"`
    NewNum   int    `json:"newNum,omitempty"`
}

// GetDiff 获取指定 worktree 的 diff
func GetDiff(repoPath string, staged bool) ([]FileDiff, error) {
    var cmd *exec.Cmd
    
    if staged {
        cmd = exec.Command("git", "diff", "--cached", "--unified=3")
    } else {
        cmd = exec.Command("git", "diff", "--unified=3")
    }
    cmd.Dir = repoPath
    
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    return parseDiff(output)
}

// GetDiffForFile 获取单个文件的 diff
func GetDiffForFile(repoPath, filePath string) (*FileDiff, error) {
    cmd := exec.Command("git", "diff", "--unified=3", "--", filePath)
    cmd.Dir = repoPath
    
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    diffs, err := parseDiff(output)
    if err != nil {
        return nil, err
    }
    
    if len(diffs) > 0 {
        return &diffs[0], nil
    }
    
    return nil, nil
}

// parseDiff 解析 git diff 输出
func parseDiff(output []byte) ([]FileDiff, error) {
    var diffs []FileDiff
    var currentDiff *FileDiff
    var currentHunk *DiffHunk
    
    scanner := bufio.NewScanner(bytes.NewReader(output))
    oldLineNum := 0
    newLineNum := 0
    
    for scanner.Scan() {
        line := scanner.Text()
        
        // 新文件开始
        if strings.HasPrefix(line, "diff --git ") {
            if currentDiff != nil {
                if currentHunk != nil {
                    currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
                }
                diffs = append(diffs, *currentDiff)
            }
            
            currentDiff = &FileDiff{}
            currentHunk = nil
            
            // 解析文件路径
            parts := strings.SplitN(line, " ", 4)
            if len(parts) >= 4 {
                // a/path/to/file b/path/to/file
                currentDiff.OldPath = strings.TrimPrefix(parts[2], "a/")
                currentDiff.NewPath = strings.TrimPrefix(parts[3], "b/")
                currentDiff.Path = currentDiff.NewPath
            }
            continue
        }
        
        if currentDiff == nil {
            continue
        }
        
        // 文件状态
        if strings.HasPrefix(line, "new file mode ") {
            currentDiff.Status = "A"
            continue
        }
        if strings.HasPrefix(line, "deleted file mode ") {
            currentDiff.Status = "D"
            continue
        }
        if strings.HasPrefix(line, "rename from ") {
            currentDiff.Status = "R"
            continue
        }
        
        // Hunk header
        if strings.HasPrefix(line, "@@ ") {
            if currentHunk != nil {
                currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
            }
            
            currentHunk = &DiffHunk{}
            fmt.Sscanf(line, "@@ -%d,%d +%d,%d @@", 
                &currentHunk.OldStart, &currentHunk.OldLines,
                &currentHunk.NewStart, &currentHunk.NewLines)
            currentHunk.Header = line
            
            oldLineNum = currentHunk.OldStart
            newLineNum = currentHunk.NewStart
            continue
        }
        
        if currentHunk == nil {
            continue
        }
        
        // Diff lines
        diffLine := DiffLine{Content: line}
        
        if strings.HasPrefix(line, "+") {
            diffLine.Type = "add"
            diffLine.NewNum = newLineNum
            newLineNum++
            currentDiff.Additions++
        } else if strings.HasPrefix(line, "-") {
            diffLine.Type = "delete"
            diffLine.OldNum = oldLineNum
            oldLineNum++
            currentDiff.Deletions++
        } else if strings.HasPrefix(line, " ") {
            diffLine.Type = "context"
            diffLine.OldNum = oldLineNum
            diffLine.NewNum = newLineNum
            oldLineNum++
            newLineNum++
        } else {
            continue
        }
        
        currentHunk.Lines = append(currentHunk.Lines, diffLine)
    }
    
    // 添加最后一个
    if currentDiff != nil {
        if currentHunk != nil {
            currentDiff.Hunks = append(currentDiff.Hunks, *currentHunk)
        }
        diffs = append(diffs, *currentDiff)
    }
    
    return diffs, scanner.Err()
}

// ListChangedFiles 列出变更文件
func ListChangedFiles(repoPath string) ([]string, error) {
    cmd := exec.Command("git", "status", "--porcelain")
    cmd.Dir = repoPath
    
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    var files []string
    scanner := bufio.NewScanner(bytes.NewReader(output))
    for scanner.Scan() {
        line := scanner.Text()
        if len(line) > 3 {
            files = append(files, strings.TrimSpace(line[3:]))
        }
    }
    
    return files, nil
}
```

### 2.4 App.go 绑定

```go
// app.go (新增)

// === Agent 检测 ===
func (a *App) GetAgents() []agent.AgentInfo {
    return a.agentDetector.GetAgents()
}

func (a *App) GetAgentsForWorktree(worktreeId string) []agent.AgentInfo {
    return a.agentDetector.GetAgentsByWorktree(worktreeId)
}

// === GitHub 集成 ===
func (a *App) GetPullRequest(branch string) (*github.PullRequest, error) {
    if a.githubClient == nil {
        return nil, fmt.Errorf("GitHub not configured")
    }
    
    // 获取 remote URL
    owner, repo, err := a.getRepoInfo()
    if err != nil {
        return nil, err
    }
    
    return a.githubClient.GetPullRequestForBranch(owner, repo, branch)
}

func (a *App) ListPullRequests() ([]github.PullRequest, error) {
    if a.githubClient == nil {
        return nil, fmt.Errorf("GitHub not configured")
    }
    
    owner, repo, err := a.getRepoInfo()
    if err != nil {
        return nil, err
    }
    
    return a.githubClient.ListPullRequests(owner, repo)
}

func (a *App) getRepoInfo() (string, string, error) {
    // 从 git remote 获取
    cmd := exec.Command("git", "remote", "get-url", "origin")
    cmd.Dir = a.worktree.GetMainPath()
    output, err := cmd.Output()
    if err != nil {
        return "", "", err
    }
    
    return github.ParseRemoteURL(strings.TrimSpace(string(output)))
}

// === Diff 查看器 ===
func (a *App) GetDiff(worktreeId string, staged bool) ([]diff.FileDiff, error) {
    wt, err := a.worktree.GetWorktree(worktreeId)
    if err != nil {
        return nil, err
    }
    return diff.GetDiff(wt.Path, staged)
}

func (a *App) GetFileDiff(worktreeId, filePath string) (*diff.FileDiff, error) {
    wt, err := a.worktree.GetWorktree(worktreeId)
    if err != nil {
        return nil, err
    }
    return diff.GetDiffForFile(wt.Path, filePath)
}

func (a *App) ListChangedFiles(worktreeId string) ([]string, error) {
    wt, err := a.worktree.GetWorktree(worktreeId)
    if err != nil {
        return nil, err
    }
    return diff.ListChangedFiles(wt.Path)
}
```

### 2.5 前端 - 组件

```tsx
// frontend/src/components/AgentMonitor/StatusIndicator.tsx
import { useEffect, useState } from 'react'
import { AgentInfo, AgentEvent } from '../../types/agent'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { GetAgentsForWorktree } from '../../wailsjs/go/main/App'

interface StatusIndicatorProps {
  worktreeId: string
}

export default function StatusIndicator({ worktreeId }: StatusIndicatorProps) {
  const [agents, setAgents] = useState<AgentInfo[]>([])
  
  useEffect(() => {
    // 初始加载
    loadAgents()
    
    // 监听事件
    EventsOn('agent-event', (event: AgentEvent) => {
      loadAgents()
    })
    
    return () => {
      EventsOff('agent-event')
    }
  }, [worktreeId])
  
  const loadAgents = async () => {
    const result = await GetAgentsForWorktree(worktreeId)
    setAgents(result || [])
  }
  
  if (agents.length === 0) {
    return null
  }
  
  return (
    <div className="flex items-center gap-2">
      {agents.map(agent => (
        <div
          key={agent.pid}
          className="flex items-center gap-1 px-2 py-1 rounded bg-gray-800"
          title={`${agent.type} (PID: ${agent.pid})`}
        >
          {/* 状态指示灯 */}
          <span
            className={`
              w-2 h-2 rounded-full
              ${agent.status === 'working' ? 'bg-green-500 animate-pulse' : ''}
              ${agent.status === 'waiting' ? 'bg-yellow-500' : ''}
              ${agent.status === 'idle' ? 'bg-gray-500' : ''}
            `}
          />
          
          {/* Agent 类型图标 */}
          <AgentIcon type={agent.type} className="w-4 h-4" />
          
          <span className="text-xs text-gray-400">{agent.type}</span>
        </div>
      ))}
    </div>
  )
}

function AgentIcon({ type, className }: { type: string; className?: string }) {
  switch (type) {
    case 'claude-code':
      return <ClaudeIcon className={className} />
    case 'codex':
      return <OpenAIIcon className={className} />
    case 'cursor':
      return <CursorIcon className={className} />
    default:
      return <BotIcon className={className} />
  }
}
```

```tsx
// frontend/src/components/Diff/DiffViewer.tsx
import { useEffect, useState } from 'react'
import { FileDiff } from '../../types/diff'
import { GetDiff } from '../../wailsjs/go/main/App'

interface DiffViewerProps {
  worktreeId: string
  staged?: boolean
}

export default function DiffViewer({ worktreeId, staged = false }: DiffViewerProps) {
  const [diffs, setDiffs] = useState<FileDiff[]>([])
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  
  useEffect(() => {
    loadDiff()
  }, [worktreeId, staged])
  
  const loadDiff = async () => {
    setLoading(true)
    try {
      const result = await GetDiff(worktreeId, staged)
      setDiffs(result || [])
    } catch (err) {
      console.error('Failed to load diff:', err)
    }
    setLoading(false)
  }
  
  if (loading) {
    return <div className="p-4 text-gray-500">Loading...</div>
  }
  
  if (diffs.length === 0) {
    return (
      <div className="p-4 text-center text-gray-500">
        No changes
      </div>
    )
  }
  
  const selectedDiff = diffs.find(d => d.path === selectedFile) || diffs[0]
  
  return (
    <div className="flex h-full">
      {/* 文件列表 */}
      <div className="w-48 border-r border-gray-700 overflow-auto">
        {diffs.map(diff => (
          <div
            key={diff.path}
            className={`
              px-3 py-2 cursor-pointer text-sm
              ${selectedDiff?.path === diff.path ? 'bg-gray-700' : 'hover:bg-gray-800'}
            `}
            onClick={() => setSelectedFile(diff.path)}
          >
            <div className="flex items-center gap-2">
              <span className={`
                ${diff.status === 'A' ? 'text-green-500' : ''}
                ${diff.status === 'M' ? 'text-yellow-500' : ''}
                ${diff.status === 'D' ? 'text-red-500' : ''}
              `}>
                {diff.status}
              </span>
              <span className="truncate">{diff.path}</span>
            </div>
            <div className="text-xs text-gray-500 mt-0.5">
              +{diff.additions} -{diff.deletions}
            </div>
          </div>
        ))}
      </div>
      
      {/* Diff 内容 */}
      <div className="flex-1 overflow-auto font-mono text-sm">
        {selectedDiff?.hunks.map((hunk, hunkIdx) => (
          <div key={hunkIdx}>
            {/* Hunk header */}
            <div className="px-4 py-1 bg-gray-800 text-gray-400 text-xs sticky top-0">
              {hunk.header}
            </div>
            
            {/* Lines */}
            {hunk.lines.map((line, lineIdx) => (
              <div
                key={lineIdx}
                className={`
                  flex
                  ${line.type === 'add' ? 'bg-green-900/30' : ''}
                  ${line.type === 'delete' ? 'bg-red-900/30' : ''}
                `}
              >
                {/* 行号 */}
                <div className="w-10 text-right pr-2 text-gray-600 select-none border-r border-gray-700">
                  {line.oldNum || ''}
                </div>
                <div className="w-10 text-right pr-2 text-gray-600 select-none border-r border-gray-700">
                  {line.newNum || ''}
                </div>
                
                {/* 内容 */}
                <pre className="flex-1 px-2 whitespace-pre overflow-x-auto">
                  <span className={`
                    ${line.type === 'add' ? 'text-green-400' : ''}
                    ${line.type === 'delete' ? 'text-red-400' : ''}
                  `}>
                    {line.content}
                  </span>
                </pre>
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}
```

```tsx
// frontend/src/components/GitHub/PRCard.tsx
import { PullRequest } from '../../types/github'

interface PRCardProps {
  pr: PullRequest
}

export default function PRCard({ pr }: PRCardProps) {
  return (
    <a
      href={pr.url}
      target="_blank"
      rel="noopener noreferrer"
      className="block p-3 bg-gray-800 rounded border border-gray-700 hover:border-gray-600 transition-colors"
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-medium truncate">{pr.title}</span>
            {pr.draft && (
              <span className="px-1.5 py-0.5 bg-gray-700 text-gray-400 text-xs rounded">
                Draft
              </span>
            )}
          </div>
          <div className="flex items-center gap-2 mt-1 text-sm text-gray-500">
            <span>#{pr.number}</span>
            <span>•</span>
            <span>{pr.branch} → {pr.baseRef}</span>
          </div>
        </div>
        
        {/* 状态 */}
        <span
          className={`
            px-2 py-0.5 text-xs rounded
            ${pr.state === 'open' ? 'bg-green-900 text-green-300' : ''}
            ${pr.state === 'closed' ? 'bg-red-900 text-red-300' : ''}
            ${pr.state === 'merged' ? 'bg-purple-900 text-purple-300' : ''}
          `}
        >
          {pr.state}
        </span>
      </div>
    </a>
  )
}
```

---

## 3. 测试计划

### 3.1 单元测试

- [ ] Agent 类型检测
- [ ] Diff 解析
- [ ] GitHub URL 解析

### 3.2 集成测试

- [ ] 启动 Claude Code，验证检测到
- [ ] 创建 PR，验证能获取
- [ ] 修改文件，验证 diff 正确

### 3.3 手动测试清单

- [ ] Agent 状态指示器显示正确
- [ ] 点击 PR 卡片能打开浏览器
- [ ] Diff 查看器显示正确
- [ ] 文件列表点击切换正常

---

## 4. 验收标准

| 标准 | 描述 |
|------|------|
| Agent 检测 | 能检测 Claude Code 运行 |
| PR 检测 | 能获取当前分支的 PR |
| Diff 显示 | 能显示文件变更 |
| UI 集成 | 状态指示器在 Sidebar 显示 |

---

## 5. 发布检查清单

- [ ] 所有测试通过
- [ ] 手动测试完成
- [ ] 更新 CHANGELOG.md
- [ ] 创建 Git tag: v1.0.0-rc1
- [ ] GitHub Release 构建

---

## 6. 依赖更新

### Go 依赖

```go
require (
    github.com/shirou/gopsutil/v3 v3.24.0
    github.com/google/go-github/v62 v62.0.0
    golang.org/x/oauth2 v0.18.0
)
```

---

## 7. 时间估算

| 任务 | 时间 |
|------|------|
| Agent 检测实现 | 2 天 |
| GitHub 集成 | 2 天 |
| Diff 查看器 | 2 天 |
| 前端组件 | 2 天 |
| 测试 | 1 天 |
| 文档和发布 | 1 天 |
| **总计** | **10 天 (2 周)** |
