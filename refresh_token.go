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


func (db *DB) RevokeToken(refreshToken string) (error) {
	dbs, err := db.loadDB()
	if err != nil {
		return err
	}
	_, ok := dbs.RefreshTokens[refreshToken]
	if ok {
		delete(dbs.RefreshTokens, refreshToken)
	}
	db.writeDB(dbs)
	return nil

}

func (db *DB) GetRefreshToken(refreshToken string) (*RefreshToken, error) {
	dbs, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	token, ok := dbs.RefreshTokens[refreshToken]
	if !ok {
		return nil, errors.New("token not found")
	}

	return &token, nil
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

	generatedToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	newToken := RefreshToken{
		UserID: user.ID,
		Token: generatedToken,
		ExpirationTime: time.Now().UTC().AddDate(0,0,60),
	}
	dbs.RefreshTokens[generatedToken] = newToken
	db.writeDB(dbs)
	return &newToken, nil
}

func generateRefreshToken() (string, error) {
	randBytes := make([]byte, 32)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randBytes), nil
}
