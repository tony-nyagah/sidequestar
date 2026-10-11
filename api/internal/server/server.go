package server

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tony-nyagah/sidequestar/api/internal/ollama"
	"github.com/tony-nyagah/sidequestar/api/internal/store"
)

type Server struct {
	store      *store.Store
	photosDir  string
	gen        ollama.Generator
	background sync.WaitGroup
	limits     *limits
}

func New(st *store.Store, photosDir string) *Server {
	if err := os.MkdirAll(photosDir, 0o755); err != nil {
		panic("create photos dir: " + err.Error())
	}
	return &Server{store: st, photosDir: photosDir, limits: newLimits()}
}

func (s *Server) SetGenerator(g ollama.Generator) {
	s.gen = g
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("POST /api/auth/register", s.handleRegister)
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", s.handleLogout)

	// Everything below belongs to the signed-in user.
	mux.HandleFunc("GET /api/me", s.requireUser(s.handleMe))
	mux.HandleFunc("POST /api/me/password", s.requireUser(s.handleChangePassword))
	mux.HandleFunc("DELETE /api/me", s.requireUser(s.handleDeleteAccount))
	mux.HandleFunc("GET /api/quests", s.requireUser(s.handleListQuests))
	mux.HandleFunc("POST /api/quests", s.requireUser(s.handleCreateQuest))
	mux.HandleFunc("POST /api/quests/generate", s.requireUser(s.handleGenerate))
	mux.HandleFunc("POST /api/quests/{id}/accept", s.requireUser(s.handleAcceptQuest))
	mux.HandleFunc("POST /api/quests/{id}/complete", s.requireUser(s.handleCompleteQuest))
	mux.HandleFunc("DELETE /api/quests/{id}", s.requireUser(s.handleDeleteQuest))
	mux.HandleFunc("GET /api/photos/{name}", s.requireUser(s.handlePhoto))
	mux.HandleFunc("GET /api/profile", s.requireUser(s.handleGetProfile))
	mux.HandleFunc("PATCH /api/profile", s.requireUser(s.handleUpdateProfile))
	mux.HandleFunc("POST /api/profile/summary", s.requireUser(s.handleRefreshSummary))

	return logRequests(mux)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "dur", time.Since(start).Round(time.Millisecond))
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handlePhoto(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name != filepath.Base(name) || !allowedPhotoExts[strings.ToLower(filepath.Ext(name))] {
		writeError(w, http.StatusNotFound, "photo not found")
		return
	}
	owned, err := s.store.Queries().PhotoBelongsTo(r.Context(), store.PhotoBelongsToParams{
		PhotoPath: toNullString(name),
		UserID:    currentUser(r).ID,
	})
	if err != nil || owned == 0 {
		writeError(w, http.StatusNotFound, "photo not found")
		return
	}
	http.ServeFile(w, r, s.photoPathFor(name))
}
