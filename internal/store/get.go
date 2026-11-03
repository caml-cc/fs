package store

import (
	"fs/pkg/db"
	"fs/pkg/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func GetFile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "missing file id", http.StatusBadRequest)
		return
	}

	stored, err := db.GetFile(id)
	if err != nil || !stored.Complete {
		http.NotFound(w, r)
		return
	}

	f, err := os.Open(filepath.Join(utils.Conf.UPLOAD_DIR, filepath.Base(id)))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+stored.Filename)
	if stored.IsDir {
		w.Header().Set("X-Folder", "1")
	}

	http.ServeContent(w, r, stored.Filename, info.ModTime(), f)
}
