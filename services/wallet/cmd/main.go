package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("wallet-service is healthy"))
	})

	fmt.Printf("Wallet Service starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
