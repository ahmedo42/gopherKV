package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	store := &kvStore{
		data: make(map[string]interface{}),
	}

	loadSnapshot(store)

	router := mux.NewRouter()
	router.HandleFunc("/put/{key}", putHandler(store)).Methods("PUT")
	router.HandleFunc("/get/{key}", getHandler(store)).Methods("GET")
	router.HandleFunc("/delete/{key}", deleteHandler(store)).Methods("DELETE")

	ticker := time.NewTicker(1 * time.Minute)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		for range ticker.C {
			snapshot(store)
		}
	}()

	go func() {
		<-quit
		log.Println("Shutting down... saving final snapshot")
		snapshot(store)
		ticker.Stop()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Println("Server started on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server exited")
}
