package main

import (
	"net/http"
	"time"
)

func main() {
	kvStore := getKVStore()
	http.HandleFunc("/get/", getHandler(kvStore))
	http.HandleFunc("/put/", putHandler(kvStore))
	http.HandleFunc("/delete/", deleteHandler(kvStore))
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			snapshot(kvStore)
		}
	}()

	http.ListenAndServe(":8080", nil)
}
