// Package db handles database operations with SQLite.
package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database wraps gorm.DB.
type Database struct {
	*gorm.DB
}

// Init initializes the database.
func Init(path string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate
	db.AutoMigrate(
		&UsageRecord{},
		&Session{},
	)

	return &Database{db}, nil
}

// UsageRecord tracks API usage (Phase 4).
type UsageRecord struct {
	gorm.Model
	ProfileName  string
	Provider     string
	ModelName    string
	InputTokens  int
	OutputTokens int
	Cost         float64
	Timestamp    int64
}

// Session stores terminal session info (Phase 3).
type Session struct {
	gorm.Model
	SessionID   string `gorm:"uniqueIndex"`
	WorktreeID  string `gorm:"index"`
	CWD         string
	TmuxSession string // tmux session 名称
	Cols        uint16 // 终端列数
	Rows        uint16 // 终端行数
	Active      bool   // 是否活跃（应用关闭时保存，启动时恢复）
	StartedAt   int64
}

// SaveSessionRecord 保存或更新会话记录
func (d *Database) SaveSessionRecord(sessionID, worktreeID, cwd, tmuxSession string, cols, rows uint16, active bool) error {
	session := Session{
		SessionID:   sessionID,
		WorktreeID:  worktreeID,
		CWD:         cwd,
		TmuxSession: tmuxSession,
		Cols:        cols,
		Rows:        rows,
		Active:      active,
	}

	// 使用 GORM 的 FirstOrCreate 或 Updates
	result := d.Where(Session{SessionID: sessionID}).Assign(session).FirstOrCreate(&session)
	return result.Error
}

// SessionRecord 用于返回给 terminal 包
type SessionRecord struct {
	SessionID   string
	WorktreeID  string
	CWD         string
	TmuxSession string
	Cols        uint16
	Rows        uint16
	Active      bool
}

// GetActiveSessionRecords 获取所有活跃的会话记录
func (d *Database) GetActiveSessionRecords() ([]SessionRecord, error) {
	var sessions []Session
	result := d.Where("active = ?", true).Find(&sessions)
	if result.Error != nil {
		return nil, result.Error
	}

	records := make([]SessionRecord, len(sessions))
	for i, s := range sessions {
		records[i] = SessionRecord{
			SessionID:   s.SessionID,
			WorktreeID:  s.WorktreeID,
			CWD:         s.CWD,
			TmuxSession: s.TmuxSession,
			Cols:        s.Cols,
			Rows:        s.Rows,
			Active:      s.Active,
		}
	}
	return records, nil
}

// MarkSessionInactive 将会话标记为非活跃
func (d *Database) MarkSessionInactive(sessionID string) error {
	result := d.Model(&Session{}).Where("session_id = ?", sessionID).Update("active", false)
	return result.Error
}
