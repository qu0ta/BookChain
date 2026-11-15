package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/qu0ta/BookChain/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)

	addr := ":" + cfg.HTTPPort

	log.Printf("BookChain API started on %s\n", addr)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
