package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/babylonchain/babylon-finality-gadget/verifier/db"
)

type Server struct {
	port 			string
	pg 				*db.PostgresHandler
}

type ServerConfig struct {
	Port 				string
	ConnString 	string
}

func StartServer(ctx context.Context, cfg *ServerConfig) (*Server, error) {
	// Create Postgres handler.
	pg, err := db.NewPostgresHandler(ctx, cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres handler: %w", err)
	}

	// Define server.
	s := &Server{
		port: cfg.Port,
		pg:   pg,
	}

	// Create router.
	router := mux.NewRouter().StrictSlash(true)
	
	// Define routes.
	router.HandleFunc("/getBlockStatusByHeight", s.getBlockStatusByHeight)
	router.HandleFunc("/getBlockStatusByHash", s.getBlockStatusByHash)
	router.HandleFunc("/getLatest", s.getLatestConsecutivelyFinalizedBlock)

	// Start server.
	log.Printf("Starting server on port %s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
			log.Fatalf("Could not start server: %s\n", err.Error())
	}

	return s, nil
}