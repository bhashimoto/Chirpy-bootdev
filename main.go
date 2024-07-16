package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.HandleChirpDelete)
	serveMux.HandleFunc("POST /api/users", apiCfg.HandleUserCreate)
	serveMux.HandleFunc("PUT /api/users", apiCfg.HandleUserUpdate)
	serveMux.HandleFunc("GET /api/users", apiCfg.HandleUserList)
	serveMux.HandleFunc("POST /api/login", apiCfg.HandleUserLogin)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.HandleRefreshJWT)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.HandleRevokeToken)
	serveMux.HandleFunc("POST /api/polka/webhooks", apiCfg.HandlePolkaWebhooks)

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

func getAPIKey(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")	
	}

	tokenParts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(tokenParts) != 2 || tokenParts[0] != "ApiKey" {
		return "", errors.New("invalid authorization header format")	
	}
	tokenString := tokenParts[1]

	return tokenString, nil
}

func getAuthToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")	
	}

	tokenParts := strings.Split(r.Header.Get("Authorization"), " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")	
	}
	tokenString := tokenParts[1]

	return tokenString, nil
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
