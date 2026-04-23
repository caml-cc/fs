
-- +migrate Up
ALTER TABLE FILES RENAME TO FILES_old;

CREATE TABLE FILES (
	id TEXT NOT NULL,
	filename TEXT NOT NULL,
	created_at DATETIME DEFAULT (datetime('now', 'localtime')),
	expires_at DATETIME NOT NULL
);

INSERT INTO FILES (id, filename, created_at, expires_at)
SELECT id, filename, created_at, expires_at
FROM FILES_old;

DROP TABLE FILES_old;

-- +migrate Down
ALTER TABLE FILES RENAME TO FILES_new;

CREATE TABLE FILES (
	id TEXT NOT NULL,
	filename TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	expires_at DATETIME NOT NULL
);

INSERT INTO FILES (id, filename, created_at, expires_at)
SELECT id, filename, created_at, expires_at
FROM FILES_new;

DROP TABLE FILES_new;
