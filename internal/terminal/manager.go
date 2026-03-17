package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// NewManager 创建终端管理器
func NewManager(ctx context.Context) *Manager {
	tmuxPath, err := exec.LookPath("tmux")
	hasTmux := err == nil

	return &Manager{
		sessions: make(map[string]*Session),
		ctx:      ctx,
		tmuxPath: tmuxPath,
		hasTmux:  hasTmux,
	}
}

// CreateOrAttachSession 创建或附加到现有会话 (F2.2)
func (m *Manager) CreateOrAttachSession(id, worktreeID, cwd string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在 detached 的会话
	if session, exists := m.sessions[id]; exists {
		if session.State == StateDetached {
			// 重新附加
			return m.attachSession(session)
		}
		return fmt.Errorf("session already running: %s", id)
	}

	// 创建新会话
	session := &Session{
		ID:         id,
		WorktreeID: worktreeID,
		CWD:        cwd,
		State:      StateCreating,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
	m.sessions[id] = session

	var ptyFile *os.File
	var cmd *exec.Cmd
	var tmuxSession string

	if m.hasTmux {
		// 使用 tmux（-A 如果不存在则创建）
		tmuxSession = fmt.Sprintf("agent-orch-%s", id)
		cmd = exec.Command(m.tmuxPath, "new-session",
			"-A",              // 如果不存在则创建
			"-s", tmuxSession, // session 名称
			"-c", cwd, // 工作目录
		)
	} else {
		// 直接使用 shell
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
		cmd = exec.Command(shell)
		cmd.Dir = cwd
	}

	// 启动 PTY
	var err error
	ptyFile, err = pty.Start(cmd)
	if err != nil {
		session.State = StateDestroyed
		m.emitState(session)
		delete(m.sessions, id)
		return fmt.Errorf("failed to start pty: %w", err)
	}

	session.PTY = ptyFile
	session.Cmd = cmd
	session.TmuxSession = tmuxSession
	session.State = StateRunning
	m.emitState(session)

	// 启动输出读取协程
	go m.readOutput(session)

	return nil
}

// attachSession 重新附加到 detached 的会话
func (m *Manager) attachSession(session *Session) error {
	if session.TmuxSession == "" || !m.hasTmux {
		return fmt.Errorf("cannot reattach to non-tmux session")
	}

	// tmux session 仍在运行，只需重新附加
	session.State = StateRunning
	session.LastActive = time.Now()
	m.emitState(session)

	// 重新启动输出读取
	go m.readOutput(session)

	return nil
}

// readOutput 读取 PTY 输出并发送到前端（统一事件协议）
func (m *Manager) readOutput(session *Session) {
	buf := make([]byte, 4096)

	for {
		n, err := session.PTY.Read(buf)
		if err != nil {
			// PTY 关闭或出错
			m.handleSessionExit(session.ID)
			return
		}

		if n > 0 {
			// 统一事件协议
			m.emitOutput(session.ID, string(buf[:n]))
		}
	}
}

// emitOutput 发送输出事件（统一协议）
func (m *Manager) emitOutput(sessionID, data string) {
	if m.ctx == nil {
		return
	}
	runtime.EventsEmit(m.ctx, "terminal:output", EventPayload{
		SessionID: sessionID,
		Type:      "output",
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// emitState 发送状态变更事件
func (m *Manager) emitState(session *Session) {
	if m.ctx == nil {
		return
	}
	runtime.EventsEmit(m.ctx, "terminal:state", EventPayload{
		SessionID: session.ID,
		Type:      "state",
		State:     string(session.State),
		Timestamp: time.Now().UnixMilli(),
	})
}

// emitError 发送错误事件
func (m *Manager) emitError(sessionID, errMsg string) {
	if m.ctx == nil {
		return
	}
	runtime.EventsEmit(m.ctx, "terminal:error", EventPayload{
		SessionID: sessionID,
		Type:      "error",
		Error:     errMsg,
		Timestamp: time.Now().UnixMilli(),
	})
}

// SendInput 发送输入到 PTY
func (m *Manager) SendInput(id, data string) error {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}

	if session.State != StateRunning {
		return fmt.Errorf("session not running: %s (state: %s)", id, session.State)
	}

	_, err := session.PTY.Write([]byte(data))
	return err
}

// Resize 调整 PTY 尺寸
func (m *Manager) Resize(id string, cols, rows uint16) error {
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}

	return pty.Setsize(session.PTY, &pty.Winsize{
		Cols: cols,
		Rows: rows,
	})
}

// DetachSession 断开会话（tmux 保活）
func (m *Manager) DetachSession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[id]
	if !ok {
		return nil // 已经不存在
	}

	// 关闭 PTY 文件句柄（但不杀死进程）
	if session.PTY != nil {
		session.PTY.Close()
		session.PTY = nil
	}

	// 更新状态
	session.State = StateDetached
	m.emitState(session)

	// 从内存中移除（tmux session 仍在后台运行）
	delete(m.sessions, id)

	return nil
}

// DestroySession 彻底销毁会话
func (m *Manager) DestroySession(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[id]
	if !ok {
		return nil // 已经不存在
	}

	// 关闭 PTY
	if session.PTY != nil {
		session.PTY.Close()
	}

	// 如果使用 tmux，杀死 session
	if session.TmuxSession != "" && m.hasTmux {
		exec.Command(m.tmuxPath, "kill-session", "-t", session.TmuxSession).Run()
	}

	// 等待进程结束
	if session.Cmd != nil && session.Cmd.Process != nil {
		session.Cmd.Process.Kill()
	}

	// 发送销毁事件
	session.State = StateDestroyed
	m.emitState(session)

	delete(m.sessions, id)
	return nil
}

// handleSessionExit 处理会话退出（进程自然结束）
func (m *Manager) handleSessionExit(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[id]
	if !ok {
		return
	}

	session.State = StateExited
	m.emitState(session)

	// 清理资源
	if session.PTY != nil {
		session.PTY.Close()
	}

	delete(m.sessions, id)
}

// CloseAll 销毁所有会话
func (m *Manager) CloseAll() {
	m.mu.Lock()
	// 先收集所有 ID，避免遍历时修改
	ids := make([]string, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	m.mu.Unlock()

	// 逐个销毁（不持锁调用 DestroySession）
	for _, id := range ids {
		m.DestroySession(id)
	}
}

// ListSessions 列出所有活跃会话
func (m *Manager) ListSessions() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var infos []SessionInfo
	for _, session := range m.sessions {
		infos = append(infos, SessionInfo{
			ID:         session.ID,
			WorktreeID: session.WorktreeID,
			CWD:        session.CWD,
			State:      string(session.State),
			CreatedAt:  session.CreatedAt,
			LastActive: session.LastActive,
		})
	}
	return infos
}

// GetSessionState 获取会话状态
func (m *Manager) GetSessionState(id string) (SessionState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	if !ok {
		return "", fmt.Errorf("session not found: %s", id)
	}
	return session.State, nil
}

// HasTmux 返回 tmux 是否可用
func (m *Manager) HasTmux() bool {
	return m.hasTmux
}
