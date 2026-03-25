import { useState } from 'react'
import { useAppStore } from '../../stores/appStore'
import { CreateWorktreeDialog, DeleteWorktreeDialog } from '../ui/WorktreeDialog'

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
  const {
    projects,
    selectedWorktreeId,
    expandedProjectIds,
    selectWorktree,
    toggleProjectExpand,
    createWorktree,
    deleteWorktree,
  } = useAppStore()

  // Dialog state
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [pendingDelete, setPendingDelete] = useState<{ id: string; name: string } | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleCreateWorktree = async (opts: { branch: string; baseBranch: string; createNew: boolean }) => {
    try {
      await createWorktree({
        name: '',  // 后端会自动从 branch 生成
        branch: opts.branch,
        baseBranch: opts.baseBranch,
        createNew: opts.createNew,
      })
    } catch (err) {
      setError(`Failed to create worktree: ${String(err)}`)
    }
  }

  const handleDeleteClick = (
    e: React.MouseEvent<HTMLButtonElement>,
    worktreeId: string,
    worktreeName: string
  ) => {
    e.stopPropagation()
    setPendingDelete({ id: worktreeId, name: worktreeName })
    setDeleteDialogOpen(true)
  }

  const handleDeleteConfirm = async (force: boolean) => {
    if (!pendingDelete) return
    try {
      await deleteWorktree(pendingDelete.id, force)
    } catch (err) {
      const msg = String(err)
      if (!force && (msg.includes('ERR_HAS_CHANGES') || msg.includes('ERR_HAS_UNPUSHED'))) {
        // Reopen with force option visible
        setDeleteDialogOpen(true)
        return
      }
      setError(`Failed to delete worktree: ${msg}`)
    }
    setPendingDelete(null)
  }

  return (
    <div className="py-1">
      {/* Error toast */}
      {error && (
        <div className="fixed top-4 right-4 z-50 px-4 py-2 bg-error/90 text-white text-xs rounded shadow-lg animate-fade-in">
          {error}
          <button onClick={() => setError(null)} className="ml-2 opacity-70 hover:opacity-100">×</button>
        </div>
      )}

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
                  const canDelete = project.id === 'agent-orch' && wt.id !== 'main' && wt.name !== 'main'
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
                      <div className="ml-auto flex items-center gap-1">
                        {wt.terminalTabs.length > 0 && (
                          <span className="text-[10px] text-text-muted">
                            {wt.terminalTabs.length} tab{wt.terminalTabs.length > 1 ? 's' : ''}
                          </span>
                        )}
                        {canDelete && (
                          <button
                            className="w-4 h-4 flex items-center justify-center rounded text-text-muted hover:text-error hover:bg-error/10 transition-colors"
                            title={`Delete worktree ${wt.name}`}
                            onClick={(e) => handleDeleteClick(e, wt.id, wt.name)}
                          >
                            ×
                          </button>
                        )}
                      </div>
                    </div>
                  )
                })}
                <button
                  className={`w-full mt-1 px-2 py-1 text-left text-xs rounded transition-colors ${
                    project.id === 'agent-orch'
                      ? 'text-accent hover:bg-accent/10'
                      : 'text-text-muted/60 cursor-not-allowed'
                  }`}
                  onClick={() => project.id === 'agent-orch' && setCreateDialogOpen(true)}
                  title={project.id === 'agent-orch' ? 'Create worktree' : 'Mock project only'}
                >
                  + New Worktree
                </button>
              </div>
            </div>
          </div>
        )
      })}

      {/* Dialogs */}
      <CreateWorktreeDialog
        isOpen={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        onSubmit={handleCreateWorktree}
      />
      <DeleteWorktreeDialog
        isOpen={deleteDialogOpen}
        onClose={() => {
          setDeleteDialogOpen(false)
          setPendingDelete(null)
        }}
        onConfirm={handleDeleteConfirm}
        worktreeName={pendingDelete?.name || ''}
      />
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
