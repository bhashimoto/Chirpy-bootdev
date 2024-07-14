package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)



type DB struct {
	path string
	mux *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users map[int]UserCredential `json:"users"`
	RefreshTokens map[RefreshToken]int `json:"refresh_tokens"`
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

func (db *DB) UpdateRefreshToken(id int, token string) (error) {
	dbs, err := db.loadDB()
	if err != nil  {
		return err
	}
	oldUser, found := dbs.Users[id]
	if !found {
		return errors.New("user not found")
	}

	
	user := UserCredential {
		User: User {
			ID: oldUser.ID,
			Email: oldUser.Email,
		},
		Password: oldUser.Password,

	}
	dbs.Users[id] = user
	err = db.writeDB(dbs)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateUser(id int, email string, password string) (User, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	oldUser, found := dbs.Users[id]
	if !found {
		return User{}, errors.New("user not found")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return User{}, err
	}

	user := UserCredential {
		User: User {
			ID: oldUser.ID,
			Email: email, 
		},
		Password: hashed,
	}
	dbs.Users[id] = user

	

	db.writeDB(dbs)
	return user.User, nil
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return User{}, nil
	}
	user := UserCredential{
		User: User{
			ID: len(dbs.Users) + 1,
			Email: email,
		},
		Password: hashed,
	}
	dbs.Users[user.ID] = user
	err = db.writeDB(dbs)
	if err != nil {
		return User{}, err
	}
	return user.User, nil
}

func (db *DB) GetUsers() ([]User, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return []User{}, err
	}
	users := []User{}

	for _, user := range dbs.Users {
		users = append(users, user.User)
	}
	return users, nil
}

func (db *DB) GetUserByEmail(email string) (UserCredential, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return UserCredential{}, err
	}

	for _, user := range dbs.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return UserCredential{}, errors.New("user not found")
}


func (db *DB) CreateChirp(body string) (Chirp, error) {
	log.Printf("creating new chirp: %v", body)
	dbStructure, _ := db.loadDB()
	chirp := Chirp{
		ID: len(dbStructure.Chirps) + 1,
		Body: body,
	}
	dbStructure.Chirps[chirp.ID] = chirp
	err := db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
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

func (db *DB) CreateDB() (error) {
	dbs := DBStructure {
		Chirps: map[int]Chirp{},
		Users: map[int]UserCredential{},
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
