import { useAppStore } from '../../stores/appStore'

export default function ProjectList() {
  const { projects, selectedWorktreeId, expandedProjectIds, selectWorktree, toggleProjectExpand } = useAppStore()

  return (
    <div className="flex flex-col h-full">
      <div className="px-3 py-2 text-xs font-semibold text-text-muted uppercase tracking-wide">
        Projects
      </div>
      <div className="flex-1 overflow-auto">
        {projects.map((project) => {
          const isExpanded = expandedProjectIds.includes(project.id)
          return (
            <div key={project.id}>
              <div
                className="flex items-center gap-1 px-2 py-1.5 cursor-pointer hover:bg-base-surface"
                onClick={() => toggleProjectExpand(project.id)}
              >
                <span className={`text-text-muted transition-transform ${isExpanded ? 'rotate-90' : ''}`}>
                  ▶
                </span>
                <span className="text-sm font-medium text-text">{project.name}</span>
                <span className="text-xs text-text-muted ml-auto">
                  {project.worktrees.length}
                </span>
              </div>
              {isExpanded && (
                <div className="ml-4">
                  {project.worktrees.map((wt) => {
                    const isSelected = wt.id === selectedWorktreeId
                    return (
                      <div
                        key={wt.id}
                        className={`flex items-center gap-2 px-2 py-1 cursor-pointer ${
                          isSelected ? 'bg-accent/20 text-accent' : 'hover:bg-base-surface text-text-muted'
                        }`}
                        onClick={() => selectWorktree(wt.id)}
                      >
                        <span className={`w-2 h-2 rounded-full ${wt.hasChanges ? 'bg-warning' : 'bg-success'}`} />
                        <span className="text-sm">{wt.name}</span>
                        <span className="text-xs text-text-muted ml-auto truncate">
                          {wt.branch}
                        </span>
                      </div>
                    )
                  })}
                  <div className="px-2 py-1 text-xs text-text-muted hover:text-accent cursor-pointer">
                    + New Worktree
                  </div>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
