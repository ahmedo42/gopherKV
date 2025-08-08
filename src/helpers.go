package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

func snapshot(kvStore map[string]interface{}) {
	file, err := os.Create("./snapshot.gob")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(kvStore)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Snapshot saved successfully")
}

func getKVStore() map[string]interface{} {
	kvStore := make(map[string]interface{})
	file, err := os.Open("./snapshot.gob")

	if err != nil {
		return kvStore
	}

	decoder := gob.NewDecoder(file)
	defer file.Close()

	if err := decoder.Decode(&kvStore); err != nil {
		fmt.Println(err)
	}

	return kvStore
}
