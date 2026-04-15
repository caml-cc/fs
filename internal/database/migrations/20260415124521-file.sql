-- +migrate Up
CREATE TABLE
    IF NOT EXISTS FILES (
        filename TEXT NOT NULL,
        size INTEGER NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        expires_at DATETIME DEFAULT NOT NULL
    );
-- +migrate Down
DROP TABLE IF EXISTS FILES;