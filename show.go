package main

import (
	"database/sql"
	"fmt"
	"time"
)

type CommandDetail struct {
	ID         int64
	Command    string
	Cwd        string
	ExecutedAt time.Time
}

func showTag(db *sql.DB, tagID int64) error {
	var note, createdAtStr string
	err := db.QueryRow(`SELECT note, created_at FROM tags WHERE id = ?`, tagID).Scan(&note, &createdAtStr)
	if err == sql.ErrNoRows {
		return fmt.Errorf("tag #%d not found", tagID)
	} else if err != nil {
		return err
	}

	createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
	timeFormatted := createdAt.Local().Format("2006-01-02 15:04:05")

	fmt.Printf("Tag #%d: %q\n", tagID, note)
	fmt.Printf("Created: %s\n\n", timeFormatted)

	rows, err := db.Query(`SELECT id, command, cwd, executed_at FROM commands WHERE tag_id = ? ORDER BY id ASC`, tagID)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 1
	for rows.Next() {
		var cd CommandDetail
		var execStr string
		if err := rows.Scan(&cd.ID, &cd.Command, &cd.Cwd, &execStr); err != nil {
			continue
		}
		execTime, _ := time.Parse(time.RFC3339, execStr)
		fmt.Printf("  %d. %s\n", i, cd.Command)
		if cd.Cwd != "" {
			fmt.Printf("     cwd: %s\n", cd.Cwd)
		}
		fmt.Printf("     time: %s\n", execTime.Local().Format("15:04:05"))
		i++
	}

	return nil
}

func listRecent(db *sql.DB, limit int) error {
	if limit <= 0 {
		limit = 10
	}

	rows, err := db.Query(`SELECT id, note, created_at FROM tags ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tags []TagResult
	for rows.Next() {
		var tr TagResult
		var createdAtStr string
		if err := rows.Scan(&tr.ID, &tr.Note, &createdAtStr); err == nil {
			tr.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
			tags = append(tags, tr)
		}
	}

	if len(tags) == 0 {
		fmt.Println("No tags created yet.")
		return nil
	}

	for _, t := range tags {
		cmdRows, err := db.Query(`SELECT command FROM commands WHERE tag_id = ? ORDER BY id ASC`, t.ID)
		if err == nil {
			for cmdRows.Next() {
				var cmd string
				if err := cmdRows.Scan(&cmd); err == nil {
					t.Commands = append(t.Commands, cmd)
				}
			}
			cmdRows.Close()
		}

		timeFormatted := t.CreatedAt.Local().Format("2006-01-02 15:04")
		fmt.Printf("[#%d] %s  %q\n", t.ID, timeFormatted, t.Note)
		for i, cmd := range t.Commands {
			fmt.Printf("  %d. %s\n", i+1, cmd)
		}
		fmt.Println()
	}

	return nil
}
