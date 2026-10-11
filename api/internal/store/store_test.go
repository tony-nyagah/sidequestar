package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// A database from before accounts existed keeps its quests and profile:
// the first account to register inherits them instead of getting seed quests.
func TestFirstAccountInheritsPreAccountData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	old, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = old.Exec(`
		CREATE TABLE quests (
			id TEXT PRIMARY KEY, title TEXT NOT NULL, description TEXT NOT NULL,
			duration_bucket TEXT NOT NULL, source TEXT NOT NULL, status TEXT NOT NULL,
			tags TEXT NOT NULL DEFAULT '[]', photo_path TEXT, xp_awarded INTEGER,
			created_at TEXT NOT NULL, completed_at TEXT);
		CREATE TABLE profile (id INTEGER PRIMARY KEY, location TEXT NOT NULL DEFAULT '', note TEXT NOT NULL DEFAULT '');
		INSERT INTO profile (id, location, note) VALUES (1, 'Nairobi', 'Keep it cheap');
		INSERT INTO quests (id, title, description, duration_bucket, source, status, xp_awarded, created_at, completed_at)
		VALUES ('q1', 'Old walk', 'Walk', 'under-30m', 'hand', 'completed', 30, '2026-10-01T00:00:00Z', '2026-10-01T01:00:00Z');`)
	old.Close()
	if err != nil {
		t.Fatal(err)
	}

	st, err := Open(path)
	if err != nil {
		t.Fatalf("open migrated store: %v", err)
	}
	defer st.Close()
	ctx := context.Background()

	first, err := st.RegisterUser(ctx, "tony", "hash")
	if err != nil {
		t.Fatal(err)
	}
	quests, _ := st.Queries().ListQuests(ctx, first.ID)
	if len(quests) != 1 || quests[0].Title != "Old walk" {
		t.Fatalf("first account quests = %+v, want only the old quest", quests)
	}
	p, _ := st.Queries().GetProfile(ctx, first.ID)
	if p.Location != "Nairobi" || p.Note != "Keep it cheap" {
		t.Fatalf("first account profile = %+v", p)
	}

	second, err := st.RegisterUser(ctx, "amina", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := st.Queries().CountQuests(ctx, second.ID); n != int64(len(seeds)) {
		t.Fatalf("second account has %d quests, want %d seeds", n, len(seeds))
	}
	if p, _ := st.Queries().GetProfile(ctx, second.ID); p.Note != "" {
		t.Fatalf("second account inherited note %q", p.Note)
	}
}
