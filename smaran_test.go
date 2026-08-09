package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_history.db")

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
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
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to exec test schema: %v", err)
	}

	return db
}

func TestFilterAndRedact(t *testing.T) {
	if !shouldIgnoreCommand("ls -la") {
		t.Errorf("expected ls -la to be ignored")
	}
	if !shouldIgnoreCommand("cd /tmp") {
		t.Errorf("expected cd /tmp to be ignored")
	}
	if shouldIgnoreCommand("kubectl get pods") {
		t.Errorf("expected kubectl get pods to NOT be ignored")
	}

	rawCmd := "kubectl get secrets --token=Bearer eyJhbGciOiJIUzI1NiJ9"
	redacted := redactSecrets(rawCmd)
	if redacted == rawCmd {
		t.Errorf("expected command secrets to be redacted, got: %s", redacted)
	}
}

func TestLogTagSearchUntagFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 1. Log commands
	cmds := []string{
		"pd-ctl -u http://pd:2379 scheduler add balance-leader-scheduler",
		"kubectl exec -it pd-0 -- pd-ctl config set leader-schedule-limit 4",
	}

	for _, c := range cmds {
		err := logCommand(db, c, "/tmp", "test-session")
		if err != nil {
			t.Fatalf("failed to log command: %v", err)
		}
	}

	// Verify logged commands
	var count int
	db.QueryRow(`SELECT count(*) FROM commands`).Scan(&count)
	if count != 2 {
		t.Fatalf("expected 2 logged commands, got %d", count)
	}

	// 2. Tag commands
	note := "fixed region rebalance stall after tikv-3 crash"
	tagID, taggedCount, err := tagLastN(db, note, 2)
	if err != nil {
		t.Fatalf("failed to tag commands: %v", err)
	}
	if tagID <= 0 || taggedCount != 2 {
		t.Fatalf("expected tagID > 0 and taggedCount 2, got tagID %d, count %d", tagID, taggedCount)
	}

	// 3. Search tag
	err = search(db, "region rebalance")
	if err != nil {
		t.Fatalf("search returned error: %v", err)
	}

	// 4. Untag
	err = untag(db, tagID)
	if err != nil {
		t.Fatalf("untag returned error: %v", err)
	}

	var postUntagCount int
	db.QueryRow(`SELECT count(*) FROM tags`).Scan(&postUntagCount)
	if postUntagCount != 0 {
		t.Fatalf("expected 0 tags after untag, got %d", postUntagCount)
	}
}
