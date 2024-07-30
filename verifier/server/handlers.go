package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type StatusResponse struct {
	IsFinalized bool `json:"isFinalized"`
}

func (s *Server) getBlockStatusByHeight(w http.ResponseWriter, r *http.Request) {
	// Fetch params and run validation check
	blockHeightStr := r.URL.Query().Get("blockHeight")
	if blockHeightStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing required params: blockHeight\n")
		return
	}

	// Parse params
	blockHeight, err := strconv.Atoi(blockHeightStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid blockHeight\n")
		return
	}

	// Fetch status from DB
	isFinal := s.db.GetBlockStatusByHeight(uint64(blockHeight))

	// Marshal and return status
	jsonResponse, err := json.Marshal(StatusResponse{
		IsFinalized: isFinal,
	})
	if err != nil {
		http.Error(w, "unable to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResponse); err != nil {
		http.Error(w, "unable to write JSON response", http.StatusInternalServerError)
	}
}

func (s *Server) getBlockStatusByHash(w http.ResponseWriter, r *http.Request) {
	// Fetch params and run validation check
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing required params: hash\n")
		return
	}

	// Fetch status from DB
	isFinal := s.db.GetBlockStatusByHash(hash)

	// Marshal and return status
	jsonResponse, err := json.Marshal(StatusResponse{
		IsFinalized: isFinal,
	})
	if err != nil {
		http.Error(w, "unable to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResponse); err != nil {
		http.Error(w, "unable to write JSON response", http.StatusInternalServerError)
	}
}

func (s *Server) getLatestBlock(w http.ResponseWriter, r *http.Request) {
	// Fetch status from DB
	block, err := s.db.GetLatestBlock()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error getting latest block: %v\n", err)
		return
	}

	// Marshal and return status
	jsonResponse, err := json.Marshal(block)
	if err != nil {
		http.Error(w, "unable to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResponse); err != nil {
		http.Error(w, "unable to write JSON response", http.StatusInternalServerError)
	}
}
