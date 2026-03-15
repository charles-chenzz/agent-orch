package worktree

import "fmt"

// WorktreeError 结构化错误类型
type WorktreeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *WorktreeError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 错误码定义
const (
	ErrNameRequired     = "ERR_NAME_REQUIRED"
	ErrNameInvalid      = "ERR_NAME_INVALID"
	ErrNameConflict     = "ERR_NAME_CONFLICT"
	ErrBranchRequired   = "ERR_BRANCH_REQUIRED"
	ErrBranchNotFound   = "ERR_BRANCH_NOT_FOUND"
	ErrBaseNotFound     = "ERR_BASE_NOT_FOUND"
	ErrPathExists       = "ERR_PATH_EXISTS"
	ErrGitFailed        = "ERR_GIT_FAILED"
	ErrCreateFailed     = "ERR_CREATE_FAILED"
	ErrCannotDeleteMain = "ERR_CANNOT_DELETE_MAIN"
	ErrNotFound         = "ERR_NOT_FOUND"
	ErrHasChanges       = "ERR_HAS_CHANGES"
	ErrHasUnpushed      = "ERR_HAS_UNPUSHED"
	ErrDeleteFailed     = "ERR_DELETE_FAILED"
)

// NewWorktreeError 创建新的 WorktreeError
func NewWorktreeError(code, message string) *WorktreeError {
	return &WorktreeError{Code: code, Message: message}
}
