package main

import (
	"log"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr: "localhost:8080",
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

