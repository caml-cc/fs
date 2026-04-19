package store

import (
	"database/sql"
	"errors"
	"fs/pkg/db"
	"log"
	"os"
	"path/filepath"
	"time"
)

func StartExpiryCleanupWorker(interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}

	if deleted, err := CleanupExpiredFiles(); err != nil {
		log.Printf("cleanup startup run failed: %v", err)
	} else if deleted > 0 {
		log.Printf("cleanup removed %d expired file(s)", deleted)
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			deleted, err := CleanupExpiredFiles()
			if err != nil {
				log.Printf("cleanup run failed: %v", err)
				continue
			}
			if deleted > 0 {
				log.Printf("cleanup removed %d expired file(s)", deleted)
			}
		}
	}()
}

func CleanupExpiredFiles() (int, error) {
	expired, err := db.GetExpiredFiles(time.Now())
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, file := range expired {
		path := filepath.Join(uploadDir, filepath.Base(file.ID))
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Printf("cleanup failed to delete file %s: %v", file.ID, err)
			continue
		}

		err := db.DeleteFile(file.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("cleanup failed to delete database row for %s: %v", file.ID, err)
			continue
		}

		deleted++
	}

	return deleted, nil
}
