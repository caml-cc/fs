package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"fs/internal/models"
	"fs/internal/store"
)

func StartServer(conf models.Config) {
	server := mux.NewRouter()

	server.HandleFunc("/", store.AddFile).Methods(http.MethodPost)
	server.HandleFunc("/fragment", store.StartFragmentedUpload).Methods(http.MethodPost)
	server.HandleFunc("/{id}", store.AppendFragment).Methods(http.MethodPatch)
	server.HandleFunc("/{id}", store.DeleteFile).Methods(http.MethodDelete)
	server.HandleFunc("/{id}", store.GetFile).Methods(http.MethodGet)
	server.HandleFunc("/", store.List).Methods("LIST")

	log.Printf("server running on port: %s (max upload per request: %d bytes)\n", conf.PORT, conf.MAX_UPLOAD)
	log.Fatal(http.ListenAndServe(":"+conf.PORT, server))
}
