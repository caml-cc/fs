package store

import (
	"fmt"
	"fs/pkg/db"
	"fs/pkg/utils"
	"net/http"
)

func List(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("K")
	if key != utils.Conf.API_KEY {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != "LIST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	files, err := db.ListFiles()
	if err != nil {
		http.Error(w, "unable to list files", http.StatusInternalServerError)
	}

	w.WriteHeader(200)
	for _, file := range files {
		fmt.Fprintf(w, "%s | %s | %s\n", file.ID, file.Filename, file.Expires_At.Format("2006-01-02 15:04:05"))
	}
}
