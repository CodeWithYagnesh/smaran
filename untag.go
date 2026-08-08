package main

import (
	"database/sql"
	"fmt"
)

func untag(db *sql.DB, tagID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var note string
	err = tx.QueryRow(`SELECT note FROM tags WHERE id = ?`, tagID).Scan(&note)
	if err == sql.ErrNoRows {
		return fmt.Errorf("tag #%d does not exist", tagID)
	} else if err != nil {
		return err
	}

	// 1. Reset tag_id in commands
	res, err := tx.Exec(`UPDATE commands SET tag_id = NULL WHERE tag_id = ?`, tagID)
	if err != nil {
		return err
	}
	unlinkedCount, _ := res.RowsAffected()

	// 2. Remove from FTS index
	if _, err := tx.Exec(`INSERT INTO tags_fts(tags_fts, rowid, note) VALUES('delete', ?, ?)`, tagID, note); err != nil {
		// Fallback delete if old SQLite version
		_, _ = tx.Exec(`DELETE FROM tags_fts WHERE rowid = ?`, tagID)
	}

	// 3. Remove tag from tags table
	if _, err := tx.Exec(`DELETE FROM tags WHERE id = ?`, tagID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("removed tag #%d %q (unlinked %d commands)\n", tagID, note, unlinkedCount)
	return nil
}
