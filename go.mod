module github.com/bhashimoto/Chirpy-bootdev

go 1.22.5

replace (
	internal/db v1.0.0 => ./internal/db/
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	internal/db v1.0.0
)

