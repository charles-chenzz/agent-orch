export interface Project {
  id: string
  name: string
  path: string
  worktrees: Worktree[]
}

export interface Worktree {
  id: string
  name: string
  branch: string
  path: string
  hasChanges: boolean
  terminalTabs: TerminalTab[]
}

export interface TerminalTab {
  id: string
  name: string
  status: 'running' | 'idle'
  agentType?: 'claude' | 'codex' | 'cursor' | 'shell'
}

export interface DiffFile {
  path: string
  additions: number
  deletions: number
  status: 'modified' | 'added' | 'deleted' | 'renamed'
}

export interface AppState {
  projects: Project[]
  selectedWorktreeId: string | null
  expandedProjectIds: string[]
  topPanelCollapsed: boolean
  diffPanelOpen: boolean
  selectedDiffFile: string | null
  activeTerminalTabId: string | null
}
