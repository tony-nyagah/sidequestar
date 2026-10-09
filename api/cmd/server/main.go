package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tony-nyagah/sidequestar/api/internal/ollama"
	"github.com/tony-nyagah/sidequestar/api/internal/server"
	"github.com/tony-nyagah/sidequestar/api/internal/store"
)

func main() {
	port := getenv("PORT", "8080")
	dataDir := getenv("DATA_DIR", "data")
	ollamaModel := os.Getenv("OLLAMA_MODEL")

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		slog.Error("create data dir", "err", err)
		os.Exit(1)
	}

	st, err := store.Open(filepath.Join(dataDir, "sidequestar.db"))
	if err != nil {
		slog.Error("open store", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	if err := st.SeedIfEmpty(context.Background()); err != nil {
		slog.Error("seed quests", "err", err)
		os.Exit(1)
	}

	srv := server.New(st, filepath.Join(dataDir, "photos"))

	gen, err := ollama.New(ollamaModel)
	if err != nil {
		slog.Error("configure ollama client", "err", err)
		os.Exit(1)
	}
	srv.SetGenerator(gen)
	slog.Info("generation configured", "model", gen.Model())

	addr := ":" + port
	slog.Info("sidequestar api listening", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
