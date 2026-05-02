package database

import (
	"database/sql"
	"embed"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
	migrate "github.com/rubenv/sql-migrate"
)

var DB *sql.DB

//go:embed migrations/*.sql
var migrationFiles embed.FS

func InitSQLiteDB() error {
	var err error

	dbDir := "./internal/database"

	if err = os.MkdirAll(dbDir, 0o755); err != nil {
		return errors.New("failed to prepare upload directory")
	}

	DB, err = sql.Open("sqlite3", "./internal/database/fs.db")
	if err != nil {
		return err
	}

	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationFiles,
		Root:       "migrations",
	}

	_, err = migrate.Exec(DB, "sqlite3", migrations, migrate.Up)
	return err
}

func DbClose() {
	if DB != nil {
		DB.Close()
	}
}
