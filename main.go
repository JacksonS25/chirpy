package main

import (
	"log"
	"net/http"
)

const filepathRoot = "/app/"
const port = "8080"

func main() {
	serveMux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	serveMux.Handle(filepathRoot, http.StripPrefix("/app", fileServer))

	serveMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
