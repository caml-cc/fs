package store

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func GetFile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "missing file id", http.StatusBadRequest)
		return
	}

	path := filepath.Join(uploadDir, filepath.Base(id))
	http.ServeFile(w, r, path)
}
