package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type nodeConfig struct {
	SelfAddr   string
	LeaderAddr string
	Peers      []string
}

func main() {

	snapshotPath := flag.String("path", getEnv("SNAPSHOT_PATH", "./"), "path to the directory containing the periodic snapshot")
	addrFlag := flag.String("addr", getEnv("ADDR", "0.0.0.0:8080"), "this node's host:port")
	leaderFlag := flag.String("leader", getEnv("LEADER", "0.0.0.0:8080"), "leader host:port")
	flag.Parse()

	cfg := nodeConfig{
		SelfAddr:   *addrFlag,
		LeaderAddr: *leaderFlag,
	}
	role := os.Getenv("ROLE")
	serviceName := os.Getenv("SERVICE_NAME")

	if role == "follower" {
		peers, err := discoverPeers(serviceName)
		if err != nil {
			log.Printf("Failed to discover peers: %v", err)
		} else {
			log.Printf("Discovered peers: %v", peers)
			cfg.Peers = peers
		}
	}

	if err := os.MkdirAll(*snapshotPath, 0755); err != nil {
		log.Fatalf("Failed to create snapshot directory %s: %v", *snapshotPath, err)
	}
	savePath := *snapshotPath + "/snapshot.gob"
	store := &kvStore{
		data: make(map[string]interface{}),
	}

	loadSnapshot(store, savePath)

	router := mux.NewRouter()
	router.HandleFunc("/put/{key}", func(w http.ResponseWriter, r *http.Request) {
		if !isLeader(cfg) {
			forwardToLeader(cfg, w, r)
			return
		}
		putHandler(store, cfg)(w, r)
	}).Methods("PUT")
	router.HandleFunc("/get/{key}", getHandler(store)).Methods("GET")
	router.HandleFunc("/delete/{key}", func(w http.ResponseWriter, r *http.Request) {
		if !isLeader(cfg) {
			forwardToLeader(cfg, w, r)
			return
		}
		deleteHandler(store, cfg)(w, r)
	}).Methods("DELETE")
	router.HandleFunc("/replicate", replicateHandler(store)).Methods("POST")

	ticker := time.NewTicker(1 * time.Minute)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.SelfAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		for range ticker.C {
			snapshot(store, savePath)
		}
	}()

	go func() {
		<-quit
		log.Println("Shutting down... saving final snapshot")
		snapshot(store, savePath)
		ticker.Stop()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Println("Server started on port", cfg.SelfAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server exited")
}
