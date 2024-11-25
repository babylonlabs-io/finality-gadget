package server

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func (s *Server) newHttpHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/transaction", s.txStatusHandler)
	mux.HandleFunc("/v1/chainSyncStatus", s.chainSyncStatusHandler)
	mux.HandleFunc("/health", s.healthHandler)
	return mux
}

func (s *Server) txStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	txHash := r.URL.Query().Get("hash")
	s.logger.Info("/v1/transaction", zap.String("txHash", txHash))

	// Get block from rpc.
	txInfo, err := s.fg.QueryTransactionStatus(txHash)
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
	s.logger.Info("/v1/chainSyncStatus")
	// Get block from rpc.
	chainSyncStatus, err := s.fg.QueryChainSyncStatus()
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
	s.logger.Info("/health")
	response := "Finality gadget is healthy"
	_, err := w.Write([]byte(response))
	if err != nil {
		s.logger.Error("Failed to write response", zap.Error(err))
	}
}
