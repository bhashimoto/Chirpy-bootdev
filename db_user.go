package main

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (db *DB) UpgradeUser(userID int) error {
	dbs, err := db.loadDB()
	if err != nil {
		return err
	}

	user, ok := dbs.Users[userID]
	if !ok {
		return errors.New("user not found")
	}

	user.IsChirpyRed = true
	dbs.Users[userID] = user
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
			IsChirpyRed: false,
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

func (db *DB) GetUser(id int) (User, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	user, ok := dbs.Users[id]
	if !ok {
		return User{}, errors.New("user not found")
	}
	return user.User, nil
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
