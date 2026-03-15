import { useAppStore } from '../../stores/appStore'

export default function StatusBar() {
  const { getCurrentWorktree, projects, selectedWorktreeId, getDiffFiles } = useAppStore()
  const worktree = getCurrentWorktree()
  const project = projects.find(p => p.worktrees.some(w => w.id === selectedWorktreeId))
  const diffFiles = getDiffFiles()
  const hasChanges = worktree?.hasChanges || diffFiles.length > 0

  return (
    <div className="flex items-center justify-between px-4 py-1.5 border-t border-border bg-base-surface/80 text-xs">
      {/* Left: Context */}
      <div className="flex items-center gap-4">
        {/* Worktree + Project */}
        <div className="flex items-center gap-1.5">
          <span className="font-medium text-text">{worktree?.name || 'No worktree'}</span>
          <span className="text-text-muted">@</span>
          <span className="text-text-muted">{project?.name || 'Unknown'}</span>
        </div>

        {/* Branch */}
        {worktree?.branch && (
          <div className="flex items-center gap-1.5 text-text-muted">
            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            <span>{worktree.branch}</span>
          </div>
        )}

        {/* Changes Indicator */}
        {hasChanges && (
          <div className="flex items-center gap-1.5 px-2 py-0.5 bg-warning/10 text-warning rounded-full">
            <span className="w-1.5 h-1.5 rounded-full bg-warning animate-pulse-soft" />
            <span>{diffFiles.length} change{diffFiles.length > 1 ? 's' : ''}</span>
          </div>
        )}
      </div>

      {/* Right: Status Info */}
      <div className="flex items-center gap-4">
        {/* Agent Status */}
        <div className="flex items-center gap-1.5">
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-success opacity-75" />
            <span className="relative inline-flex rounded-full h-2 w-2 bg-success" />
          </span>
          <span className="text-text-muted">Agent:</span>
          <span className="text-text font-medium">claude-opus</span>
        </div>

        {/* API Usage */}
        <div className="flex items-center gap-1.5 text-text-muted">
          <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>$2.34 today</span>
        </div>

        {/* Time */}
        <div className="text-text-muted/60">
          {new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
        </div>
      </div>
    </div>
  )
}
