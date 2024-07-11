package server

import (
	"log"
	"net/http"
	"os"
)

func StartServer() {
	router := NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
			port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
			log.Fatalf("Could not start server: %s\n", err.Error())
	}
}