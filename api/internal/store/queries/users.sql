-- name: CreateUser :one
INSERT INTO users (id, username, password_hash, created_at) VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ?;

-- name: CountUsers :one
SELECT COUNT(*) AS count FROM users;

-- name: CreateSession :exec
INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?);

-- name: GetSessionUser :one
SELECT users.* FROM sessions JOIN users ON users.id = sessions.user_id
WHERE sessions.token_hash = sqlc.arg(token_hash) AND sessions.expires_at > sqlc.arg(now);

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= ?;

-- name: CreateProfile :exec
INSERT INTO profiles (user_id, location, note) VALUES (?, ?, ?)
ON CONFLICT (user_id) DO NOTHING;

-- name: GetProfile :one
SELECT * FROM profiles WHERE user_id = ?;

-- name: UpdateProfile :one
UPDATE profiles SET location = sqlc.arg(location), note = sqlc.arg(note)
WHERE user_id = sqlc.arg(user_id)
RETURNING *;

-- name: SaveGeneratedSummary :exec
UPDATE profiles SET summary = sqlc.arg(summary), summary_updated_at = sqlc.arg(summary_updated_at)
WHERE user_id = sqlc.arg(user_id);

-- name: SaveEditedSummary :exec
UPDATE profiles
SET summary = sqlc.arg(summary), summary_user_edit = sqlc.arg(summary), summary_updated_at = sqlc.arg(summary_updated_at)
WHERE user_id = sqlc.arg(user_id);

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: UpdatePasswordHash :exec
UPDATE users SET password_hash = sqlc.arg(password_hash) WHERE id = sqlc.arg(id);

-- name: DeleteSessionsForUser :exec
DELETE FROM sessions WHERE user_id = ?;

-- name: DeleteOtherSessions :exec
DELETE FROM sessions WHERE user_id = sqlc.arg(user_id) AND token_hash != sqlc.arg(token_hash);

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;
