package server

import (
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Routes
	router.HandleFunc("/getBlockStatusByHeight", getBlockStatusByHeight)
	router.HandleFunc("/getBlockStatusByHash", getBlockStatusByHash)
	router.HandleFunc("/getLatest", getLatestConsecutivelyFinalizedBlock)

	return router
}