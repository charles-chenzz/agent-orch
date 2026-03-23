import { useEffect, useMemo } from 'react'
import { Terminal } from '../Terminal'
import { useTerminalStore } from '../../stores/terminalStore'
import { useAppStore } from '../../stores/appStore'

export default function TerminalPane() {
  const { selectedWorktreeId, topPanelCollapsed } = useAppStore()

  const {
    sessions,
    activeSessionId,
    loading,
    error,
    createOrAttachSession,
    detachSession,
    setActiveSession,
    fetchSessions,
    subscribeToEvents,
    unsubscribeFromEvents,
  } = useTerminalStore()

  // Subscribe to terminal events on mount
  useEffect(() => {
    subscribeToEvents()
    fetchSessions()
    return () => unsubscribeFromEvents()
  }, [])

  // Auto-create terminal session when worktree is selected
  useEffect(() => {
    if (!selectedWorktreeId) return

    // Check if there's already a session for this worktree
    const existingSession = sessions.find(s => s.worktreeId === selectedWorktreeId)

    if (existingSession) {
      setActiveSession(existingSession.id)
    } else if (!loading) {
      // Create new session for this worktree
      const sessionId = `term-${selectedWorktreeId}-${Date.now()}`
      createOrAttachSession(sessionId, selectedWorktreeId)
    }
  }, [selectedWorktreeId, sessions.length, loading])

  // Filter sessions for current worktree
  const worktreeSessions = useMemo(() =>
    sessions.filter(s => s.worktreeId === selectedWorktreeId),
    [sessions, selectedWorktreeId]
  )

  // Active session for current worktree
  const activeSession = useMemo(() =>
    worktreeSessions.find(s => s.id === activeSessionId),
    [worktreeSessions, activeSessionId]
  )

  // Handle create new terminal tab
  const handleNewTab = () => {
    if (!selectedWorktreeId || loading) return
    const sessionId = `term-${selectedWorktreeId}-${Date.now()}`
    createOrAttachSession(sessionId, selectedWorktreeId)
  }

  // Handle close tab
  const handleCloseTab = (e: React.MouseEvent, sessionId: string) => {
    e.stopPropagation()
    detachSession(sessionId)
  }

  if (topPanelCollapsed) {
    return null
  }

  return (
    <div className="flex flex-col flex-1 min-h-0">
      {/* Tab Bar */}
      <div className="flex items-center gap-0.5 px-2 py-1.5 bg-base-surface border-b border-border">
        {worktreeSessions.length === 0 ? (
          <div className="flex items-center gap-2 text-xs text-text-muted">
            <span>{loading ? 'Starting terminal...' : 'No terminal tabs'}</span>
            {!loading && (
              <button
                onClick={handleNewTab}
                className="px-2 py-0.5 rounded bg-accent/10 text-accent hover:bg-accent/20 transition-colors"
              >
                + New Tab
              </button>
            )}
          </div>
        ) : (
          <>
            {worktreeSessions.map((session) => {
              const isActive = session.id === activeSessionId
              const isRunning = session.state === 'running'
              return (
                <div
                  key={session.id}
                  className={`group flex items-center gap-2 px-3 py-1.5 rounded-t text-sm cursor-pointer transition-all duration-150 ${
                    isActive
                      ? 'bg-[#1c2128] text-text shadow-sm'
                      : 'text-text-muted hover:text-text hover:bg-base/50'
                  }`}
                  onClick={() => setActiveSession(session.id)}
                >
                  {/* Status Indicator */}
                  <span className="relative flex h-2 w-2">
                    {isRunning ? (
                      <>
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-success opacity-75" />
                        <span className="relative inline-flex rounded-full h-2 w-2 bg-success" />
                      </>
                    ) : (
                      <span className="inline-flex rounded-full h-2 w-2 bg-text-muted/40" />
                    )}
                  </span>

                  {/* Tab Name */}
                  <span className="font-medium">shell</span>

                  {/* State Badge */}
                  {session.state !== 'running' && (
                    <span className="text-[10px] px-1.5 py-0.5 rounded font-medium bg-warning/15 text-warning">
                      {session.state}
                    </span>
                  )}

                  {/* Close Button */}
                  <button
                    className={`ml-1 w-4 h-4 flex items-center justify-center rounded transition-colors ${
                      isActive
                        ? 'text-text-muted hover:text-text hover:bg-base-surface'
                        : 'text-text-muted/0 group-hover:text-text-muted group-hover:hover:text-text'
                    }`}
                    onClick={(e) => handleCloseTab(e, session.id)}
                  >
                    <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              )
            })}

            {/* Add Tab Button */}
            <button
              onClick={handleNewTab}
              disabled={loading}
              className="w-7 h-7 flex items-center justify-center text-text-muted hover:text-text hover:bg-base/50 rounded transition-colors disabled:opacity-50"
              title="New terminal tab"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </>
        )}
      </div>

      {/* Terminal Content */}
      <div className="flex-1 bg-[#1c2128] overflow-hidden relative">
        {error && (
          <div className="absolute top-0 left-0 right-0 p-2 text-xs text-error bg-error/10 z-10">
            {error}
          </div>
        )}

        {/* Empty state */}
        {worktreeSessions.length === 0 && (
          <div className="h-full flex items-center justify-center text-text-muted text-sm">
            {loading ? 'Starting terminal...' : 'Select a worktree to start terminal'}
          </div>
        )}

        {/* Render all terminals, show only active one - preserves content when switching tabs */}
        {worktreeSessions.map((session) => {
          const isActive = session.id === activeSessionId
          const isCreating = session.state === 'creating'

          return (
            <div
              key={session.id}
              className="absolute inset-0"
              style={{ display: isActive ? 'block' : 'none' }}
            >
              {isCreating ? (
                <div className="h-full flex items-center justify-center text-text-muted text-sm">
                  Creating terminal session...
                </div>
              ) : (
                <Terminal sessionId={session.id} isActive={isActive} />
              )}
            </div>
          )
        })}
      </div>

      {/* Status Bar */}
      {activeSession && (
        <div className="px-3 py-1 text-xs text-text-muted border-t border-border bg-base-surface flex justify-between">
          <span>State: {activeSession.state}</span>
          <span>CWD: {activeSession.cwd}</span>
        </div>
      )}
    </div>
  )
}
