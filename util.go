package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func loadEnv() {
	const path = ".env"
	if _, err := os.Stat(path); err != nil {
		return
	}

	if err := godotenv.Load(path); err != nil {
		log.Printf("note: could not load %s: %v", path, err)
	} else {
		log.Printf("Loaded env from %s.", path)
	}
}

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}

	return path
}

func reply(w http.ResponseWriter, code int, format string, a ...any) {
	w.WriteHeader(code)
	fmt.Fprintf(w, format, a...)
}
