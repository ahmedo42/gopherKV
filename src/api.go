package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func getHandler(kvStore map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val, ok := kvStore[key]
		var response APIResponse
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			response.Message = fmt.Sprintf("no such key: '%s'", key)

			json.NewEncoder(w).Encode(response)
			return
		}

		data := map[string]interface{}{
			"key":   key,
			"value": val,
		}
		response.Message = "success"
		response.Data = data

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	}
}

func putHandler(kvStore map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("value")

		if key == "" || val == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			response := APIResponse{
				Message: "key and value must not be empty",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		kvStore[key] = val
		response := APIResponse{
			Message: "success",
			Data: map[string]interface{}{
				"key":   key,
				"value": val,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func deleteHandler(kvStore map[string]interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")

		if key == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			response := APIResponse{
				Message: "key must not be empty",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		_, ok := kvStore[key]
		var response APIResponse
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			response.Message = fmt.Sprintf("no such key: '%s'", key)
			json.NewEncoder(w).Encode(response)
			return
		}

		delete(kvStore, key)
		response.Message = "success"
		response.Data = nil

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
