package server

import (
	"database/sql"
	"errors"
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
