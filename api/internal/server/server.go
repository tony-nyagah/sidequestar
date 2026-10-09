package server

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/tony-nyagah/sidequestar/api/internal/ollama"
	"github.com/tony-nyagah/sidequestar/api/internal/store"
)

type Server struct {
	store     *store.Store
	photosDir string
	gen       ollama.Generator
}

func New(st *store.Store, photosDir string) *Server {
	if err := os.MkdirAll(photosDir, 0o755); err != nil {
		panic("create photos dir: " + err.Error())
	}
	return &Server{store: st, photosDir: photosDir}
}

func (s *Server) SetGenerator(g ollama.Generator) {
	s.gen = g
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/quests", s.handleListQuests)
	mux.HandleFunc("POST /api/quests", s.handleCreateQuest)
	mux.HandleFunc("POST /api/quests/generate", s.handleGenerate)
	mux.HandleFunc("POST /api/quests/{id}/accept", s.handleAcceptQuest)
	mux.HandleFunc("POST /api/quests/{id}/complete", s.handleCompleteQuest)
	mux.HandleFunc("DELETE /api/quests/{id}", s.handleDeleteQuest)
	mux.HandleFunc("GET /api/profile", s.handleGetProfile)
	mux.HandleFunc("PATCH /api/profile", s.handleUpdateProfile)

	return logRequests(mux)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func absPhotoPath(photosDir, name string) string {
	return filepath.Join(photosDir, name)
}
