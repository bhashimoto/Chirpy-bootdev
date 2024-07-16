package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func ValidateJWT(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")	
	}

	tokenParts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		log.Println("Error in Split at validateToken")
		return "", errors.New("invalid authorization header format")	
	}
	tokenString := tokenParts[1]

	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func (token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil  {
		log.Println("Error in ParseWithClaims at validateToken")
		return "", err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	return userIDString, nil
}


func (cfg *apiConfig) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	tokenParts := strings.Split(r.Header.Get("Authorization"), " ")
	tokenString := tokenParts[1]

	err := cfg.db.RevokeToken(tokenString)
	if err != nil {
		respondWithError(w, 500, "internal error")
		return
	}
	respondWithJSON(w, 204, "")
}

func (cfg * apiConfig) HandleRefreshJWT(w http.ResponseWriter, r *http.Request) {
	tokenParts := strings.Split(r.Header.Get("Authorization"), " ")
	tokenString := tokenParts[1]

	refreshToken, err := cfg.db.GetRefreshToken(tokenString)
	if err != nil {
		respondWithError(w, 401, "invalid token")
		return
	}

	newJWT, err := cfg.generateJWT((*refreshToken).UserID, 60*60)
	if err != nil {
		respondWithError(w, 500, "problem generating token")
		return
	}

	type JWT struct {
		NewToken string `json:"token"`
	}

	payload := JWT{
		NewToken: newJWT,
	}
	respondWithJSON(w, 200, payload)




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

	signedJWT, err := cfg.generateJWT(user.ID, params.ExpiresInSeconds)
	if err != nil {
		respondWithError(w, 500, err.Error())
	}

	refreshToken, err := cfg.db.CreateRefreshToken(user.ID)
	if err != nil {
		respondWithError(w, 500, "error creating refres token")
		log.Fatal(err)
	}

	ret := struct {
		ID		int    `json:"id"`
		Email		string `json:"email"`
		IsChirpyRed	bool   `json:"is_chirpy_red"`
		Token		string `json:"token"`
		RefreshToken	string `json:"refresh_token"`
	}{
		ID: user.ID,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
		Token: signedJWT,
		RefreshToken: refreshToken.Token,
	}

	respondWithJSON(w, 200, ret)
	
}

func (cfg *apiConfig) generateJWT(userID int, expiresInSeconds int) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, 
		jwt.RegisteredClaims{
			Issuer: "chirpy",
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Second*time.Duration(expiresInSeconds))),
			Subject: strconv.Itoa(userID),
		},
	)


	secretKey := []byte(os.Getenv("JWT_SECRET"))
	signedJWT, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return signedJWT, nil
}

func (cfg *apiConfig) HandleUserUpdate(w http.ResponseWriter, r *http.Request) {
	userIDString, err := ValidateJWT(r)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

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
	userID, _ := strconv.Atoi(userIDString)
	user, err := cfg.db.UpdateUser(userID, params.Email, params.Password)
	if err != nil {
		respondWithError(w, 500, "could not update credentials")
		return
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
