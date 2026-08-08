package main

import (
	"database/sql"
	"fmt"
	"time"
)

func tagLastN(db *sql.DB, note string, n int) (int64, int, error) {
	if note == "" {
		return 0, 0, fmt.Errorf("tag note cannot be empty")
	}
	if n <= 0 {
		n = 1
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	// 1. Insert new tag record
	res, err := tx.Exec(`INSERT INTO tags (note, created_at) VALUES (?, ?)`, note, time.Now().UTC())
	if err != nil {
		return 0, 0, err
	}

	tagID, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	// 2. Select last N untagged command IDs
	rows, err := tx.Query(`
		SELECT id FROM commands
		WHERE tag_id IS NULL
		ORDER BY id DESC
		LIMIT ?`, n)
	if err != nil {
		return 0, 0, err
	}

	var cmdIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			cmdIDs = append(cmdIDs, id)
		}
	}
	rows.Close()

	if len(cmdIDs) == 0 {
		// Fallback: if no untagged commands found, pick last N commands overall
		rows, err = tx.Query(`
			SELECT id FROM commands
			ORDER BY id DESC
			LIMIT ?`, n)
		if err != nil {
			return 0, 0, err
		}
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err == nil {
				cmdIDs = append(cmdIDs, id)
			}
		}
		rows.Close()
	}

	// 3. Link commands to tagID
	for _, id := range cmdIDs {
		if _, err := tx.Exec(`UPDATE commands SET tag_id = ? WHERE id = ?`, tagID, id); err != nil {
			return 0, 0, err
		}
	}

	// 4. Index in FTS5
	if _, err := tx.Exec(`INSERT INTO tags_fts(rowid, note) VALUES (?, ?)`, tagID, note); err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return tagID, len(cmdIDs), nil
}
