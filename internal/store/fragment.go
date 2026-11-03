package store

import (
	"database/sql"
	"errors"
	"fmt"
	"fs/pkg/db"
	"fs/pkg/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// Fragmented uploads let clients send files larger than MAX_UPLOAD:
//
//	POST  /fragment  (Upload-Length, Upload-Name, [Upload-Keep], [Upload-Dir])
//	                 -> 201, Upload-Id + Upload-Max-Chunk headers
//	PATCH /{id}      (Upload-Offset, raw body <= Upload-Max-Chunk)
//	                 -> 204 + Upload-Offset while incomplete
//	                 -> 201 + file URL once the last byte arrives
//
// Fragments are appended to a single file on disk, so a finished upload is
// indistinguishable from a normal one.

var fragmentLocks sync.Map // id -> *sync.Mutex

func lockFragment(id string) func() {
	m, _ := fragmentLocks.LoadOrStore(id, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func StartFragmentedUpload(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("K") != utils.Conf.API_KEY {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	size, err := strconv.ParseInt(r.Header.Get("Upload-Length"), 10, 64)
	if err != nil || size < 0 {
		http.Error(w, "Upload-Length header must be a non-negative integer", http.StatusBadRequest)
		return
	}

	filename := filepath.Base(r.Header.Get("Upload-Name"))
	if filename == "" || filename == "." || filename == "/" {
		http.Error(w, "Upload-Name header is required", http.StatusBadRequest)
		return
	}

	expiry, err := resolveExpiry(r.Header.Get("Upload-Keep"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uploadDir := utils.Conf.UPLOAD_DIR
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		http.Error(w, "failed to prepare upload directory", http.StatusInternalServerError)
		return
	}

	if ok, err := hasStorageFor(size); err != nil {
		http.Error(w, "failed to inspect storage usage", http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, "insufficient storage", http.StatusInsufficientStorage)
		return
	}

	id := randomString(10)
	output, err := os.Create(filepath.Join(uploadDir, id))
	if err != nil {
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	output.Close()

	isDir := r.Header.Get("Upload-Dir") == "1"
	if err := db.AddFile(id, filename, expiry, size, false, isDir); err != nil {
		fmt.Println(err)
		os.Remove(filepath.Join(uploadDir, id))
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Upload-Id", id)
	w.Header().Set("Upload-Max-Chunk", strconv.FormatInt(utils.Conf.MAX_UPLOAD, 10))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id + "\n"))
}

func AppendFragment(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("K") != utils.Conf.API_KEY {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	offset, err := strconv.ParseInt(r.Header.Get("Upload-Offset"), 10, 64)
	if err != nil || offset < 0 {
		http.Error(w, "Upload-Offset header must be a non-negative integer", http.StatusBadRequest)
		return
	}

	unlock := lockFragment(id)
	defer unlock()

	stored, err := db.GetFile(id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "failed to read file metadata", http.StatusInternalServerError)
		return
	}

	if stored.Complete {
		http.Error(w, "upload is already complete", http.StatusConflict)
		return
	}

	path := filepath.Join(utils.Conf.UPLOAD_DIR, filepath.Base(id))
	info, err := os.Stat(path)
	if err != nil {
		http.Error(w, "upload file is missing", http.StatusInternalServerError)
		return
	}

	current := info.Size()
	if offset != current {
		w.Header().Set("Upload-Offset", strconv.FormatInt(current, 10))
		http.Error(w, "Upload-Offset does not match the uploaded size", http.StatusConflict)
		return
	}

	remaining := stored.Size - offset
	limit := min(utils.Conf.MAX_UPLOAD, remaining)

	// Reject before reading so the client does not send the body.
	if r.ContentLength > utils.Conf.MAX_UPLOAD {
		uploadTooLarge(w)
		return
	}
	if r.ContentLength > remaining {
		http.Error(w, "fragment is larger than the remaining upload size", http.StatusRequestEntityTooLarge)
		return
	}

	body := http.MaxBytesReader(w, r.Body, limit)

	output, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		http.Error(w, "failed to open upload file", http.StatusInternalServerError)
		return
	}

	written, copyErr := io.Copy(output, body)
	closeErr := output.Close()
	if copyErr != nil || closeErr != nil {
		// Roll back the partial fragment so the client can retry from offset.
		os.Truncate(path, offset)

		var maxErr *http.MaxBytesError
		if errors.As(copyErr, &maxErr) {
			http.Error(w, "fragment is larger than allowed", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "failed to save fragment", http.StatusInternalServerError)
		return
	}

	newOffset := offset + written
	if newOffset < stored.Size {
		w.Header().Set("Upload-Offset", strconv.FormatInt(newOffset, 10))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := db.MarkComplete(id); err != nil {
		http.Error(w, "failed to finalize upload", http.StatusInternalServerError)
		return
	}
	fragmentLocks.Delete(id)

	w.Header().Set("Upload-Offset", strconv.FormatInt(newOffset, 10))
	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(r.Host + "/" + id + "\n"))
}
