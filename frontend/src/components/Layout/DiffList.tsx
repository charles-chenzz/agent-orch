import { useAppStore } from '../../stores/appStore'

export default function DiffList() {
  const { getDiffFiles, selectDiffFile, selectedDiffFile } = useAppStore()
  const files = getDiffFiles()

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-3 py-2 border-b border-border">
        <span className="text-xs font-semibold text-text-muted uppercase tracking-wide">Diff</span>
        <div className="flex gap-2">
          <button className="text-xs text-text-muted hover:text-text">Accept All</button>
          <button className="text-xs text-text-muted hover:text-text">Reject All</button>
        </div>
      </div>
      <div className="flex-1 overflow-auto">
        {files.map((file) => {
          const isSelected = file.path === selectedDiffFile
          return (
            <div
              key={file.path}
              className={`flex items-center gap-2 px-3 py-1.5 cursor-pointer ${
                isSelected ? 'bg-accent/20' : 'hover:bg-base-surface'
              }`}
              onClick={() => selectDiffFile(file.path)}
            >
              <span className="text-text-muted">▸</span>
              <span className="text-sm text-text truncate flex-1">{file.path}</span>
              <span className="text-xs">
                <span className="text-success">+{file.additions}</span>
                <span className="text-error ml-1">-{file.deletions}</span>
              </span>
            </div>
          )
        })}
        {files.length === 0 && (
          <div className="text-center text-text-muted py-8 text-sm">No changes</div>
        )}
      </div>
    </div>
  )
}
