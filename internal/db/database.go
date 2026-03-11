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
	WorktreeID  string
	CWD         string
	StartedAt   int64
}
