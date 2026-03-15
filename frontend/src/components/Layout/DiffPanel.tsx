import { useAppStore } from '../../stores/appStore'

export default function DiffPanel() {
  const { diffPanelOpen, selectedDiffFile, selectDiffFile, getDiffFiles } = useAppStore()
  const files = getDiffFiles()
  const file = files.find(f => f.path === selectedDiffFile)

  if (!diffPanelOpen || !selectedDiffFile || !file) return null

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 animate-fade-in"
        onClick={() => selectDiffFile(null)}
      />

      {/* Panel */}
      <div className="fixed right-0 top-0 bottom-0 w-[520px] bg-base-surface border-l border-border z-50 shadow-2xl animate-slide-in-right flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-border bg-base/50">
          <div className="flex items-center gap-3">
            <button
              className="w-7 h-7 flex items-center justify-center text-text-muted hover:text-text hover:bg-base-surface rounded transition-colors"
              onClick={() => selectDiffFile(null)}
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <div>
              <div className="text-sm font-medium text-text">{file.path}</div>
              <div className="text-xs text-text-muted mt-0.5">
                <span className="text-success">+{file.additions}</span>
                <span className="mx-1">·</span>
                <span className="text-error">-{file.deletions}</span>
                <span className="mx-1">·</span>
                <span className="capitalize">{file.status}</span>
              </div>
            </div>
          </div>
          <button
            className="w-7 h-7 flex items-center justify-center text-text-muted hover:text-text hover:bg-base-surface rounded transition-colors text-lg"
            onClick={() => selectDiffFile(null)}
          >
            ×
          </button>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 px-4 py-2 border-b border-border bg-base/30">
          <button className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-success/15 text-success rounded-md hover:bg-success/25 transition-colors active-scale">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
            </svg>
            Accept
          </button>
          <button className="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-error/15 text-error rounded-md hover:bg-error/25 transition-colors active-scale">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
            Reject
          </button>
          <div className="flex-1" />
          <button className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-text-muted hover:text-text hover:bg-base-surface rounded-md transition-colors">
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
            </svg>
            Full Screen
          </button>
        </div>

        {/* Diff Content */}
        <div className="flex-1 overflow-auto">
          <div className="p-4">
            {/* Unified Diff View */}
            <div className="font-mono text-[13px] leading-6">
              {/* Diff Header */}
              <div className="flex items-center gap-4 mb-4 text-xs text-text-muted">
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-sm bg-error/30" />
                  <span>Removed</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="w-3 h-3 rounded-sm bg-success/30" />
                  <span>Added</span>
                </div>
              </div>

              {/* Diff Lines */}
              <div className="rounded-lg border border-border overflow-hidden">
                {/* Context Line */}
                <div className="flex bg-base/50 text-text-muted">
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px]">1</span>
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px]">1</span>
                  <span className="px-3 py-0.5">function App() {'{'}</span>
                </div>

                {/* Removed Line */}
                <div className="flex bg-error/10 border-l-2 border-error">
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px] text-error/60">2</span>
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px] text-text-muted/30">-</span>
                  <span className="px-3 py-0.5 text-error">
                    <span className="opacity-60">-</span>  return &lt;div&gt;Hello&lt;/div&gt;
                  </span>
                </div>

                {/* Added Lines */}
                <div className="flex bg-success/10 border-l-2 border-success">
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px] text-success/60">-</span>
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px]">2</span>
                  <span className="px-3 py-0.5 text-success">
                    <span className="opacity-60">+</span>  const [state, setState] = useState()
                  </span>
                </div>
                <div className="flex bg-success/10 border-l-2 border-success">
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px] text-success/60">-</span>
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px]">3</span>
                  <span className="px-3 py-0.5 text-success">
                    <span className="opacity-60">+</span>  return &lt;div className="app"&gt;{`{state}`}...&lt;/div&gt;
                  </span>
                </div>
                <div className="flex bg-success/10 border-l-2 border-success">
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px] text-success/60">-</span>
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px]">4</span>
                  <span className="px-3 py-0.5 text-success">
                    <span className="opacity-60">+</span>  // Added new line
                  </span>
                </div>

                {/* Context Line */}
                <div className="flex bg-base/50 text-text-muted">
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px]">3</span>
                  <span className="w-10 text-right pr-3 py-0.5 select-none border-r border-border text-[11px]">5</span>
                  <span className="px-3 py-0.5">{'}'}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-4 py-2 border-t border-border bg-base/30 text-xs text-text-muted">
          <span>Press <kbd className="px-1.5 py-0.5 bg-base-surface rounded text-[10px]">Esc</kbd> to close</span>
          <span>Use <kbd className="px-1.5 py-0.5 bg-base-surface rounded text-[10px]">j</kbd> / <kbd className="px-1.5 py-0.5 bg-base-surface rounded text-[10px]">k</kbd> to navigate</span>
        </div>
      </div>
    </>
  )
}
