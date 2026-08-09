package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func openDB() (*sql.DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".smaran")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dir, "history.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		note TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL,
		cwd TEXT,
		executed_at DATETIME NOT NULL,
		session_id TEXT,
		tag_id INTEGER REFERENCES tags(id) ON DELETE SET NULL
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS tags_fts USING fts5(
		note,
		content='tags',
		content_rowid='id'
	);

	CREATE INDEX IF NOT EXISTS idx_commands_tag_id ON commands(tag_id);
	CREATE INDEX IF NOT EXISTS idx_commands_executed_at ON commands(executed_at);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
