import { create } from 'zustand'
import type { AppState, Project, DiffFile } from '../types'

// Mock data for development
const mockProjects: Project[] = [
  {
    id: 'agent-orch',
    name: 'agent-orch',
    path: '/Users/rekles/AI/agent-orch',
    worktrees: [
      {
        id: 'main',
        name: 'main',
        branch: 'main',
        path: '/Users/rekles/AI/agent-orch',
        hasChanges: true,
        terminalTabs: [
          { id: 'tab-1', name: 'claude', status: 'running', agentType: 'claude' },
          { id: 'tab-2', name: 'shell', status: 'idle', agentType: 'shell' },
        ],
      },
      {
        id: 'feature-auth',
        name: 'feature-auth',
        branch: 'feature/auth',
        path: '/Users/rekles/AI/agent-orch/.worktrees/feature-auth',
        hasChanges: false,
        terminalTabs: [
          { id: 'tab-3', name: 'claude', status: 'idle', agentType: 'claude' },
        ],
      },
      {
        id: 'fix-bug',
        name: 'fix-bug',
        branch: 'fix/terminal-bug',
        path: '/Users/rekles/AI/agent-orch/.worktrees/fix-bug',
        hasChanges: true,
        terminalTabs: [],
      },
    ],
  },
  {
    id: 'another-project',
    name: 'another-project',
    path: '/Users/rekles/AI/another-project',
    worktrees: [
      {
        id: 'main-2',
        name: 'main',
        branch: 'main',
        path: '/Users/rekles/AI/another-project',
        hasChanges: false,
        terminalTabs: [],
      },
    ],
  },
]

const mockDiffFiles: DiffFile[] = [
  { path: 'src/App.tsx', additions: 45, deletions: 12, status: 'modified' },
  { path: 'src/components/Layout/TopPanel.tsx', additions: 120, deletions: 3, status: 'added' },
  { path: 'src/stores/appStore.ts', additions: 28, deletions: 0, status: 'added' },
  { path: 'src/types/index.ts', additions: 15, deletions: 5, status: 'modified' },
]

interface AppActions {
  selectWorktree: (id: string) => void
  toggleProjectExpand: (id: string) => void
  toggleTopPanel: () => void
  toggleDiffPanel: () => void
  selectDiffFile: (path: string | null) => void
  selectTerminalTab: (id: string) => void
  getDiffFiles: () => DiffFile[]
  getCurrentWorktree: () => Project['worktrees'][0] | undefined
  selectedDiffFile: string | null
}

export const useAppStore = create<AppState & AppActions>((set, get) => ({
  // State
  projects: mockProjects,
  selectedWorktreeId: 'main',
  expandedProjectIds: ['agent-orch'],
  topPanelCollapsed: false,
  diffPanelOpen: false,
  selectedDiffFile: null,
  activeTerminalTabId: 'tab-1',

  // Actions
  selectWorktree: (id) => {
    const state = get()
    const worktree = state.projects
      .flatMap(p => p.worktrees)
      .find(w => w.id === id)
    set({
      selectedWorktreeId: id,
      activeTerminalTabId: worktree?.terminalTabs[0]?.id || null,
    })
  },

  toggleProjectExpand: (id) => {
    set((state) => ({
      expandedProjectIds: state.expandedProjectIds.includes(id)
        ? state.expandedProjectIds.filter((eid) => eid !== id)
        : [...state.expandedProjectIds, id],
    }))
  },

  toggleTopPanel: () => {
    set((state) => ({ topPanelCollapsed: !state.topPanelCollapsed }))
  },

  toggleDiffPanel: () => {
    set((state) => ({ diffPanelOpen: !state.diffPanelOpen }))
  },

  selectDiffFile: (path) => {
    set({ selectedDiffFile: path, diffPanelOpen: path !== null })
  },

  selectTerminalTab: (id) => {
    set({ activeTerminalTabId: id })
  },

  getDiffFiles: () => mockDiffFiles,

  getCurrentWorktree: () => {
    const state = get()
    return state.projects
      .flatMap(p => p.worktrees)
      .find(w => w.id === state.selectedWorktreeId)
  },
}))
