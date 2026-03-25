import { beforeEach, describe, expect, it, vi, type Mock } from 'vitest'

const destroySessionsByWorktree = vi.fn()

vi.mock('../../../wailsjs/go/main/App', () => ({
  ListWorktrees: vi.fn(),
  CreateWorktree: vi.fn(),
  DeleteWorktree: vi.fn(),
}))

vi.mock('../terminalStore', () => ({
  useTerminalStore: {
    getState: vi.fn(() => ({ destroySessionsByWorktree })),
  },
}))

import { CreateWorktree, DeleteWorktree, ListWorktrees } from '../../../wailsjs/go/main/App'
import { useAppStore } from '../appStore'

const baseProjects = [
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
        terminalTabs: [],
      },
      {
        id: 'feature-auth',
        name: 'feature-auth',
        branch: 'feature/auth',
        path: '/Users/rekles/AI/agent-orch/.worktrees/feature-auth',
        hasChanges: false,
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

describe('appStore worktree flow', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    destroySessionsByWorktree.mockResolvedValue(undefined)

    useAppStore.setState({
      projects: JSON.parse(JSON.stringify(baseProjects)),
      selectedWorktreeId: 'main',
      expandedProjectIds: ['agent-orch'],
      topPanelCollapsed: false,
      diffPanelOpen: false,
      selectedDiffFile: null,
      activeTerminalTabId: 'tab-1',
    })
  })

  it('loadWorktrees should only replace agent-orch worktrees and keep other mock projects', async () => {
    ;(ListWorktrees as Mock).mockResolvedValue([
      {
        id: 'main',
        name: 'agent-orch',
        branch: 'main',
        path: '/home/zefeng/agent-orch',
        hasChanges: false,
        isMain: true,
      },
      {
        id: 'feat-runtime',
        name: 'feat-runtime',
        branch: 'feat/runtime',
        path: '/home/zefeng/feat-runtime',
        hasChanges: true,
        isMain: false,
      },
    ])

    await useAppStore.getState().loadWorktrees()

    const state = useAppStore.getState()
    const agentProject = state.projects.find((p) => p.id === 'agent-orch')
    const anotherProject = state.projects.find((p) => p.id === 'another-project')

    expect(agentProject?.worktrees.map((w) => w.id)).toEqual(['main', 'feat-runtime'])
    expect(agentProject?.path).toBe('/home/zefeng/agent-orch')
    expect(anotherProject?.worktrees.map((w) => w.id)).toEqual(['main-2'])
    expect(state.selectedWorktreeId).toBe('main')
  })

  it('createWorktree should refresh and auto-select the new worktree', async () => {
    ;(CreateWorktree as Mock).mockResolvedValue({
      id: 'feat-runtime',
      name: 'feat-runtime',
    })
    ;(ListWorktrees as Mock).mockResolvedValue([
      {
        id: 'main',
        name: 'agent-orch',
        branch: 'main',
        path: '/home/zefeng/agent-orch',
        hasChanges: false,
        isMain: true,
      },
      {
        id: 'feat-runtime',
        name: 'feat-runtime',
        branch: 'feat/runtime',
        path: '/home/zefeng/feat-runtime',
        hasChanges: false,
        isMain: false,
      },
    ])

    await useAppStore.getState().createWorktree({
      name: 'feat-runtime',
      branch: 'feat/runtime',
      baseBranch: 'main',
      createNew: true,
    })

    expect(CreateWorktree).toHaveBeenCalledWith({
      name: 'feat-runtime',
      branch: 'feat/runtime',
      baseBranch: 'main',
      createNew: true,
    })
    expect(useAppStore.getState().selectedWorktreeId).toBe('feat-runtime')
  })

  it('deleteWorktree should destroy related terminals before deleting worktree', async () => {
    ;(DeleteWorktree as Mock).mockResolvedValue(undefined)
    ;(ListWorktrees as Mock).mockResolvedValue([
      {
        id: 'main',
        name: 'agent-orch',
        branch: 'main',
        path: '/home/zefeng/agent-orch',
        hasChanges: false,
        isMain: true,
      },
    ])

    await useAppStore.getState().deleteWorktree('feature-auth', false)

    expect(destroySessionsByWorktree).toHaveBeenCalledWith('feature-auth')
    expect(DeleteWorktree).toHaveBeenCalledWith('feature-auth', false)

    const destroyOrder = destroySessionsByWorktree.mock.invocationCallOrder[0]
    const deleteOrder = (DeleteWorktree as Mock).mock.invocationCallOrder[0]
    expect(destroyOrder).toBeLessThan(deleteOrder)
  })
})
