package db

import (
	"database/sql"
	"fs/internal/database"
	"time"
)

type StoredFile struct {
	ID         string
	Filename   string
	Expires_At time.Time
}

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

func GetExpiredFiles(at time.Time) ([]StoredFile, error) {
	rows, err := database.DB.Query("SELECT id FROM FILES WHERE expires_at <= ?;", at)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expired := make([]StoredFile, 0)
	for rows.Next() {
		var file StoredFile
		if err := rows.Scan(&file.ID); err != nil {
			return nil, err
		}
		expired = append(expired, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return expired, nil
}

func ListFiles() ([]StoredFile, error) {
	rows, err := database.DB.Query("SELECT id, filename, expires_at FROM FILES;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]StoredFile, 0)
	for rows.Next() {
		var file StoredFile
		if err := rows.Scan(&file.ID, &file.Filename, &file.Expires_At); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
