import { create } from 'zustand'
import type { AppState, Project, DiffFile } from '../types'
import { ListWorktrees } from '../../wailsjs/go/main/App'

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
  loadWorktrees: () => Promise<void>
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
  loadWorktrees: async () => {
    try {
      const worktrees = await ListWorktrees()
      const realWorktrees = (worktrees || []).map(wt => ({
        id: wt.id,
        name: wt.name,
        branch: wt.branch,
        path: wt.path,
        hasChanges: wt.hasChanges,
        terminalTabs: [],
      }))

      set((state) => {
        const mainWorktreePath = worktrees?.find(wt => wt.isMain)?.path
        const fallbackPath = worktrees?.[0]?.path
        const currentProject = state.projects.find(p => p.id === 'agent-orch')
        const nextProjectPath = mainWorktreePath || fallbackPath || currentProject?.path || ''

        const hasTargetProject = state.projects.some(p => p.id === 'agent-orch')
        const projects = hasTargetProject
          ? state.projects.map(p => p.id === 'agent-orch'
            ? { ...p, path: nextProjectPath, worktrees: realWorktrees }
            : p)
          : [{
            id: 'agent-orch',
            name: 'agent-orch',
            path: nextProjectPath,
            worktrees: realWorktrees,
          }, ...state.projects]

        const selectedStillExists = state.selectedWorktreeId != null
          && realWorktrees.some(wt => wt.id === state.selectedWorktreeId)

        return {
          projects,
          selectedWorktreeId: selectedStillExists ? state.selectedWorktreeId : (realWorktrees[0]?.id ?? null),
          activeTerminalTabId: null,
        }
      })
    } catch {
      // Keep mock data when runtime is unavailable.
    }
  },

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
