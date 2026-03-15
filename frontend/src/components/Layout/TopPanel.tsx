import { useAppStore } from '../../stores/appStore'

export default function TopPanel() {
  const { topPanelCollapsed, toggleTopPanel } = useAppStore()

  return (
    <div
      className={`flex border-b border-border transition-all duration-200 ease-out overflow-hidden ${
        topPanelCollapsed ? 'h-10' : 'h-48'
      }`}
    >
      {/* Projects Section */}
      <div className="flex-[2] min-w-[180px] max-w-[280px] border-r border-border overflow-hidden">
        <div
          className={`h-full transition-opacity duration-150 ${
            topPanelCollapsed ? 'opacity-0' : 'opacity-100'
          }`}
        >
          {!topPanelCollapsed && (
            <div className="h-full flex flex-col animate-fade-in">
              <div className="px-3 py-2 text-xs font-semibold text-text-muted uppercase tracking-wider border-b border-border bg-base-surface/50">
                Projects
              </div>
              <div className="flex-1 overflow-auto">
                <ProjectListCompact />
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Diff Section */}
      <div className="flex-1 min-w-[200px] overflow-hidden">
        <div
          className={`h-full transition-opacity duration-150 ${
            topPanelCollapsed ? 'opacity-0' : 'opacity-100'
          }`}
        >
          {!topPanelCollapsed && (
            <div className="h-full flex flex-col animate-fade-in">
              <div className="px-3 py-2 text-xs font-semibold text-text-muted uppercase tracking-wider border-b border-border bg-base-surface/50 flex items-center justify-between">
                <span>Changes</span>
                <DiffCountBadge />
              </div>
              <div className="flex-1 overflow-auto">
                <DiffListCompact />
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Collapse Button */}
      <button
        className={`w-8 flex-shrink-0 flex items-center justify-center text-text-muted hover:text-text hover:bg-base-surface transition-all duration-150 active-scale ${
          topPanelCollapsed ? 'bg-base-surface/30' : ''
        }`}
        onClick={toggleTopPanel}
        title={topPanelCollapsed ? 'Expand panel (⌘B)' : 'Collapse panel (⌘B)'}
      >
        <svg
          className={`w-4 h-4 transition-transform duration-200 ${
            topPanelCollapsed ? 'rotate-180' : ''
          }`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M5 15l7-7 7 7" />
        </svg>
      </button>
    </div>
  )
}

function DiffCountBadge() {
  const { getDiffFiles } = useAppStore()
  const files = getDiffFiles()
  const totalChanges = files.reduce((acc, f) => acc + f.additions + f.deletions, 0)

  return (
    <span className="px-1.5 py-0.5 text-[10px] font-medium bg-accent/10 text-accent rounded">
      {files.length} files · +{files.reduce((a, f) => a + f.additions, 0)} -{files.reduce((a, f) => a + f.deletions, 0)}
    </span>
  )
}

function ProjectListCompact() {
  const { projects, selectedWorktreeId, expandedProjectIds, selectWorktree, toggleProjectExpand } = useAppStore()

  return (
    <div className="py-1">
      {projects.map((project) => {
        const isExpanded = expandedProjectIds.includes(project.id)
        return (
          <div key={project.id}>
            {/* Project Header */}
            <div
              className="flex items-center gap-1.5 px-2 py-1.5 cursor-pointer text-text-muted hover:text-text hover:bg-base-surface/50 transition-colors duration-100 group"
              onClick={() => toggleProjectExpand(project.id)}
            >
              <svg
                className={`w-3 h-3 transition-transform duration-150 ${isExpanded ? 'rotate-90' : ''}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
              </svg>
              <span className="text-xs font-medium">{project.name}</span>
              <span className="ml-auto text-[10px] text-text-muted opacity-0 group-hover:opacity-100 transition-opacity">
                {project.worktrees.length}
              </span>
            </div>

            {/* Worktrees */}
            <div
              className={`overflow-hidden transition-all duration-200 ${
                isExpanded ? 'max-h-40 opacity-100' : 'max-h-0 opacity-0'
              }`}
            >
              <div className="ml-4 border-l border-border pl-1">
                {project.worktrees.map((wt) => {
                  const isSelected = wt.id === selectedWorktreeId
                  return (
                    <div
                      key={wt.id}
                      className={`flex items-center gap-2 px-2 py-1 cursor-pointer text-xs rounded transition-all duration-100 ${
                        isSelected
                          ? 'bg-accent/15 text-accent border-l-2 border-accent -ml-[2px] pl-[calc(0.5rem+2px)]'
                          : 'hover:bg-base-surface/50 text-text-muted hover:text-text'
                      }`}
                      onClick={() => selectWorktree(wt.id)}
                    >
                      <span className="relative flex h-2 w-2">
                        {wt.hasChanges ? (
                          <>
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-warning opacity-75" />
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-warning" />
                          </>
                        ) : (
                          <span className="inline-flex rounded-full h-2 w-2 bg-success/60" />
                        )}
                      </span>
                      <span className="truncate">{wt.name}</span>
                      {wt.terminalTabs.length > 0 && (
                        <span className="ml-auto text-[10px] text-text-muted">
                          {wt.terminalTabs.length} tab{wt.terminalTabs.length > 1 ? 's' : ''}
                        </span>
                      )}
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}

function DiffListCompact() {
  const { getDiffFiles, selectDiffFile, selectedDiffFile } = useAppStore()
  const files = getDiffFiles()

  return (
    <div className="py-1">
      {files.map((file, index) => {
        const isSelected = file.path === selectedDiffFile
        return (
          <div
            key={file.path}
            className={`flex items-center gap-2 px-3 py-1.5 cursor-pointer transition-all duration-100 group ${
              isSelected
                ? 'bg-accent/10 border-l-2 border-accent'
                : 'hover:bg-base-surface/50 border-l-2 border-transparent hover:border-border'
            }`}
            onClick={() => selectDiffFile(file.path)}
            style={{ animationDelay: `${index * 30}ms` }}
          >
            {/* File Icon */}
            <span className={`text-xs ${isSelected ? 'text-accent' : 'text-text-muted'}`}>
              {getFileIcon(file.path)}
            </span>

            {/* File Path */}
            <span className={`text-xs truncate flex-1 ${isSelected ? 'text-text font-medium' : 'text-text-muted group-hover:text-text'}`}>
              {file.path}
            </span>

            {/* Changes */}
            <div className="flex items-center gap-1.5 text-[10px] font-mono">
              <span className="text-success">+{file.additions}</span>
              <span className="text-error">-{file.deletions}</span>
            </div>

            {/* Click hint */}
            <svg
              className="w-3 h-3 text-text-muted opacity-0 group-hover:opacity-100 transition-opacity"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        )
      })}
    </div>
  )
}

function getFileIcon(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase()
  const icons: Record<string, string> = {
    ts: 'TS',
    tsx: '⚛',
    js: 'JS',
    jsx: '⚛',
    go: 'Go',
    rs: 'RS',
    py: 'Py',
    css: '§',
    scss: '§',
    json: '{}',
    md: '¶',
    html: '🌐',
  }
  return icons[ext || ''] || '📄'
}
