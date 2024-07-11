package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)



type DB struct {
	path string
	currentID int
	mux *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}


func (db *DB) loadDB() (DBStructure, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	var chirps DBStructure = DBStructure{Chirps: map[int]Chirp{}}
	data, err := os.ReadFile(db.path)
	if err != nil {
		dat, _ := json.Marshal(chirps)
		os.WriteFile(db.path, dat, 0666)
		data, err = os.ReadFile(db.path)
	}
	err = json.Unmarshal(data, &chirps)
	if err != nil {
		log.Printf("error unmarshling json: %v", err)
		return chirps, err
	}
	db.currentID = len(chirps.Chirps)
	return chirps, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	data, err := json.Marshal(dbStructure)
	if err != nil {
		log.Printf("error unmarshling json in writeDB: %v", err)
		return err
	}

	err = os.WriteFile(db.path, data, 0666)
	if err != nil {
		log.Printf("error writing db to file: %v", err)
		return err
	}

	return nil

}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	log.Printf("creating new chirp: %v", body)
	db.currentID++
	chirp := Chirp{
		ID: db.currentID,
		Body: body,
	}
	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	log.Println("getting chirps")
	dbStructure, err := db.loadDB()
	if err != nil {
		log.Printf("error loading db in GetChirps: %v", err)
		return []Chirp{}, err
	}
	chirps := []Chirp{}

	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}
	return chirps, nil

}

func (db *DB) GetChirp(id int) (Chirp, bool) {
	dbs, _ := db.loadDB()
	chirp, ok := dbs.Chirps[id]
	if !ok {
		return Chirp{}, false
	}
	return chirp, true
}
