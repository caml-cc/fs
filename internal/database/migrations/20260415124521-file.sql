-- +migrate Up
CREATE TABLE
    IF NOT EXISTS FILES (
        id TEXT NOT NULL,
        filename TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        expires_at DATETIME NOT NULL
    );
-- +migrate Down
DROP TABLE IF EXISTS FILES;