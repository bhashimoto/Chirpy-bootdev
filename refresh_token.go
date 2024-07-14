package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

type RefreshToken struct {
	UserID		int
	Token		string
	ExpirationTime	time.Time
}

func (db *DB) CreateRefreshToken(userID int) (*RefreshToken, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	user, found := dbs.Users[userID]
	if !found {
		return nil, errors.New("user not found")
	}

	tokenDuration, err := time.ParseDuration("60d")
	if err != nil {
		return nil, err
	}
	generatedToken, err := generateToken()
	if err != nil {
		return nil, err
	}

	newToken := RefreshToken{
		UserID: user.ID,
		Token: generatedToken,
		ExpirationTime: time.Now().UTC().Add(tokenDuration),
	}
	dbs.RefreshTokens[generatedToken] = user.ID
	db.writeDB(dbs)
	return &newToken, nil
}

func generateToken() (string, error) {
	randBytes := make([]byte, 32)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randBytes), nil
}
