import { beforeEach, describe, expect, it, vi, type Mock } from 'vitest'

vi.mock('../../../wailsjs/go/main/App', () => ({
  CreateOrAttachTerminal: vi.fn(),
  DetachTerminal: vi.fn(),
  DestroyTerminal: vi.fn(),
  ListTerminalSessions: vi.fn(),
}))

vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
}))

import { DestroyTerminal, ListTerminalSessions } from '../../../wailsjs/go/main/App'
import { useTerminalStore } from '../terminalStore'

describe('terminalStore destroySessionsByWorktree', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    useTerminalStore.setState({
      sessions: [
        {
          id: 's1',
          worktreeId: 'wt-1',
          cwd: '/tmp/wt-1',
          state: 'running',
          createdAt: '',
          lastActive: '',
        },
        {
          id: 's2',
          worktreeId: 'wt-1',
          cwd: '/tmp/wt-1',
          state: 'detached',
          createdAt: '',
          lastActive: '',
        },
        {
          id: 's3',
          worktreeId: 'wt-2',
          cwd: '/tmp/wt-2',
          state: 'running',
          createdAt: '',
          lastActive: '',
        },
      ],
      activeSessionId: 's1',
      loading: false,
      error: null,
    })
  })

  it('should destroy only sessions belonging to the target worktree', async () => {
    ;(DestroyTerminal as Mock).mockResolvedValue(undefined)
    ;(ListTerminalSessions as Mock).mockResolvedValue([
      {
        id: 's3',
        worktreeId: 'wt-2',
        cwd: '/tmp/wt-2',
        state: 'running',
        createdAt: '',
        lastActive: '',
      },
    ])

    await useTerminalStore.getState().destroySessionsByWorktree('wt-1')

    expect(DestroyTerminal).toHaveBeenCalledTimes(2)
    expect(DestroyTerminal).toHaveBeenNthCalledWith(1, 's1')
    expect(DestroyTerminal).toHaveBeenNthCalledWith(2, 's2')
    expect(ListTerminalSessions).toHaveBeenCalledTimes(1)

    const state = useTerminalStore.getState()
    expect(state.sessions.map((s) => s.id)).toEqual(['s3'])
    expect(state.activeSessionId).toBe('s3')
  })

  it('should no-op when no session belongs to the target worktree', async () => {
    await useTerminalStore.getState().destroySessionsByWorktree('wt-3')

    expect(DestroyTerminal).not.toHaveBeenCalled()
    expect(ListTerminalSessions).not.toHaveBeenCalled()
  })

  it('should surface errors from DestroyTerminal', async () => {
    ;(DestroyTerminal as Mock).mockRejectedValue(new Error('destroy failed'))

    await expect(
      useTerminalStore.getState().destroySessionsByWorktree('wt-1')
    ).rejects.toThrow('destroy failed')

    expect(useTerminalStore.getState().error).toContain('destroy failed')
  })
})
