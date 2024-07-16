package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

func validatePolkaKeys(r *http.Request) error {
	theirPolkaKey, err := getAPIKey(r)
	if err != nil {
		return err
	}

	ourPolkaKey := os.Getenv("POLKA_KEY")
	if theirPolkaKey != ourPolkaKey {
		return errors.New("invalid key")
	}
	return nil
}

func (cfg *apiConfig) HandlePolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	err := validatePolkaKeys(r)
	if err != nil {
		respondWithError(w, 401, "invalid auth token")
		return
	}


	type parameters struct {
		Event string `json:"event"`
		Data struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 204, "invalid request")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithError(w, 204, "invalid event")
		return
	}

	err = cfg.db.UpgradeUser(params.Data.UserID)
	if err != nil {
		
		if err.Error() == "user not found" {
			respondWithError(w, 404, "user not found")
			return
		}
		respondWithError(w, 500, "error upgrading user")
		return
	}
	respondWithJSON(w, 204, "")


}
