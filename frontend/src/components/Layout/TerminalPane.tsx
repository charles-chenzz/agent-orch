import { useAppStore } from '../../stores/appStore'

export default function TerminalPane() {
  const { getCurrentWorktree, activeTerminalTabId, selectTerminalTab, topPanelCollapsed } = useAppStore()
  const worktree = getCurrentWorktree()
  const tabs = worktree?.terminalTabs || []

  return (
    <div className={`flex flex-col flex-1 min-h-0 transition-all duration-200 ${topPanelCollapsed ? '' : ''}`}>
      {/* Tab Bar */}
      <div className="flex items-center gap-0.5 px-2 py-1.5 bg-base-surface border-b border-border">
        {tabs.length === 0 ? (
          <div className="flex items-center gap-2 text-xs text-text-muted">
            <span>No terminal tabs</span>
            <button className="px-2 py-0.5 rounded bg-accent/10 text-accent hover:bg-accent/20 transition-colors">
              + New Tab
            </button>
          </div>
        ) : (
          <>
            {tabs.map((tab) => {
              const isActive = tab.id === activeTerminalTabId
              return (
                <div
                  key={tab.id}
                  className={`group flex items-center gap-2 px-3 py-1.5 rounded-t text-sm cursor-pointer transition-all duration-150 ${
                    isActive
                      ? 'bg-base text-text shadow-sm'
                      : 'text-text-muted hover:text-text hover:bg-base/50'
                  }`}
                  onClick={() => selectTerminalTab(tab.id)}
                >
                  {/* Status Indicator */}
                  <span className="relative flex h-2 w-2">
                    {tab.status === 'running' ? (
                      <>
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-success opacity-75" />
                        <span className="relative inline-flex rounded-full h-2 w-2 bg-success" />
                      </>
                    ) : (
                      <span className="inline-flex rounded-full h-2 w-2 bg-text-muted/40" />
                    )}
                  </span>

                  {/* Tab Name */}
                  <span className="font-medium">{tab.name}</span>

                  {/* Agent Badge */}
                  {tab.agentType && tab.agentType !== 'shell' && (
                    <span className={`text-[10px] px-1.5 py-0.5 rounded font-medium ${
                      isActive
                        ? 'bg-accent/15 text-accent'
                        : 'bg-base-surface text-text-muted group-hover:bg-base-surface'
                    }`}>
                      {tab.agentType}
                    </span>
                  )}

                  {/* Close Button */}
                  <button
                    className={`ml-1 w-4 h-4 flex items-center justify-center rounded transition-colors ${
                      isActive
                        ? 'text-text-muted hover:text-text hover:bg-base-surface'
                        : 'text-text-muted/0 group-hover:text-text-muted group-hover:hover:text-text'
                    }`}
                    onClick={(e) => {
                      e.stopPropagation()
                      // TODO: close tab
                    }}
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
              className="w-7 h-7 flex items-center justify-center text-text-muted hover:text-text hover:bg-base/50 rounded transition-colors"
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
      <div className="flex-1 bg-[#1c2128] overflow-hidden">
        <div className="h-full p-4 font-mono text-[13px] leading-relaxed overflow-auto">
          {/* Mock Terminal Output */}
          <div className="space-y-1">
            <div className="text-text-muted">
              <span className="text-success">$</span> claude
            </div>
            <div className="text-text mt-2">
              <span className="text-accent">✓</span> Connected to claude-opus-4-6
            </div>
            <div className="text-text mt-3">
              <span className="text-warning">→</span> Working on authentication implementation...
            </div>

            {/* Progress Bar */}
            <div className="mt-4 flex items-center gap-3">
              <div className="flex-1 h-1.5 bg-base-surface rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-accent to-accent-muted rounded-full transition-all duration-300"
                  style={{ width: '60%' }}
                />
              </div>
              <span className="text-xs text-text-muted font-medium">60%</span>
            </div>

            {/* Current Task */}
            <div className="mt-3 p-2 bg-base-surface/50 rounded border border-border">
              <div className="text-xs text-text-muted mb-1">Current task:</div>
              <div className="text-sm text-text">Refactoring auth middleware for better error handling</div>
            </div>

            {/* Cursor */}
            <div className="mt-2 flex items-center">
              <span className="text-success">→</span>
              <span className="ml-2 w-2 h-4 bg-accent animate-pulse" />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
