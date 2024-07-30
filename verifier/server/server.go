package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/babylonchain/babylon-finality-gadget/verifier/db"
)

type Server struct {
	db 				*db.BBoltHandler
	port 			string
}

type ServerConfig struct {
	Db  			*db.BBoltHandler
	Port 			string
}

func StartServer(cfg *ServerConfig) (*Server, error) {
	// Define server.
	s := &Server{
		port: cfg.Port,
		db:   cfg.Db,
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
			return nil, err
	}

	return s, nil
}