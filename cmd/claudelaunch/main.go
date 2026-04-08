package main

import (
	"log/slog"
	"net/http"
	"os"

	"maragu.dev/claudelaunch"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	port := os.Getenv("PORT")
	if port == "" {
		port = "6677"
	}

	s := &claudelaunch.Server{
		Log: log,
	}

	log.Info("Starting server", "port", port)

	if err := http.ListenAndServe(":"+port, s.Handler()); err != nil {
		log.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
