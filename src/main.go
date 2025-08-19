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

func main() {

	snapshotPath := flag.String("path", getEnv("SNAPSHOT_PATH", "./"), "path to the directory containing the periodic snapshot")
	srvPort := flag.String("port", getEnv("PORT", "8080"), "where the API will be served")
	flag.Parse()

	if err := os.MkdirAll(*snapshotPath, 0755); err != nil {
		log.Fatalf("Failed to create snapshot directory %s: %v", *snapshotPath, err)
	}
	savePath := *snapshotPath + "/snapshot.gob"
	store := &kvStore{
		data: make(map[string]interface{}),
	}

	loadSnapshot(store, savePath)

	router := mux.NewRouter()
	router.HandleFunc("/put/{key}", putHandler(store)).Methods("PUT")
	router.HandleFunc("/get/{key}", getHandler(store)).Methods("GET")
	router.HandleFunc("/delete/{key}", deleteHandler(store)).Methods("DELETE")

	ticker := time.NewTicker(1 * time.Minute)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         ":" + *srvPort,
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

	log.Println("Server started on port", *srvPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server exited")
}
