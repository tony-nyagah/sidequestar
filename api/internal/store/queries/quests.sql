-- name: ListQuestsByStatus :many
SELECT * FROM quests WHERE user_id = sqlc.arg(user_id) AND status = sqlc.arg(status) ORDER BY created_at DESC;

-- name: ListQuests :many
SELECT * FROM quests WHERE user_id = sqlc.arg(user_id) ORDER BY created_at DESC;

-- name: GetQuest :one
SELECT * FROM quests WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: CreateQuest :one
INSERT INTO quests (id, title, description, duration_bucket, source, status, tags, created_at, user_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: AcceptQuest :one
UPDATE quests SET status = 'active'
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) AND status = 'generated'
RETURNING *;

-- name: CompleteQuest :one
UPDATE quests
SET status = 'completed', completed_at = sqlc.arg(completed_at), photo_path = sqlc.arg(photo_path), xp_awarded = sqlc.arg(xp_awarded)
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) AND status = 'active'
RETURNING *;

-- name: DeleteQuest :exec
DELETE FROM quests WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

-- name: CountQuests :one
SELECT COUNT(*) AS count FROM quests WHERE user_id = sqlc.arg(user_id);

-- name: PhotoBelongsTo :one
SELECT COUNT(*) AS count FROM quests WHERE photo_path = sqlc.arg(photo_path) AND user_id = sqlc.arg(user_id);

-- name: TotalCompletedXp :one
SELECT CAST(COALESCE(SUM(xp_awarded), 0) AS INTEGER) AS total
FROM quests WHERE user_id = sqlc.arg(user_id) AND status = 'completed';

-- name: CountCompleted :one
SELECT COUNT(*) AS count FROM quests WHERE user_id = sqlc.arg(user_id) AND status = 'completed';

-- name: ClaimUnownedQuests :exec
UPDATE quests SET user_id = sqlc.arg(user_id) WHERE user_id = '';
