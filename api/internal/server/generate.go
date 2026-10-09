package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tony-nyagah/sidequestar/api/internal/ollama"
	"github.com/tony-nyagah/sidequestar/api/internal/store"
	"github.com/tony-nyagah/sidequestar/api/internal/xp"
)

// generateTimeout covers a cold model load plus one answer on modest hardware.
const generateTimeout = 2 * time.Minute

type generateRequest struct {
	Context string `json:"context"`
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if s.gen == nil {
		writeError(w, http.StatusServiceUnavailable, "no model available; start Ollama")
		return
	}
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Context) == "" {
		writeError(w, http.StatusBadRequest, "context is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), generateTimeout)
	defer cancel()
	idea, err := s.gen.Generate(ctx, s.buildPrompt(ctx, strings.TrimSpace(req.Context)))
	if err != nil {
		slog.Error("generate quest", "err", err)
		switch {
		case errors.Is(err, ollama.ErrUnreachable):
			writeError(w, http.StatusServiceUnavailable, "Ollama isn't running. Start it with `ollama serve`, then try again.")
		case errors.Is(err, ollama.ErrModelMissing):
			writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("The model isn't downloaded yet. Run `ollama pull %s`, then try again.", s.gen.Model()))
		case errors.Is(err, context.DeadlineExceeded):
			writeError(w, http.StatusGatewayTimeout, "The quest-giver took too long to answer. Try again.")
		default:
			writeError(w, http.StatusBadGateway, "The quest-giver couldn't come up with a quest. Try again.")
		}
		return
	}

	if !xp.ValidBucket(idea.DurationBucket) {
		idea.DurationBucket = string(xp.Under30m)
	}

	q, err := s.store.Queries().CreateQuest(r.Context(), store.CreateQuestParams{
		ID:             uuid.NewString(),
		Title:          strings.TrimSpace(idea.Title),
		Description:    strings.TrimSpace(idea.Description),
		DurationBucket: idea.DurationBucket,
		Source:         "generated",
		Status:         "generated",
		Tags:           store.MarshalTags(idea.Tags),
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		slog.Error("save generated quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save quest")
		return
	}
	writeJSON(w, http.StatusCreated, toQuestDTO(q))
}

func (s *Server) buildPrompt(ctx context.Context, situation string) string {
	p, _ := s.store.Queries().GetProfile(ctx)
	interests := s.interests(ctx)

	var b strings.Builder
	b.WriteString("You are Sidequestar, a quest-giver that nudges one person to do a single real-life thing away from work and school.\n")
	b.WriteString("Invent ONE concrete, safe, offline sidequest they can actually start today. Be specific and actionable, not generic.\n")
	if p.Location != "" {
		fmt.Fprintf(&b, "Their location: %s.\n", p.Location)
	}
	if p.Note != "" {
		fmt.Fprintf(&b, "About them: %s.\n", p.Note)
	}
	if len(interests) > 0 {
		fmt.Fprintf(&b, "Things they seem to enjoy: %s.\n", strings.Join(interests, ", "))
	}
	fmt.Fprintf(&b, "Their situation right now: %s.\n\n", situation)
	b.WriteString("Respond with ONLY a JSON object, no markdown fences, matching exactly this shape:\n")
	b.WriteString(`{"title": string, "description": string, "duration_bucket": one of "under-30m","30-60m","1-2h","half-day","full-day", "tags": [2-4 short lowercase words]}`)
	return b.String()
}
