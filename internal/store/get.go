package store

import (
	"fs/pkg/db"
	"fs/pkg/utils"
	"io"
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

	filename, err := db.GetFilename(id)

	f, err := os.Open(filepath.Join(utils.Conf.UPLOAD_DIR, filepath.Base(id)))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/octet-stream")

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	io.Copy(w, f)
}
