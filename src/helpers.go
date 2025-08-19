package main

import (
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func writeJSON(w http.ResponseWriter, status int, payload APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func snapshot(store *kvStore, snapshotPath string) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	file, err := os.Create(snapshotPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(store.data)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Snapshot saved successfully")
}

func loadSnapshot(store *kvStore, snapshotPath string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	file, err := os.Open(snapshotPath)
	if os.IsNotExist(err) {
		log.Println("No snapshot found, starting fresh")
		return
	}

	decoder := gob.NewDecoder(file)
	defer file.Close()

	if err := decoder.Decode(&store.data); err != nil {
		log.Println(err)
	}
}
