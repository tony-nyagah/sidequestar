package store

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	db *sql.DB
	q  *Queries
}

func Open(path string) (*Store, error) {
	dsn := "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db, q: New(db)}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Queries() *Queries { return s.q }

// timestampLayout is fixed-width so timestamps sort correctly as text.
const timestampLayout = "2006-01-02T15:04:05.000000000Z07:00"

// Timestamp formats t (in UTC) for the created_at/completed_at columns.
func Timestamp(t time.Time) string {
	return t.UTC().Format(timestampLayout)
}

func ParseTags(raw string) []string {
	var tags []string
	if raw == "" {
		return tags
	}
	_ = json.Unmarshal([]byte(raw), &tags)
	return tags
}

func MarshalTags(tags []string) string {
	if tags == nil {
		tags = []string{}
	}
	b, _ := json.Marshal(tags)
	return string(b)
}
