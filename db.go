package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"sync"
)



type DB struct {
	path string
	mux *sync.RWMutex
}

type DBStructure struct {
	Chirps		map[int]Chirp		`json:"chirps"`
	Users		map[int]UserCredential	`json:"users"`
	RefreshTokens	map[string]RefreshToken	`json:"refresh_tokens"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux: &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
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




func (db *DB) CreateChirp(body string, userID int) (Chirp, error) {
	log.Printf("creating new chirp: %v", body)
	dbStructure, _ := db.loadDB()
	chirp := Chirp{
		ID: len(dbStructure.Chirps) + 1,
		Body: body,
		AuthorID: userID,
	}
	dbStructure.Chirps[chirp.ID] = chirp
	err := db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}
	return chirp, nil
}

func (db *DB) GetChirps(userID int, sortDirection string) ([]Chirp, error) {
	log.Println("getting chirps")
	dbStructure, err := db.loadDB()
	if err != nil {
		log.Printf("error loading db in GetChirps: %v", err)
		return []Chirp{}, err
	}
	chirps := []Chirp{}

	for _, chirp := range dbStructure.Chirps {
		if userID == 0 || userID == chirp.AuthorID {
			chirps = append(chirps, chirp)
		}
	}

	if sortDirection == "asc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID > chirps[j].ID
		})
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

func (db *DB) DeleteChirp(chirpID int) error {
	dbs, err := db.loadDB()
	if err != nil {
		return err
	}
	_, ok := dbs.Chirps[chirpID]
	if ok {
		delete(dbs.Chirps, chirpID)
	}
	db.writeDB(dbs)
	return nil
}

func (db *DB) CreateDB() (error) {
	dbs := DBStructure {
		Chirps: map[int]Chirp{},
		Users: map[int]UserCredential{},
		RefreshTokens: map[string]RefreshToken{},
	}
	return db.writeDB(dbs)
}


func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.CreateDB()
	}
	return err
}
