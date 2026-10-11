package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/tony-nyagah/sidequestar/api/internal/store"
	"github.com/tony-nyagah/sidequestar/api/internal/xp"
)

const maxSummaryLength = 1200

type profileDTO struct {
	Username         string   `json:"username"`
	Location         string   `json:"location"`
	Note             string   `json:"note"`
	Summary          string   `json:"summary"`
	SummaryUpdatedAt string   `json:"summary_updated_at,omitempty"`
	Interests        []string `json:"interests"`
	XpTotal          int      `json:"xp_total"`
	Completed        int64    `json:"completed_count"`
	Level            int      `json:"level"`
	XpIntoLevel      int      `json:"xp_into_level"`
	XpNeeded         int      `json:"xp_needed"`
	Title            string   `json:"title"`
	KnownFacts       []string `json:"known_facts"`
	Model            string   `json:"model"`
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	dto, err := s.profileFor(r.Context(), currentUser(r))
	if err != nil {
		slog.Error("get profile", "err", err)
		writeError(w, http.StatusInternalServerError, "could not read profile")
		return
	}
	writeJSON(w, http.StatusOK, dto)
}

func (s *Server) profileFor(ctx context.Context, user store.User) (profileDTO, error) {
	q := s.store.Queries()
	p, err := q.GetProfile(ctx, user.ID)
	if err != nil {
		return profileDTO{}, err
	}
	xpTotal, _ := q.TotalCompletedXp(ctx, user.ID)
	completed, _ := q.CountCompleted(ctx, user.ID)
	level := xp.LevelForXp(int(xpTotal))

	return profileDTO{
		Username:         user.Username,
		Location:         p.Location,
		Note:             p.Note,
		Summary:          p.Summary,
		SummaryUpdatedAt: p.SummaryUpdatedAt.String,
		Interests:        s.interests(ctx, user.ID),
		XpTotal:          int(xpTotal),
		Completed:        completed,
		Level:            level.Level,
		XpIntoLevel:      level.Into,
		XpNeeded:         level.Needed,
		Title:            xp.Title(level.Level),
		KnownFacts:       s.knownFacts(ctx, user.ID),
		Model:            s.modelName(),
	}, nil
}

// Omitted fields are left as they are, so the about form and the summary form
// can save independently.
type updateProfileRequest struct {
	Location *string `json:"location"`
	Note     *string `json:"note"`
	Summary  *string `json:"summary"`
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	user := currentUser(r)
	q := s.store.Queries()
	p, err := q.GetProfile(r.Context(), user.ID)
	if err != nil {
		slog.Error("get profile", "err", err)
		writeError(w, http.StatusInternalServerError, "could not update profile")
		return
	}

	if req.Location != nil || req.Note != nil {
		location, note := p.Location, p.Note
		if req.Location != nil {
			location = strings.TrimSpace(*req.Location)
		}
		if req.Note != nil {
			note = strings.TrimSpace(*req.Note)
		}
		if _, err := q.UpdateProfile(r.Context(), store.UpdateProfileParams{UserID: user.ID, Location: location, Note: note}); err != nil {
			slog.Error("update profile", "err", err)
			writeError(w, http.StatusInternalServerError, "could not update profile")
			return
		}
	}
	if req.Summary != nil {
		summary := strings.TrimSpace(*req.Summary)
		if len(summary) > maxSummaryLength {
			writeError(w, http.StatusBadRequest, "Keep the summary under 1200 characters.")
			return
		}
		err := s.store.Queries().SaveEditedSummary(r.Context(), store.SaveEditedSummaryParams{
			UserID:           user.ID,
			Summary:          summary,
			SummaryUpdatedAt: toNullString(time.Now().UTC().Format(time.RFC3339)),
		})
		if err != nil {
			slog.Error("update summary", "err", err)
			writeError(w, http.StatusInternalServerError, "could not update summary")
			return
		}
	}

	s.handleGetProfile(w, r)
}

// handleRefreshSummary asks the model to rewrite the summary now and waits for it.
func (s *Server) handleRefreshSummary(w http.ResponseWriter, r *http.Request) {
	if s.gen == nil {
		writeError(w, http.StatusServiceUnavailable, "no model available; start Ollama")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), generateTimeout)
	defer cancel()
	if err := s.refreshSummary(ctx, currentUser(r).ID); err != nil {
		slog.Error("refresh summary", "err", err)
		s.writeModelError(w, err, "The quest-giver couldn't write a summary. Try again.")
		return
	}
	s.handleGetProfile(w, r)
}

func (s *Server) modelName() string {
	if s.gen == nil {
		return ""
	}
	return s.gen.Model()
}

func (s *Server) interests(ctx context.Context, userID string) []string {
	quests, err := s.store.Queries().ListQuests(ctx, userID)
	if err != nil {
		return []string{}
	}

	counts := map[string]int{}
	for _, q := range quests {
		for _, t := range store.ParseTags(q.Tags) {
			counts[t]++
		}
	}

	type entry struct {
		name  string
		count int
	}
	list := make([]entry, 0, len(counts))
	for name, count := range counts {
		list = append(list, entry{name, count})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].count != list[j].count {
			return list[i].count > list[j].count
		}
		return list[i].name < list[j].name
	})

	const maxInterests = 8
	out := make([]string, 0, maxInterests)
	for i, e := range list {
		if i >= maxInterests {
			break
		}
		out = append(out, e.name)
	}
	return out
}
