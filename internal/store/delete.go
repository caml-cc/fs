package store

import (
	"database/sql"
	"errors"
	"fs/pkg/db"
	"fs/pkg/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("K")
	if key != utils.Conf.API_KEY {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "missing file id", http.StatusBadRequest)
		return
	}

	path := filepath.Join(uploadDir, filepath.Base(id))
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete file", http.StatusInternalServerError)
		return
	}

	if err := db.DeleteFile(id); err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "failed to delete file metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
