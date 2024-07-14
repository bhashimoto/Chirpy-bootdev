package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

const path string = "./database.json"

type apiConfig struct {
	fileserverHits int
	db *DB
}


func main() {
	godotenv.Load()
	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr: "localhost:8080",
	}
	
	db, err := NewDB(path)
	if err != nil {
		log.Fatal(err)
	}
	apiCfg := apiConfig{
		fileserverHits: 0,
		db: db,	
	}

	serveMux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", apiCfg.HandleHealthz)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.HandleFileServerHits )
	serveMux.HandleFunc("/api/reset", apiCfg.HandleResetFileServerHits)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.HandleCreateChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.HandleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.HandleCreateChirp)
	serveMux.HandleFunc("POST /api/users", apiCfg.HandleUserCreate)
	serveMux.HandleFunc("PUT /api/users", apiCfg.HandleUserUpdate)
	serveMux.HandleFunc("GET /api/users", apiCfg.HandleUserList)
	serveMux.HandleFunc("POST /api/login", apiCfg.HandleUserLogin)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.HandleRefresh)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(fmt.Sprintf("{%s}", message))
	if err != nil {
		log.Printf("error marshaling error message: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
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
	chirps, err := cfg.db.GetChirps()
	if err != nil {
		log.Println("error loading chirps")
		respondWithError(w, 500, "error loading chirps")
	}
	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Chirp string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "Something went wrong")
		return
	}
	

	if len(params.Chirp) > 140 {
		respondWithError(w, 500, "Chirp is too long")
		return
	}

	dbStructure, err := cfg.db.loadDB()
	if err != nil {
		log.Printf("error loading db")
	}
	chirp, err := cfg.db.CreateChirp(cleanChirp(params.Chirp))
	if err != nil {
		log.Println("error creating chirp")
		return
	}
	dbStructure.Chirps[chirp.ID] = chirp

	err = cfg.db.writeDB(dbStructure)
	if err != nil {
		log.Println("error writing database")
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


func (cfg *apiConfig) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	
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

func createDB() (DB) {
	return DB{
		path: path,
		mux: &sync.RWMutex{},
	}
}
