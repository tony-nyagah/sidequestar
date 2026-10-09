package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tony-nyagah/sidequestar/api/internal/store"
	"github.com/tony-nyagah/sidequestar/api/internal/xp"
)

const (
	generateTimeout  = 90 * time.Second
	recentQuestLimit = 10
)

type generateRequest struct {
	Context string `json:"context"`
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if s.gen == nil {
		writeError(w, http.StatusServiceUnavailable, "no model available; start Ollama")
		return
	}
	var req generateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Context) == "" {
		writeError(w, http.StatusBadRequest, "context is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), generateTimeout)
	defer cancel()

	prompt := s.buildPrompt(ctx, strings.TrimSpace(req.Context), time.Now())
	idea, err := s.gen.Generate(ctx, prompt)
	if err != nil {
		slog.Error("generate quest", "err", err)
		writeError(w, http.StatusBadGateway, "could not generate a quest; is Ollama running?")
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
		Tags:           store.MarshalTags(normalizeTags(idea.Tags)),
		CreatedAt:      store.Timestamp(time.Now()),
	})
	if err != nil {
		slog.Error("save generated quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not save quest")
		return
	}
	writeJSON(w, http.StatusCreated, toQuestDTO(q))
}

func (s *Server) buildPrompt(ctx context.Context, situation string, now time.Time) string {
	p, _ := s.store.Queries().GetProfile(ctx)
	interests := s.interests(ctx)
	recent := s.recentTitles(ctx)

	var b strings.Builder
	b.WriteString("You are Sidequestar, a quest-giver that gets one person off their screen and out into the world.\n")
	b.WriteString("Invent ONE concrete, safe sidequest they can start today. Strongly prefer something outdoors: walking, nature, parks, trails, gardening, birding, exploring their neighbourhood. Only suggest an indoor quest if their situation clearly rules out going outside.\n")
	b.WriteString("Be specific and actionable, not generic. The screen should be the shortest part of the experience.\n")
	fmt.Fprintf(&b, "It is currently %s. Make sure the quest suits the time of day (no hikes after dark).\n", now.Format("Monday 2 January 2006, 15:04"))
	if p.Location != "" {
		fmt.Fprintf(&b, "Their location: %s. Factor in the likely season and climate there.\n", p.Location)
	}
	if p.Note != "" {
		fmt.Fprintf(&b, "About them: %s.\n", p.Note)
	}
	if len(interests) > 0 {
		fmt.Fprintf(&b, "Things they have enjoyed doing: %s.\n", strings.Join(interests, ", "))
	}
	if len(recent) > 0 {
		fmt.Fprintf(&b, "Do NOT repeat or closely copy any of these recent quests: %s.\n", strings.Join(recent, "; "))
	}
	fmt.Fprintf(&b, "Their situation right now: %s.\n\n", situation)
	b.WriteString("Respond with a JSON object: a short title, a one or two sentence description, a duration_bucket, and 2-4 short lowercase tags.")
	return b.String()
}

func (s *Server) recentTitles(ctx context.Context) []string {
	quests, err := s.store.Queries().ListQuests(ctx)
	if err != nil {
		return nil
	}
	out := make([]string, 0, recentQuestLimit)
	for _, q := range quests {
		if len(out) == recentQuestLimit {
			break
		}
		out = append(out, q.Title)
	}
	return out
}
