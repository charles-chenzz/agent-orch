import { create } from 'zustand'
import type { SessionInfo, SessionState, TerminalEvent } from '../types/terminal'
import { TERMINAL_EVENTS } from '../types/terminal'

// Wails bindings
import {
  CreateOrAttachTerminal,
  DetachTerminal,
  DestroyTerminal,
  ListTerminalSessions,
} from '../../wailsjs/go/main/App'

// Wails runtime for events
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
  destroySessionsByWorktree: (worktreeId: string) => Promise<void>
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
      const rawSessions = await ListTerminalSessions()
      // Convert wails SessionInfo class instances to plain objects
      const sessions: SessionInfo[] = (rawSessions || []).map(s => ({
        id: s.id,
        worktreeId: s.worktreeId,
        cwd: s.cwd,
        state: s.state as SessionState,
        createdAt: String(s.createdAt),
        lastActive: String(s.lastActive),
      }))
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
      // If detaching the active session, clear selection
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
      // If destroying the active session, clear selection
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

  destroySessionsByWorktree: async (worktreeId: string) => {
    const sessionIds = get().sessions
      .filter((s) => s.worktreeId === worktreeId)
      .map((s) => s.id)

    if (sessionIds.length === 0) {
      return
    }

    try {
      for (const sessionId of sessionIds) {
        await DestroyTerminal(sessionId)
      }
      await get().fetchSessions()

      if (sessionIds.includes(get().activeSessionId || '')) {
        const remaining = get().sessions.filter((s) => !sessionIds.includes(s.id))
        set({
          activeSessionId: remaining.length > 0 ? remaining[0].id : null,
        })
      }
    } catch (err) {
      set({ error: String(err) })
      throw err
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
    // Subscribe to unified event protocol
    EventsOn(TERMINAL_EVENTS.STATE, (event: TerminalEvent) => {
      if (event.state) {
        get().updateSessionState(event.sessionId, event.state)
      }
    })

    EventsOn(TERMINAL_EVENTS.ERROR, (event: TerminalEvent) => {
      if (event.error) {
        set({ error: event.error })
      }
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
