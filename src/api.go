package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type kvStore struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

type APIResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ValueRequest struct {
	Value interface{} `json:"value"`
}

func getHandler(store *kvStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := mux.Vars(r)["key"]

		store.mu.RLock()
		defer store.mu.RUnlock()
		val, ok := store.data[key]
		if !ok {
			writeJSON(w, http.StatusNotFound, APIResponse{Message: "no such key"})
			return
		}

		data := map[string]interface{}{
			"key":   key,
			"value": val,
		}

		writeJSON(w, http.StatusOK, APIResponse{Message: "success", Data: data})
	}
}

func putHandler(store *kvStore, cfg nodeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := mux.Vars(r)["key"]

		if key == "" {
			writeJSON(w, http.StatusBadRequest, APIResponse{Message: "Empty Key"})
			return
		}

		var req ValueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{Message: "Invalid JSON"})
			return
		}

		store.mu.Lock()
		defer store.mu.Unlock()
		store.data[key] = req.Value
		op := replicateOp{Type: "put", Key: key, Value: req.Value}
		if err := replicateToPeers(cfg, op); err != nil {
			http.Error(w, "replication failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		response := APIResponse{
			Message: "success",
			Data: map[string]interface{}{
				"key":   key,
				"value": req.Value,
			},
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func deleteHandler(store *kvStore, cfg nodeConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := mux.Vars(r)["key"]

		if key == "" {
			writeJSON(w, http.StatusBadRequest, APIResponse{Message: "Empty Key"})
			return
		}

		store.mu.Lock()
		defer store.mu.Unlock()
		_, ok := store.data[key]
		if !ok {
			writeJSON(w, http.StatusNotFound, APIResponse{Message: "no such key"})
			return
		}

		delete(store.data, key)
		op := replicateOp{Type: "delete", Key: key}
		if err := replicateToPeers(cfg, op); err != nil {
			http.Error(w, "replication failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, APIResponse{Message: "success"})
	}
}

func replicateHandler(store *kvStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var op replicateOp
		if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
			http.Error(w, "bad replication request", http.StatusBadRequest)
			return
		}
		switch op.Type {
		case "put":
			store.mu.Lock()
			store.data[op.Key] = op.Value
			store.mu.Unlock()
		case "delete":
			store.mu.Lock()
			delete(store.data, op.Key)
			store.mu.Unlock()
		default:
			http.Error(w, "unknown operation", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

}
