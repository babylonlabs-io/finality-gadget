package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/babylonchain/babylon-finality-gadget/verifier/db"
)

type StatusResponse struct {
	IsFinalized bool `json:"isFinalized"`
}

func getBlockStatusByHeight(w http.ResponseWriter, r *http.Request) {
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
	pg, err := db.NewPostgresHandler(ctx, os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating postgres handler: %v\n", err)
		return
	}
	isFinal := pg.GetBlockStatusByHeight(ctx, uint64(blockHeight))

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

func getBlockStatusByHash(w http.ResponseWriter, r *http.Request) {
	// Fetch params and run validation check
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Missing required params: hash\n")
		return
	}
	
	// Fetch status from DB
	ctx := context.Background()
	pg, err := db.NewPostgresHandler(ctx, os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating postgres handler: %v\n", err)
		return
	}
	isFinal := pg.GetBlockStatusByHash(ctx, hash)

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

func getLatestConsecutivelyFinalizedBlock(w http.ResponseWriter, r *http.Request) {
	// Fetch status from DB
	ctx := context.Background()
	pg, err := db.NewPostgresHandler(ctx, os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating postgres handler: %v\n", err)
		return
	}
	block, err := pg.GetLatestConsecutivelyFinalizedBlock(ctx)
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

