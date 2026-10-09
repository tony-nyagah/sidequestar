package server

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/tony-nyagah/sidequestar/api/internal/store"
	"github.com/tony-nyagah/sidequestar/api/internal/xp"
)

type profileDTO struct {
	Location    string   `json:"location"`
	Note        string   `json:"note"`
	Interests   []string `json:"interests"`
	XpTotal     int      `json:"xp_total"`
	Completed   int64    `json:"completed_count"`
	Level       int      `json:"level"`
	XpIntoLevel int      `json:"xp_into_level"`
	XpNeeded    int      `json:"xp_needed"`
	Title       string   `json:"title"`
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	p, err := s.store.Queries().GetProfile(r.Context())
	if err != nil {
		slog.Error("get profile", "err", err)
		writeError(w, http.StatusInternalServerError, "could not read profile")
		return
	}
	xpTotal, _ := s.store.Queries().TotalCompletedXp(r.Context())
	completed, _ := s.store.Queries().CountCompleted(r.Context())
	level := xp.LevelForXp(int(xpTotal))

	writeJSON(w, http.StatusOK, profileDTO{
		Location:    p.Location,
		Note:        p.Note,
		Interests:   s.interests(r.Context()),
		XpTotal:     int(xpTotal),
		Completed:   completed,
		Level:       level.Level,
		XpIntoLevel: level.Into,
		XpNeeded:    level.Needed,
		Title:       xp.Title(level.Level),
	})
}

type updateProfileRequest struct {
	Location string `json:"location"`
	Note     string `json:"note"`
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req updateProfileRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	p, err := s.store.Queries().UpdateProfile(r.Context(), store.UpdateProfileParams{
		Location: strings.TrimSpace(req.Location),
		Note:     strings.TrimSpace(req.Note),
	})
	if err != nil {
		slog.Error("update profile", "err", err)
		writeError(w, http.StatusInternalServerError, "could not update profile")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"location": p.Location, "note": p.Note})
}

// interests ranks tags from completed quests only. Counting generated or
// seeded quests would let the model's own tags feed back into its prompt.
func (s *Server) interests(ctx context.Context) []string {
	quests, err := s.store.Queries().ListQuestsByStatus(ctx, "completed")
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
