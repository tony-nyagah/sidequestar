package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tony-nyagah/sidequestar/api/internal/ollama"
	"github.com/tony-nyagah/sidequestar/api/internal/server"
	"github.com/tony-nyagah/sidequestar/api/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	host := getenv("HOST", "127.0.0.1")
	port := getenv("PORT", "8080")
	dataDir := getenv("DATA_DIR", "data")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}

	st, err := store.Open(filepath.Join(dataDir, "sidequestar.db"))
	if err != nil {
		return err
	}
	defer st.Close()

	if err := st.SeedIfEmpty(ctx); err != nil {
		return err
	}

	srv := server.New(st, filepath.Join(dataDir, "photos"))

	gen, err := ollama.New(os.Getenv("OLLAMA_MODEL"))
	if err != nil {
		slog.Warn("ollama client unavailable; generation disabled", "err", err)
	} else {
		srv.SetGenerator(gen)
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if err := gen.Ping(pingCtx); err != nil {
			slog.Warn("ollama not reachable yet; generation will retry per request", "err", err)
		} else {
			slog.Info("ollama connected", "model", gen.Model())
		}
		cancel()
	}

	httpSrv := &http.Server{
		Addr:              net.JoinHostPort(host, port),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		// Generation on a small local model can take a while.
		WriteTimeout: 2 * time.Minute,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("sidequestar api listening", "addr", httpSrv.Addr)
		errc <- httpSrv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
