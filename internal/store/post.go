package store

import (
	"fmt"
	"fs/pkg/db"
	"fs/pkg/utils"
	"io"
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

	id := randomString(10)
	filename := filepath.Base(fileHeader.Filename)
	path := filepath.Join(uploadDir, id)

	err = db.AddFile(id, filename, time.Now().Add(time.Hour*24))
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
