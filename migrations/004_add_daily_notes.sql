CREATE TABLE IF NOT EXISTS daily_notes (
    date       TEXT PRIMARY KEY,  -- YYYY-MM-DD
    content    TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
