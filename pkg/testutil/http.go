package testutil

import (
	_ "embed"
	"net/http"
)

func BuildLocalServer() *http.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testHTML))
	})
	srv := http.NewServeMux()
	srv.HandleFunc("/", handler)
	return &http.Server{
		Handler: srv,
	}
}

//go:embed test_index.html
var testHTML string
