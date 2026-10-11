package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

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
	if err := migrateLegacy(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db, q: New(db)}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// migrateLegacy upgrades a database from before accounts existed: quests gain
// an empty user_id, which the first account to register then claims.
func migrateLegacy(db *sql.DB) error {
	hasQuests, err := tableExists(db, "quests")
	if err != nil || !hasQuests {
		return err
	}
	var n int
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('quests') WHERE name = 'user_id'`).Scan(&n)
	if err != nil || n > 0 {
		return err
	}
	_, err = db.Exec(`ALTER TABLE quests ADD COLUMN user_id TEXT NOT NULL DEFAULT ''`)
	return err
}

func tableExists(db interface {
	QueryRow(string, ...any) *sql.Row
}, name string) (bool, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&n)
	return n > 0, err
}

// RegisterUser creates the account and its profile. The very first account
// inherits any quests and profile from before accounts existed; an account that
// ends up with no quests gets the seed quests.
func (s *Store) RegisterUser(ctx context.Context, username, passwordHash string) (User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()
	q := s.q.WithTx(tx)

	existing, err := q.CountUsers(ctx)
	if err != nil {
		return User{}, err
	}
	user, err := q.CreateUser(ctx, CreateUserParams{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return User{}, err
	}

	location, note := "", ""
	if existing == 0 {
		if err := q.ClaimUnownedQuests(ctx, user.ID); err != nil {
			return User{}, err
		}
		if legacy, err := tableExists(tx, "profile"); err != nil {
			return User{}, err
		} else if legacy {
			_ = tx.QueryRowContext(ctx, `SELECT location, note FROM profile WHERE id = 1`).Scan(&location, &note)
		}
	}
	if err := q.CreateProfile(ctx, CreateProfileParams{UserID: user.ID, Location: location, Note: note}); err != nil {
		return User{}, err
	}

	count, err := q.CountQuests(ctx, user.ID)
	if err != nil {
		return User{}, err
	}
	if count == 0 {
		if err := seedFor(ctx, q, user.ID); err != nil {
			return User{}, err
		}
	}
	return user, tx.Commit()
}

func (s *Store) Queries() *Queries { return s.q }

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

// DeleteUser removes the account and everything that belongs to it, and returns
// the photo file names so the caller can delete the files.
func (s *Store) DeleteUser(ctx context.Context, userID string) ([]string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	q := s.q.WithTx(tx)

	rows, err := q.ListPhotoPaths(ctx, userID)
	if err != nil {
		return nil, err
	}
	photos := make([]string, 0, len(rows))
	for _, p := range rows {
		photos = append(photos, p.String)
	}
	if err := q.DeleteQuestsForUser(ctx, userID); err != nil {
		return nil, err
	}
	// Profile and sessions go with the user (ON DELETE CASCADE).
	if err := q.DeleteUser(ctx, userID); err != nil {
		return nil, err
	}
	return photos, tx.Commit()
}
