package main

import (
	"database/sql"
	"time"
)

func logCommand(db *sql.DB, cmd, cwd, sessionID string) error {
	cmd = redactSecrets(cmd)
	if shouldIgnoreCommand(cmd) {
		return nil
	}

	_, err := db.Exec(
		`INSERT INTO commands (command, cwd, executed_at, session_id) VALUES (?, ?, ?, ?)`,
		cmd, cwd, time.Now().UTC(), sessionID,
	)
	return err
}
