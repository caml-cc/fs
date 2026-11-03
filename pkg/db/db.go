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
	Size       int64
	Complete   bool
	IsDir      bool
}

func AddFile(id string, filename string, expirey time.Time, size int64, complete bool, isDir bool) error {
	_, err := database.DB.Exec(
		"INSERT INTO FILES (id, filename, expires_at, size, complete, is_dir) VALUES (?, ?, ?, ?, ?, ?);",
		id, filename, expirey, size, complete, isDir,
	)
	return err
}

func GetFile(id string) (StoredFile, error) {
	var file StoredFile
	err := database.DB.QueryRow(
		"SELECT id, filename, expires_at, size, complete, is_dir FROM FILES WHERE id = ?;", id,
	).Scan(&file.ID, &file.Filename, &file.Expires_At, &file.Size, &file.Complete, &file.IsDir)
	if err != nil {
		return StoredFile{}, err
	}

	return file, nil
}

func MarkComplete(id string) error {
	_, err := database.DB.Exec("UPDATE FILES SET complete = 1 WHERE id = ?;", id)
	return err
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
	rows, err := database.DB.Query("SELECT id, filename, expires_at, complete FROM FILES;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]StoredFile, 0)
	for rows.Next() {
		var file StoredFile
		if err := rows.Scan(&file.ID, &file.Filename, &file.Expires_At, &file.Complete); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
