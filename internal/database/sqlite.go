package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	migrate "github.com/rubenv/sql-migrate"
)

var DB *sql.DB

func InitSQLiteDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "./internal/database/fs.db")
	if err != nil {
		return err
	}

	_, err = DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	migrations := &migrate.FileMigrationSource{
		Dir: "internal/database/migrations",
	}

	_, err = migrate.Exec(DB, "sqlite3", migrations, migrate.Up)
	return err
}

func DbClose() {
	if DB != nil {
		DB.Close()
	}
}
