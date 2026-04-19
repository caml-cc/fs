package store

import (
	"fmt"
	"fs/pkg/db"
	"fs/pkg/utils"
	"io"
	"io/fs"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const uploadDir = "./internal/uploads"

type fileResponse struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
}

func AddFile(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("K")
	if key != utils.Conf.API_KEY {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	const maxUploadSize int64 = 15 << 30
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, "request must include a multipart file named file (max 15GiB)", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "request must include a multipart file named file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		http.Error(w, "failed to prepare upload directory", http.StatusInternalServerError)
		return
	}

	currentUsage, err := getDirSize(uploadDir)
	if err != nil {
		http.Error(w, "failed to inspect storage usage", http.StatusInternalServerError)
		return
	}

	maxStorageBytes := int64(utils.Conf.MAX_STORAGE) << 30
	if fileHeader.Size < 0 {
		http.Error(w, "invalid file size", http.StatusBadRequest)
		return
	}

	if currentUsage+fileHeader.Size > maxStorageBytes {
		http.Error(w, "insufficient storage", http.StatusInsufficientStorage)
		return
	}

	id := randomString(10)
	filename := filepath.Base(fileHeader.Filename)
	path := filepath.Join(uploadDir, id)
	expiresAt := time.Now().Add(time.Duration(utils.Conf.KEEP) * time.Second)

	err = db.AddFile(id, filename, expiresAt)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	output, err := os.Create(path)
	if err != nil {
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer output.Close()

	_, err = io.Copy(output, file)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(r.Host + "/" + id + "\n"))
}

func randomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	rand.NewSource(time.Now().UnixNano())
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func getDirSize(root string) (int64, error) {
	var total int64
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		total += info.Size()
		return nil
	})
	if err != nil {
		return 0, err
	}

	return total, nil
}
