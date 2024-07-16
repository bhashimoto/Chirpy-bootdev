package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (cfg *apiConfig) HandleChirpDelete(w http.ResponseWriter, r *http.Request) {
	userIDString, err := ValidateJWT(r)
	if err != nil {
		respondWithError(w, 403, "unauthorized")
		return
	}
	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal error")
	}

	chirpID, err := strconv.Atoi(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 403, "unauthorized")
		return
	}

	chirp, found := cfg.db.GetChirp(chirpID)
	if !found {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}
	if chirp.AuthorID != userID{
		respondWithError(w, 403, "unauthorized")
		return
	}
	err = cfg.db.DeleteChirp(chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error deleting chirp")
		return
	}
	respondWithJSON(w, 204, "")

}

func (cfg *apiConfig) HandleGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("error getting id: %v", err)
	}
	db := createDB()
	chirp, found := db.GetChirp(id)
	if !found {
		respondWithError(w, 404, "Chirp not found")
	}
	respondWithJSON(w, 200, chirp)
}

func (cfg *apiConfig) HandleGetChirps(w http.ResponseWriter, r *http.Request) {
	cfg.db.loadDB()

	userID := 0
	userIDString := r.URL.Query().Get("author_id")
	if userIDString != "" {
		userID, _ = strconv.Atoi(userIDString)
	}

	sortDirection := r.URL.Query().Get("sort")
	if sortDirection != "desc" {
		sortDirection = "asc"
	}

	chirps, err := cfg.db.GetChirps(userID, sortDirection)
	if err != nil {
		log.Println("error loading chirps")
		respondWithError(w, 500, "error loading chirps")
	}
	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	userIDString, err := ValidateJWT(r)
	if err != nil {
		respondWithError(w, 401, "user not authorized")
		return
	}

	type parameters struct {
		Chirp string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	

	if len(params.Chirp) > 140 {
		respondWithError(w, 500, "Chirp is too long")
		return
	}

	userID, _ := strconv.Atoi(userIDString)
	chirp, err := cfg.db.CreateChirp(cleanChirp(params.Chirp), userID)
	if err != nil {
		respondWithError(w, 500, "error creating chirp")
		return
	}

	respondWithJSON(w, 201, chirp)
}

func cleanChirp(chirp string) string {
	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert": true,
		"fornax": true,
	}
	words := strings.Split(chirp, " ")
	for i, word := range words {
		if badWords[strings.ToLower(word)] {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")

}
