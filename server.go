package main

import (
	"log"
	"net/http"
	"os"
)

func server() {
	mux := http.NewServeMux()

	registerRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
