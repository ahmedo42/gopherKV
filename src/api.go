package main

import (
	"fmt"
	"net/http"
	"sync"
)

type kvStore struct {
	data map[interface{}]interface{}
	mu   sync.RWMutex
}

type APIResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func getHandler(store *kvStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")

		store.mu.RLock()
		defer store.mu.RUnlock()
		val, ok := store.data[key]
		if !ok {
			writeJSON(w, http.StatusNotFound, APIResponse{Message: "no such key"})
			return
		}

		data := map[interface{}]interface{}{
			"key":   key,
			"value": val,
		}
		fmt.Println(data)

		writeJSON(w, http.StatusOK, APIResponse{Message: "success", Data: data})
	}
}

func putHandler(store *kvStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("value")

		if key == "" {
			writeJSON(w, http.StatusBadRequest, APIResponse{Message: "Empty Key"})
			return
		}
		store.mu.Lock()
		defer store.mu.Unlock()
		store.data[key] = val
		response := APIResponse{
			Message: "success",
			Data: map[interface{}]interface{}{
				"key":   key,
				"value": val,
			},
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func deleteHandler(store *kvStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")

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

		store.mu.Lock()
		defer store.mu.Unlock()
		delete(store.data, key)
		writeJSON(w, http.StatusOK, APIResponse{Message: "success"})
	}
}
