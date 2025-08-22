package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

type replicateOp struct {
	Key   string
	Value interface{}
	Type  string
}

func isLeader(cfg nodeConfig) bool {
	return cfg.LeaderAddr == cfg.SelfAddr
}

func discoverPeers(serviceName string) ([]string, error) {
	var peers []string

	ips, err := net.LookupHost(serviceName)
	if err != nil {
		return nil, err
	}

	port := os.Getenv("PEER_PORT")
	if port == "" {
		port = "8080"
	}

	for _, ip := range ips {
		peers = append(peers, fmt.Sprintf("%s:%s", ip, port))
	}

	return peers, nil
}

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

func replicateToPeers(cfg nodeConfig, op replicateOp) error {
	body, err := json.Marshal(op)
	if err != nil {
		return fmt.Errorf("failed to marshal operation: %w", err)
	}

	for _, peer := range cfg.Peers {
		url := "http://" + peer + "/replicate"
		resp, err := http.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to replicate to %s: %w", peer, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("replication to %s failed with status: %d", peer, resp.StatusCode)
		}
		resp.Body.Close()
	}
	return nil
}
func forwardToLeader(cfg nodeConfig, w http.ResponseWriter, r *http.Request) {
	url := "http://" + cfg.LeaderAddr + r.URL.RequestURI()
	req, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
