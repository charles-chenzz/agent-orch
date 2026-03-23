// Terminal types for Phase 2 implementation

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

// Unified event protocol
export interface TerminalEvent {
  sessionId: string
  type: 'output' | 'state' | 'error' | 'exit'
  data?: string
  state?: SessionState
  error?: string
  ts: number // timestamp in ms
}

// Event type constants
export const TERMINAL_EVENTS = {
  OUTPUT: 'terminal:output',
  STATE: 'terminal:state',
  ERROR: 'terminal:error',
  EXIT: 'terminal:exit',
} as const
