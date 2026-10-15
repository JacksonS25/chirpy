package main

import (
	"log"
	"net/http"
)

const filepathRoot = "."
const port = "8080"

func main() {
	serveMux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(filepathRoot))

	serveMux.Handle("/", fileServer)

	server := http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
