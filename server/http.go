package server

import (
	"encoding/json"
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

func (s *Server) chainSyncStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Get block from rpc.
	chainSyncStatus, err := s.rpcServer.fg.QueryChainSyncStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(chainSyncStatus)
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
	response := "Finality gadget is healthy"
	_, err := w.Write([]byte(response))
	if err != nil {
		s.logger.Error("Failed to write response", zap.Error(err))
	}
}
