package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)


func main() {
	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr: "localhost:8080",
	}
	
	apiCfg := apiConfig{}

	serveMux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", HandleHealthz)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.HandleFileServerHits )
	serveMux.HandleFunc("/api/reset", apiCfg.HandleResetFileServerHits)
	serveMux.HandleFunc("POST /api/chirps", validateChirp)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

type returnVals struct {
	Error string `json:"error,omitempty"`
	Body string `json:"body,omitempty"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marhsaling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Chirp string `json:"body"`
	}
	returnCode := 0
	decoder := json.NewDecoder(r.Body)

	retVals := returnVals{}

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		retVals.Error = "Something went wrong"
		w.WriteHeader(500)
		return
	}
	

	if len(params.Chirp) > 140 {
		retVals.Error = "Chirp is too long"
		returnCode = 500
	} else {
		retVals.Body = cleanChirp(params.Chirp)
		returnCode = 200
	}

	respondWithJSON(w, returnCode, retVals)

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


func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	
}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		log.Println("increasing metric")
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) HandleFileServerHits (w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
		`, cfg.fileserverHits)
	w.Write([]byte(hits))
}

func (cfg *apiConfig) HandleResetFileServerHits(w http.ResponseWriter, r *http.Request) {
	log.Println("Resetting hits count")
	cfg.fileserverHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits have been reset"))
	
}
