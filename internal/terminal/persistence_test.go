package terminal

import (
	"path/filepath"
	"testing"
	"time"

	"agent-orch/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB 创建临时 SQLite 数据库用于测试
func setupTestDB(t *testing.T) *db.Database {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.Init(dbPath)
	require.NoError(t, err, "failed to init test database")

	t.Cleanup(func() {
		sqlDB, _ := database.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	return database
}

// === 真实数据库测试 ===

func TestDatabase_SaveSessionRecord_CRUD(t *testing.T) {
	database := setupTestDB(t)

	// Create
	err := database.SaveSessionRecord("session-1", "wt-1", "/path/to/cwd", "tmux-1", 120, 40, true)
	require.NoError(t, err)

	// Read - 验证保存的数据
	records, err := database.GetActiveSessionRecords()
	require.NoError(t, err)
	require.Len(t, records, 1)

	record := records[0]
	assert.Equal(t, "session-1", record.SessionID)
	assert.Equal(t, "wt-1", record.WorktreeID)
	assert.Equal(t, "/path/to/cwd", record.CWD)
	assert.Equal(t, "tmux-1", record.TmuxSession)
	assert.Equal(t, uint16(120), record.Cols)
	assert.Equal(t, uint16(40), record.Rows)
	assert.True(t, record.Active)
}

func TestDatabase_SaveSessionRecord_Update(t *testing.T) {
	database := setupTestDB(t)

	// 首次保存
	err := database.SaveSessionRecord("session-1", "wt-1", "/path/a", "tmux-1", 80, 24, true)
	require.NoError(t, err)

	// 更新同一 session
	err = database.SaveSessionRecord("session-1", "wt-1", "/path/b", "tmux-1", 120, 40, true)
	require.NoError(t, err)

	// 验证只有一条记录，且已更新
	records, err := database.GetActiveSessionRecords()
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "/path/b", records[0].CWD)
	assert.Equal(t, uint16(120), records[0].Cols)
}

func TestDatabase_GetActiveSessionRecords_Filter(t *testing.T) {
	database := setupTestDB(t)

	// 添加多条记录
	err := database.SaveSessionRecord("s1", "wt-1", "/a", "", 80, 24, true)
	require.NoError(t, err)
	err = database.SaveSessionRecord("s2", "wt-2", "/b", "", 80, 24, true)
	require.NoError(t, err)
	err = database.SaveSessionRecord("s3", "wt-1", "/c", "", 80, 24, false) // 非活跃
	require.NoError(t, err)

	records, err := database.GetActiveSessionRecords()
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

func TestDatabase_MarkSessionInactive(t *testing.T) {
	database := setupTestDB(t)

	err := database.SaveSessionRecord("s1", "wt-1", "/a", "", 80, 24, true)
	require.NoError(t, err)

	// 标记为非活跃
	err = database.MarkSessionInactive("s1")
	require.NoError(t, err)

	// 验证不再返回该记录
	records, err := database.GetActiveSessionRecords()
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestDatabase_MarkSessionInactive_NotExist(t *testing.T) {
	database := setupTestDB(t)

	// 标记不存在的 session 不应报错
	err := database.MarkSessionInactive("nonexistent")
	assert.NoError(t, err)
}

// === Manager 持久化测试（使用真实数据库）===

func TestManager_SaveSession_NoDB(t *testing.T) {
	m := &Manager{
		sessions: make(map[string]*Session),
		db:       nil,
	}

	err := m.SaveSession("nonexistent")
	assert.NoError(t, err) // 没有 db 时不报错
}

func TestManager_SaveSession_WithDB(t *testing.T) {
	database := setupTestDB(t)

	m := &Manager{
		sessions: map[string]*Session{
			"session-1": {
				ID:          "session-1",
				WorktreeID:  "wt-1",
				CWD:         "/path/to/cwd",
				State:       StateRunning,
				TmuxSession: "tmux-session-1",
				CreatedAt:   time.Now(),
				LastActive:  time.Now(),
			},
		},
		db: database,
	}

	err := m.SaveSession("session-1")
	require.NoError(t, err)

	// 验证数据已保存到数据库
	records, err := database.GetActiveSessionRecords()
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "session-1", records[0].SessionID)
	assert.Equal(t, "wt-1", records[0].WorktreeID)
}

func TestManager_SaveSession_NotFound(t *testing.T) {
	database := setupTestDB(t)

	m := &Manager{
		sessions: make(map[string]*Session),
		db:       database,
	}

	err := m.SaveSession("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
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
	database := setupTestDB(t)

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
		db: database,
	}

	err := m.SaveAllSessions()
	require.NoError(t, err)

	// 实现语义：
	// - Running 会话保存为 active=true
	// - Detached 会话会被保存，但 active=false（不会出现在 active 列表）
	records, err := database.GetActiveSessionRecords()
	require.NoError(t, err)
	require.Len(t, records, 1)

	ids := make(map[string]bool)
	for _, r := range records {
		ids[r.SessionID] = true
	}
	assert.True(t, ids["s1"])
	assert.False(t, ids["s3"])

	// 验证 s2 确实被保存，只是 active=false
	sqlDB, err := database.DB.DB()
	require.NoError(t, err)

	var count int
	err = sqlDB.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id = ?", "s2").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
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
	database := setupTestDB(t)

	m := &Manager{
		sessions: make(map[string]*Session),
		db:       database,
	}

	err := m.RestoreSessions()
	require.NoError(t, err)
	assert.Empty(t, m.sessions)
}

func TestManager_RestoreSessions_WithRecords(t *testing.T) {
	database := setupTestDB(t)

	// 预先保存会话记录
	err := database.SaveSessionRecord("session-1", "wt-1", "/path/to/cwd", "", 120, 40, true)
	require.NoError(t, err)

	// 创建 Manager（无 tmux，会尝试创建新会话）
	m := &Manager{
		sessions: make(map[string]*Session),
		db:       database,
		ctx:      nil, // 无 context，不会发送事件
		hasTmux:  false,
	}

	// 注意：RestoreSessions 会尝试创建 PTY，在没有真实终端的环境下会失败
	// 但我们应该验证它至少读取了数据库并尝试恢复
	// 由于无法创建 PTY，session 会被标记为 inactive
	_ = m.RestoreSessions()

	// 验证：由于无法创建 PTY，session 应被标记为 inactive
	records, err := database.GetActiveSessionRecords()
	require.NoError(t, err)
	// 会话创建失败后应被标记为 inactive
	assert.Empty(t, records)
}
