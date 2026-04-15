package db

import (
	"database/sql"
	"fs/internal/database"
	"time"
)

func AddFile(id string, filename string, expirey time.Time) error {
	_, err := database.DB.Exec("INSERT INTO FILES (id, filename, expires_at) VALUES ($1, $2, $3);", id, filename, expirey)
	return err
}

func GetFilename(id string) (string, error) {
	var filename string
	err := database.DB.QueryRow("SELECT filename FROM FILES WHERE id = ?;", id).Scan(&filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func DeleteFile(id string) error {
	result, err := database.DB.Exec("DELETE FROM FILES WHERE id = ?;", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
