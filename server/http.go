package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

func (s *Server) txStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	txHash := r.URL.Query().Get("hash")
	s.logger.Debug("Received transaction hash", zap.String("txHash", txHash))

	// Get block from rpc.
	txInfo, err := s.rpcServer.fg.QueryTransactionStatus(txHash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(txInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		s.logger.Error("Failed to write response", zap.Error(err))
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	name := r.URL.Query().Get("name")
	age := r.URL.Query().Get("age")

	// Respond with the parameters
	response := fmt.Sprintf("Name: %s, Age: %s", name, age)
	_, err := w.Write([]byte(response))
	if err != nil {
		s.logger.Error("Failed to write response", zap.Error(err))
	}
}
