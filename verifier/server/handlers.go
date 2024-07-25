package server

import (
	"context"
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
	ctx := context.Background()
	isFinal := s.pg.GetBlockStatusByHeight(ctx, uint64(blockHeight))

	// Marshal and return status
	jsonResponse, err := json.Marshal(StatusResponse {
		IsFinalized: isFinal,
	})
	if err != nil {
		http.Error(w, "unable to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
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
	ctx := context.Background()
	isFinal := s.pg.GetBlockStatusByHash(ctx, hash)

	// Marshal and return status
	jsonResponse, err := json.Marshal(StatusResponse {
		IsFinalized: isFinal,
	})
	if err != nil {
		http.Error(w, "unable to marshal JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (s *Server) getLatestConsecutivelyFinalizedBlock(w http.ResponseWriter, r *http.Request) {
	// Fetch status from DB
	ctx := context.Background()
	block, err := s.pg.GetLatestConsecutivelyFinalizedBlock(ctx)
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
	w.Write(jsonResponse)
}

