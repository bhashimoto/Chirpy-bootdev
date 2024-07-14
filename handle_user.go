package main

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (cfg * apiConfig) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	
}

func (cfg *apiConfig) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password	 string `json:"password"`
		Email		 string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Println("Error at decode.Decode:", err)
		respondWithError(w, 400, "bad request")
		return
	}

	user, err := cfg.db.GetUserByEmail(params.Email)
	if err != nil {
		respondWithError(w, 404, "user not found")
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(params.Password))
	if err != nil {
		log.Println("Error at bcrypt.CompareHashAndPassword:", err)
		respondWithError(w, 401, "unauthorized")
		return
	}

	// No expiration time passed or time passed is over 24 hours
	if params.ExpiresInSeconds == 0 || params.ExpiresInSeconds > 60*60 {
		params.ExpiresInSeconds = 60*60
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, 
		jwt.RegisteredClaims{
			Issuer: "chirpy",
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Second*time.Duration(params.ExpiresInSeconds))),
			Subject: strconv.Itoa(user.ID),
		},
	)


	secretKey := []byte(os.Getenv("JWT_SECRET"))
	signedJWT, err := token.SignedString(secretKey)
	refreshToken := make([]byte, 32)
	_ , err = rand.Read(refreshToken)
	if err != nil {
		log.Fatal("Error getting random number")
	}

	if err != nil {
		log.Fatal(err)
	}

	ret := struct {
		ID		int    `json:"id"`
		Email		string `json:"email"`
		Token		string `json:"token"`
	}{
		ID: user.ID,
		Email: user.Email,
		Token: signedJWT,
	}



	respondWithJSON(w, 200, ret)
	
}

func (cfg *apiConfig) HandleUserUpdate(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, 401, "authorization header missing")
	}

	tokenParts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		respondWithError(w, 401, "invalid authorization header format")
	}
	tokenString := tokenParts[1]

	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func (token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		respondWithError(w, 401, "invalid token")
		return
	}

	userID, err := strconv.Atoi(claims.Subject)

	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "bad request")
		return
	}

	user, err := cfg.db.UpdateUser(userID, params.Email, params.Password)
	if err != nil {
		respondWithError(w, 500, "could not update credentials")
	}
	respondWithJSON(w, 200, user)
}

func (cfg *apiConfig) HandleUserCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "bad request")
		return
	}

	user, err := cfg.db.CreateUser(params.Email, params.Password)
	if err != nil {
		respondWithError(w, 500, "Error creating user.")
		return
	}
	respondWithJSON(w, 201, user)
}

func (cfg *apiConfig) HandleUserList(w http.ResponseWriter, r *http.Request) {
	users, err := cfg.db.GetUsers()
	if err != nil {
		respondWithError(w, 500, "error loading database")
		return
	}

	respondWithJSON(w, 200, users)
}
