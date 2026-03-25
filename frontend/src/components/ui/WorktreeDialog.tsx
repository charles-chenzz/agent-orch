import { useState } from 'react'
import { Modal, Input } from './Modal'

interface CreateWorktreeDialogProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (opts: { branch: string; baseBranch: string; createNew: boolean }) => void
}

export function CreateWorktreeDialog({ isOpen, onClose, onSubmit }: CreateWorktreeDialogProps) {
  const [branch, setBranch] = useState('')
  const [baseBranch, setBaseBranch] = useState('main')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const trimmedBranch = branch.trim()
    if (!trimmedBranch) return

    onSubmit({
      branch: trimmedBranch,
      baseBranch: baseBranch.trim() || 'main',
      createNew: true,
    })
    resetAndClose()
  }

  const resetAndClose = () => {
    setBranch('')
    setBaseBranch('main')
    onClose()
  }

  return (
    <Modal isOpen={isOpen} onClose={resetAndClose} title="Create Worktree">
      <form onSubmit={handleSubmit}>
        <Input
          label="Branch name"
          value={branch}
          onChange={(e) => setBranch(e.target.value)}
          placeholder="feature/xyz"
          autoFocus
        />
        <p className="text-xs text-text-muted -mt-2 mb-3">
          Directory will be auto-generated from branch name
        </p>
        <Input
          label="Base branch"
          value={baseBranch}
          onChange={(e) => setBaseBranch(e.target.value)}
          placeholder="main"
        />
        <div className="flex justify-end gap-2 mt-4">
          <button
            type="button"
            onClick={resetAndClose}
            className="px-3 py-1.5 text-xs rounded text-text-muted hover:text-text hover:bg-base-surface transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={!branch.trim()}
            className="px-3 py-1.5 text-xs rounded bg-accent text-white hover:bg-accent/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            Create
          </button>
        </div>
      </form>
    </Modal>
  )
}

interface DeleteWorktreeDialogProps {
  isOpen: boolean
  onClose: () => void
  onConfirm: (force: boolean) => void
  worktreeName: string
  hasChanges?: boolean
}

export function DeleteWorktreeDialog({
  isOpen,
  onClose,
  onConfirm,
  worktreeName,
  hasChanges,
}: DeleteWorktreeDialogProps) {
  const [showForceOption, setShowForceOption] = useState(false)

  const handleConfirm = () => {
    onConfirm(showForceOption)
    onClose()
  }

  if (!isOpen) return null

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Delete Worktree">
      <p className="text-sm text-text mb-4">
        Delete worktree <strong className="text-accent">{worktreeName}</strong>?
      </p>
      <p className="text-xs text-text-muted mb-4">
        Related terminal sessions will be destroyed first.
      </p>

      {hasChanges && (
        <label className="flex items-center gap-2 mb-4 text-xs text-text-muted">
          <input
            type="checkbox"
            checked={showForceOption}
            onChange={(e) => setShowForceOption(e.target.checked)}
            className="rounded border-border"
          />
          Force delete (has uncommitted changes)
        </label>
      )}

      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onClose}
          className="px-3 py-1.5 text-xs rounded text-text-muted hover:text-text hover:bg-base-surface transition-colors"
        >
          Cancel
        </button>
        <button
          type="button"
          onClick={handleConfirm}
          className="px-3 py-1.5 text-xs rounded bg-error text-white hover:bg-error/90 transition-colors"
        >
          Delete
        </button>
      </div>
    </Modal>
  )
}
