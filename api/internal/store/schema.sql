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
    completed_at    TEXT
);

CREATE TABLE IF NOT EXISTS profile (
    id       INTEGER PRIMARY KEY CHECK (id = 1),
    location TEXT NOT NULL DEFAULT '',
    note     TEXT NOT NULL DEFAULT ''
);

INSERT INTO profile (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
