package main

import (
	"database/sql"
	"fmt"
	"regexp"
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

	fmt.Printf("%s %s\n",
		Colorize(Yellow, fmt.Sprintf("Tag #%d:", tagID)),
		Colorize(Magenta, note))
	fmt.Printf("%s %s\n\n", Colorize(Gray, "Created:"), Colorize(Gray, timeFormatted))

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
		fmt.Printf("  %d. %s\n", i, Colorize(Cyan, stripANSI(cd.Command)))
		if cd.Cwd != "" {
			fmt.Printf("     %s %s\n", Colorize(Gray, "cwd:"), stripANSI(cd.Cwd))
		}
		fmt.Printf("     %s %s\n", Colorize(Gray, "time:"), Colorize(Gray, execTime.Local().Format("15:04:05")))
		i++
	}

	return nil
}

func listCommands(db *sql.DB, limit int) error {
	if limit <= 0 {
		limit = 10
	}
	rows, err := db.Query(`SELECT id, command, cwd, executed_at FROM commands ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var cd CommandDetail
		var execStr string
		if err := rows.Scan(&cd.ID, &cd.Command, &cd.Cwd, &execStr); err != nil {
			continue
		}
		execTime, _ := time.Parse(time.RFC3339, execStr)
		count++
		fmt.Printf("%d. [%d] %s\n", count, cd.ID, Colorize(Cyan, stripANSI(cd.Command)))
		if cd.Cwd != "" {
			fmt.Printf("   %s %s\n", Colorize(Gray, "cwd:"), stripANSI(cd.Cwd))
		}
		fmt.Printf("   %s %s\n", Colorize(Gray, "time:"), Colorize(Gray, execTime.Local().Format("2006-01-02 15:04:05")))
	}

	return nil
}

func listCommandsByTag(db *sql.DB, tagID int64, limit int) error {
	if limit <= 0 {
		limit = 10
	}

	rows, err := db.Query(`SELECT id, command, cwd, executed_at FROM commands WHERE tag_id = ? ORDER BY id DESC LIMIT ?`, tagID, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var cd CommandDetail
		var execStr string
		if err := rows.Scan(&cd.ID, &cd.Command, &cd.Cwd, &execStr); err != nil {
			continue
		}
		execTime, _ := time.Parse(time.RFC3339, execStr)
		count++
		fmt.Printf("%d. [%d] %s\n", count, cd.ID, Colorize(Cyan, stripANSI(cd.Command)))
		if cd.Cwd != "" {
			fmt.Printf("   %s %s\n", Colorize(Gray, "cwd:"), stripANSI(cd.Cwd))
		}
		fmt.Printf("   %s %s\n", Colorize(Gray, "time:"), Colorize(Gray, execTime.Local().Format("2006-01-02 15:04:05")))
	}

	if count == 0 {
		fmt.Printf(Colorize(Yellow, "No commands found for tag #%d.\n"), tagID)
	}
	return nil
}

func deleteCommand(db *sql.DB, commandID int64) error {
	_, err := db.Exec(`DELETE FROM commands WHERE id = ?`, commandID)
	return err
}

func updateTagNote(db *sql.DB, tagID int64, note string) error {
	if note == "" {
		return fmt.Errorf("tag note cannot be empty")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE tags SET note = ? WHERE id = ?`, note, tagID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tag #%d not found", tagID)
	}

	if _, err := tx.Exec(`INSERT INTO tags_fts(tags_fts, rowid, note) VALUES('delete', ?, ?)`, tagID, note); err != nil {
		_, _ = tx.Exec(`DELETE FROM tags_fts WHERE rowid = ?`, tagID)
	}
	if _, err := tx.Exec(`INSERT INTO tags_fts(rowid, note) VALUES (?, ?)`, tagID, note); err != nil {
		return err
	}

	return tx.Commit()
}

func listTags(db *sql.DB, limit int) error {
	if limit <= 0 {
		limit = 10
	}

	rows, err := db.Query(`SELECT id, note, created_at FROM tags ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var tr TagResult
		var createdAtStr string
		if err := rows.Scan(&tr.ID, &tr.Note, &createdAtStr); err != nil {
			continue
		}
		tr.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		timeFormatted := tr.CreatedAt.Local().Format("2006-01-02 15:04")
		count++
		fmt.Printf("%s %s  %s\n",
			Colorize(Yellow, fmt.Sprintf("[#%d]", tr.ID)),
			Colorize(Gray, timeFormatted),
			Colorize(Magenta, stripANSI(tr.Note)))
	}

	if count == 0 {
		fmt.Println("No tags created yet.")
	}
	return nil
}

func searchHistory(db *sql.DB, pattern string, limit int) error {
	if limit <= 0 {
		limit = 10
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex: %w", err)
	}

	rows, err := db.Query(`SELECT id, command, cwd, executed_at FROM commands ORDER BY id DESC`)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var cd CommandDetail
		var execStr string
		if err := rows.Scan(&cd.ID, &cd.Command, &cd.Cwd, &execStr); err != nil {
			continue
		}
		if !re.MatchString(cd.Command) {
			continue
		}

		execTime, _ := time.Parse(time.RFC3339, execStr)
		count++
		fmt.Printf("%d. %s\n", count, Colorize(Cyan, stripANSI(cd.Command)))
		if cd.Cwd != "" {
			fmt.Printf("   %s %s\n", Colorize(Gray, "cwd:"), stripANSI(cd.Cwd))
		}
		fmt.Printf("   %s %s\n", Colorize(Gray, "time:"), Colorize(Gray, execTime.Local().Format("2006-01-02 15:04:05")))

		if count >= limit {
			break
		}
	}

	if count == 0 {
		fmt.Printf(Colorize(Yellow, "No history entries matching %q.\n"), pattern)
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
		fmt.Printf("%s %s  %s\n",
			Colorize(Yellow, fmt.Sprintf("[#%d]", t.ID)),
			Colorize(Gray, timeFormatted),
			Colorize(Magenta, stripANSI(t.Note)))
		for i := len(t.Commands) - 1; i >= 0; i-- {
			fmt.Printf("  %d. %s\n", len(t.Commands)-i, Colorize(Cyan, stripANSI(t.Commands[i])))
		}
		fmt.Println()
	}

	return nil
}
