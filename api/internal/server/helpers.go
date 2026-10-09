package server

import (
	"database/sql"
	"errors"
	"strings"
)

func isNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNullInt(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

const maxTags = 6

// normalizeTags lowercases, trims and de-duplicates tags so "Music" and
// "music" count as one interest.
func normalizeTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
		if len(out) == maxTags {
			break
		}
	}
	return out
}
