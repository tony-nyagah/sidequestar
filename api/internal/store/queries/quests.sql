-- name: ListQuestsByStatus :many
SELECT * FROM quests WHERE status = ? ORDER BY created_at DESC;

-- name: ListQuests :many
SELECT * FROM quests ORDER BY created_at DESC;

-- name: GetQuest :one
SELECT * FROM quests WHERE id = ?;

-- name: CreateQuest :one
INSERT INTO quests (id, title, description, duration_bucket, source, status, tags, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: AcceptQuest :one
UPDATE quests SET status = 'active' WHERE id = ? AND status = 'generated'
RETURNING *;

-- name: CompleteQuest :one
UPDATE quests
SET status = 'completed', completed_at = ?, photo_path = ?, xp_awarded = ?
WHERE id = ? AND status = 'active'
RETURNING *;

-- name: DeleteQuest :exec
DELETE FROM quests WHERE id = ?;

-- name: CountQuests :one
SELECT COUNT(*) AS count FROM quests;

-- name: GetProfile :one
SELECT * FROM profile WHERE id = 1;

-- name: UpdateProfile :one
UPDATE profile SET location = ?, note = ? WHERE id = 1
RETURNING *;

-- name: TotalCompletedXp :one
SELECT CAST(COALESCE(SUM(xp_awarded), 0) AS INTEGER) AS total FROM quests WHERE status = 'completed';

-- name: CountCompleted :one
SELECT COUNT(*) AS count FROM quests WHERE status = 'completed';
