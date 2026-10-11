CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
    password_hash TEXT NOT NULL,
    created_at    TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS profiles (
    user_id            TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    location           TEXT NOT NULL DEFAULT '',
    note               TEXT NOT NULL DEFAULT '',
    summary            TEXT NOT NULL DEFAULT '',
    summary_updated_at TEXT,
    -- The summary as the person last edited it by hand. Rewrites are given this,
    -- never the model's own previous output, so invented details don't snowball.
    summary_user_edit  TEXT NOT NULL DEFAULT ''
);

-- user_id is last because older databases gain it via ALTER TABLE ADD COLUMN,
-- and sqlc reads SELECT * by position. It is '' only for quests from before
-- accounts existed; the first account to register claims them.
CREATE TABLE IF NOT EXISTS quests (
    id              TEXT PRIMARY KEY,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL,
    duration_bucket TEXT NOT NULL,
    source          TEXT NOT NULL,
    status          TEXT NOT NULL,
    tags            TEXT NOT NULL DEFAULT '[]',
    photo_path      TEXT,
    xp_awarded      INTEGER,
    created_at      TEXT NOT NULL,
    completed_at    TEXT,
    user_id         TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS quests_user_status ON quests (user_id, status);
