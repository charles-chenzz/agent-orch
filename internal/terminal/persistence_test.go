package terminal

import (
	"testing"
	"time"

	"agent-orch/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDB 实现 DBProvider 接口用于测试
type MockDB struct {
	records map[string]*db.SessionRecord
	saveErr error
	getErr  error
	markErr error
}

func NewMockDB() *MockDB {
	return &MockDB{
		records: make(map[string]*db.SessionRecord),
	}
}

func (m *MockDB) SaveSessionRecord(sessionID, worktreeID, cwd, tmuxSession string, cols, rows uint16, active bool) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.records[sessionID] = &db.SessionRecord{
		SessionID:   sessionID,
		WorktreeID:  worktreeID,
		CWD:         cwd,
		TmuxSession: tmuxSession,
		Cols:        cols,
		Rows:        rows,
		Active:      active,
	}
	return nil
}

func (m *MockDB) GetActiveSessionRecords() ([]db.SessionRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var result []db.SessionRecord
	for _, r := range m.records {
		if r.Active {
			result = append(result, *r)
		}
	}
	return result, nil
}

func (m *MockDB) MarkSessionInactive(sessionID string) error {
	if m.markErr != nil {
		return m.markErr
	}
	if r, ok := m.records[sessionID]; ok {
		r.Active = false
	}
	return nil
}

// === 测试用例 ===

func TestMockDB_SaveSessionRecord(t *testing.T) {
	db := NewMockDB()

	err := db.SaveSessionRecord("session-1", "wt-1", "/path/to/cwd", "tmux-1", 120, 40, true)
	require.NoError(t, err)

	record, ok := db.records["session-1"]
	require.True(t, ok)
	assert.Equal(t, "session-1", record.SessionID)
	assert.Equal(t, "wt-1", record.WorktreeID)
	assert.Equal(t, "/path/to/cwd", record.CWD)
	assert.Equal(t, "tmux-1", record.TmuxSession)
	assert.Equal(t, uint16(120), record.Cols)
	assert.Equal(t, uint16(40), record.Rows)
	assert.True(t, record.Active)
}

func TestMockDB_GetActiveSessionRecords(t *testing.T) {
	db := NewMockDB()

	// 添加多条记录
	db.SaveSessionRecord("s1", "wt-1", "/a", "", 80, 24, true)
	db.SaveSessionRecord("s2", "wt-2", "/b", "", 80, 24, true)
	db.SaveSessionRecord("s3", "wt-1", "/c", "", 80, 24, false) // 非活跃

	records, err := db.GetActiveSessionRecords()
	require.NoError(t, err)
	assert.Len(t, records, 2)

	// 验证只返回活跃记录
	ids := make(map[string]bool)
	for _, r := range records {
		ids[r.SessionID] = true
	}
	assert.True(t, ids["s1"])
	assert.True(t, ids["s2"])
	assert.False(t, ids["s3"])
}

func TestMockDB_MarkSessionInactive(t *testing.T) {
	db := NewMockDB()

	db.SaveSessionRecord("s1", "wt-1", "/a", "", 80, 24, true)

	err := db.MarkSessionInactive("s1")
	require.NoError(t, err)

	assert.False(t, db.records["s1"].Active)
}

func TestMockDB_SaveSessionRecord_Error(t *testing.T) {
	db := NewMockDB()
	db.saveErr = assert.AnError

	err := db.SaveSessionRecord("s1", "wt-1", "/a", "", 80, 24, true)
	assert.Error(t, err)
}

func TestMockDB_GetActiveSessionRecords_Error(t *testing.T) {
	db := NewMockDB()
	db.getErr = assert.AnError

	_, err := db.GetActiveSessionRecords()
	assert.Error(t, err)
}

// === Manager 持久化测试 ===

func TestManager_SaveSession_NoDB(t *testing.T) {
	// 没有 db 的情况下，SaveSession 应该返回 nil（静默跳过）
	m := &Manager{
		sessions: make(map[string]*Session),
		db:       nil,
	}

	err := m.SaveSession("nonexistent")
	assert.NoError(t, err) // 没有 db 时不报错
}

func TestManager_SaveAllSessions_NoDB(t *testing.T) {
	m := &Manager{
		sessions: make(map[string]*Session),
		db:       nil,
	}

	err := m.SaveAllSessions()
	assert.NoError(t, err)
}

func TestManager_SaveAllSessions_WithSessions(t *testing.T) {
	db := NewMockDB()
	m := &Manager{
		sessions: map[string]*Session{
			"s1": {
			ID:          "s1",
			WorktreeID:  "wt-1",
			CWD:         "/path/a",
			State:       StateRunning,
			TmuxSession: "tmux-s1",
			CreatedAt:   time.Now(),
			LastActive:  time.Now(),
		},
			"s2": {
			ID:          "s2",
			WorktreeID:  "wt-2",
			CWD:         "/path/b",
			State:       StateDetached,
			TmuxSession: "tmux-s2",
			CreatedAt:   time.Now(),
			LastActive:  time.Now(),
		},
			"s3": {
			ID:          "s3",
			WorktreeID:  "wt-1",
			CWD:         "/path/c",
			State:       StateExited, // 非活跃状态，不应保存
			CreatedAt:   time.Now(),
			LastActive:  time.Now(),
		},
		},
		db: db,
	}

	err := m.SaveAllSessions()
	require.NoError(t, err)

	// 验证只有 Running 和 Detached 状态的会话被保存
	assert.NotNil(t, db.records["s1"])
	assert.NotNil(t, db.records["s2"])
	assert.Nil(t, db.records["s3"])
}

func TestManager_RestoreSessions_NoDB(t *testing.T) {
	m := &Manager{
		sessions: make(map[string]*Session),
		db:       nil,
	}

	err := m.RestoreSessions()
	assert.NoError(t, err)
}

func TestManager_RestoreSessions_EmptyDB(t *testing.T) {
	db := NewMockDB()
	m := &Manager{
		sessions: make(map[string]*Session),
		db:       db,
	}

	err := m.RestoreSessions()
	require.NoError(t, err)
	assert.Empty(t, m.sessions)
}
