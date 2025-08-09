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

func snapshot(store *kvStore) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	file, err := os.Create("./snapshot.gob")
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

func loadSnapshot(store *kvStore) {
	store.mu.Lock()
	defer store.mu.Unlock()
	file, err := os.Open("./snapshot.gob")
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
