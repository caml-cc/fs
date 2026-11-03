package store

import (
	"errors"
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

// multipartOverhead is the slack allowed on top of MAX_UPLOAD for multipart
// boundaries, headers and small form fields such as "keep".
const multipartOverhead int64 = 1 << 20

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

	maxUpload := utils.Conf.MAX_UPLOAD

	// Reject oversized uploads before reading any of the body. Because the body
	// is never read, Go does not send "100 Continue", so clients like curl stop
	// before transmitting the file.
	if r.ContentLength > maxUpload+multipartOverhead {
		uploadTooLarge(w)
		return
	}

	// Covers chunked / unknown-length bodies.
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload+multipartOverhead)

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			uploadTooLarge(w)
			return
		}
		http.Error(w, "request must include a multipart file named file", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "request must include a multipart file named file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if fileHeader.Size < 0 {
		http.Error(w, "invalid file size", http.StatusBadRequest)
		return
	}

	if fileHeader.Size > maxUpload {
		uploadTooLarge(w)
		return
	}

	expiry, err := resolveExpiry(r.FormValue("keep"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uploadDir := utils.Conf.UPLOAD_DIR
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		http.Error(w, "failed to prepare upload directory", http.StatusInternalServerError)
		return
	}

	if ok, err := hasStorageFor(fileHeader.Size); err != nil {
		http.Error(w, "failed to inspect storage usage", http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, "insufficient storage", http.StatusInsufficientStorage)
		return
	}

	id := randomString(10)
	filename := filepath.Base(fileHeader.Filename)
	path := filepath.Join(uploadDir, id)
	isDir := r.FormValue("dir") == "1"

	err = db.AddFile(id, filename, expiry, fileHeader.Size, true, isDir)
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

func uploadTooLarge(w http.ResponseWriter) {
	msg := fmt.Sprintf(
		"file exceeds the max upload size of %s; use cc-store to upload larger files in fragments",
		formatBytes(utils.Conf.MAX_UPLOAD),
	)
	http.Error(w, msg, http.StatusRequestEntityTooLarge)
}

// resolveExpiry turns an optional Go duration string (e.g. "100h") into an
// expiry time, falling back to the configured KEEP period.
func resolveExpiry(keep string) (time.Time, error) {
	if keep == "" {
		return time.Now().Add(time.Duration(utils.Conf.KEEP) * time.Second), nil
	}

	keepDuration, err := time.ParseDuration(keep)
	if err != nil {
		return time.Time{}, errors.New("invalid keep value")
	}
	if keepDuration <= 0 {
		return time.Time{}, errors.New("keep must be a positive duration")
	}

	return time.Now().Add(keepDuration), nil
}

func hasStorageFor(size int64) (bool, error) {
	currentUsage, err := getDirSize(utils.Conf.UPLOAD_DIR)
	if err != nil {
		return false, err
	}

	maxStorageBytes := int64(utils.Conf.MAX_STORAGE) << 30
	return currentUsage+size <= maxStorageBytes, nil
}

func formatBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.2f GiB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.2f MiB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.2f KiB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
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
