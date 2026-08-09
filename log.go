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

func addTaggedCommand(db *sql.DB, cmd, cwd, note string) (int64, int, error) {
	// Create the tag and insert the command linked to it.
	tx, err := db.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO tags (note, created_at) VALUES (?, ?)`, note, time.Now().UTC())
	if err != nil {
		return 0, 0, err
	}

	tagID, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	redacted := redactSecrets(cmd)
	if _, err := tx.Exec(`INSERT INTO commands (command, cwd, executed_at, session_id, tag_id) VALUES (?, ?, ?, ?, ?)`, redacted, cwd, time.Now().UTC(), "", tagID); err != nil {
		return 0, 0, err
	}

	if _, err := tx.Exec(`INSERT INTO tags_fts(rowid, note) VALUES (?, ?)`, tagID, note); err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return tagID, 1, nil
}
