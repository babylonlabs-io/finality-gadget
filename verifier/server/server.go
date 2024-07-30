package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/babylonchain/babylon-finality-gadget/verifier/db"
)

type Server struct {
	server *http.Server
	db     *db.BBoltHandler
}

type ServerConfig struct {
	Db   *db.BBoltHandler
	Port string
}

func Start(cfg *ServerConfig) (*Server, error) {
	// Create router.
	router := mux.NewRouter().StrictSlash(true)

	// Define server.
	s := &Server{
		server: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: router,
		},
		db: cfg.Db,
	}

	// Define routes.
	router.HandleFunc("/getBlockStatusByHeight", s.getBlockStatusByHeight)
	router.HandleFunc("/getBlockStatusByHash", s.getBlockStatusByHash)
	router.HandleFunc("/getLatest", s.getLatestConsecutivelyFinalizedBlock)

	// Start server in a goroutine.
	go func() {
		log.Printf("Starting server on port %s...", cfg.Port)
		if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
			log.Fatalf("Could not start server: %s\n", err.Error())
		}
	}()

	return s, nil
}

func (s *Server) Stop() error {
	log.Println("Stopping server...")
	return s.server.Close()
}
