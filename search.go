package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type TagResult struct {
	ID        int64
	Note      string
	CreatedAt time.Time
	Commands  []string
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func search(db *sql.DB, query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	// Prepare FTS search query (support substring / wildcard search)
	ftsQuery := query
	if !strings.Contains(query, "*") && !strings.Contains(query, "\"") {
		terms := strings.Fields(query)
		for i, t := range terms {
			terms[i] = t + "*"
		}
		ftsQuery = strings.Join(terms, " ")
	}

	rows, err := db.Query(`
		SELECT t.id, t.note, t.created_at
		FROM tags_fts f
		JOIN tags t ON t.id = f.rowid
		WHERE tags_fts MATCH ?
		ORDER BY t.created_at DESC`, ftsQuery)
	if err != nil {
		// Fallback to standard LIKE search if FTS syntax error
		rows, err = db.Query(`
			SELECT id, note, created_at
			FROM tags
			WHERE note LIKE ?
			ORDER BY created_at DESC`, "%"+query+"%")
		if err != nil {
			return err
		}
	}
	defer rows.Close()

	var results []TagResult
	for rows.Next() {
		var r TagResult
		var createdAtStr string
		if err := rows.Scan(&r.ID, &r.Note, &createdAtStr); err != nil {
			continue
		}

		if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			r.CreatedAt = t
		} else if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			r.CreatedAt = t
		}

		cmdRows, err := db.Query(`SELECT command FROM commands WHERE tag_id = ? ORDER BY id ASC`, r.ID)
		if err == nil {
			for cmdRows.Next() {
				var cmd string
				if err := cmdRows.Scan(&cmd); err == nil {
					r.Commands = append(r.Commands, cmd)
				}
			}
			cmdRows.Close()
		}

		results = append(results, r)
	}

	if len(results) == 0 {
		fmt.Printf(Colorize(Yellow, "No fixes found matching %q.\n"), query)
		return nil
	}

	for _, res := range results {
		timeFormatted := res.CreatedAt.Local().Format("2006-01-02 15:04")
		fmt.Printf("%s %s  %s\n",
			Colorize(Yellow, fmt.Sprintf("[#%d]", res.ID)),
			Colorize(Gray, timeFormatted),
			Colorize(Magenta, stripANSI(res.Note)))
		for i := len(res.Commands) - 1; i >= 0; i-- {
			fmt.Printf("  %d. %s\n", len(res.Commands)-i, Colorize(Cyan, stripANSI(res.Commands[i])))
		}
		fmt.Println()
	}

	return nil
}
