package main

import (
	"cmp"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func health(w http.ResponseWriter, r *http.Request) {
	reply(w, http.StatusOK, "Server up and running.")
}

func main() {
	loadEnv()

	secret = os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("WEBHOOK_SECRET environment variable is required")
	}

	configPath := cmp.Or(os.Getenv("CONFIG_PATH"), "config.yaml")
	if err := loadConfig(configPath); err != nil {
		log.Fatalf("failed to load config %s: %v", configPath, err)
	}

	log.Printf("Loaded %d project(s) from %s", len(config.Projects), configPath)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", health)
	mux.HandleFunc("POST /webhook", webhook)

	port := cmp.Or(config.Server.Port, 8080)
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server starting on :%d", port)
	log.Fatal(srv.ListenAndServe())
}
