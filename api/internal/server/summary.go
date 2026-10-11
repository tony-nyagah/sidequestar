package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/tony-nyagah/sidequestar/api/internal/store"
	"github.com/tony-nyagah/sidequestar/api/internal/xp"
)

// How much quest history the summary prompt sees; enough to spot patterns,
// small enough for a 3B model's context.
const summaryHistory = 25

// refreshSummary asks the model to rewrite what it has learned about the person
// from their own words and their quests, keeping any corrections they made.
func (s *Server) refreshSummary(ctx context.Context, userID string) error {
	q := s.store.Queries()
	p, err := q.GetProfile(ctx, userID)
	if err != nil {
		return err
	}
	quests, err := q.ListQuests(ctx, userID)
	if err != nil {
		return err
	}

	release, err := s.limits.useModel(ctx)
	if err != nil {
		return err
	}
	summary, err := s.gen.Complete(ctx, buildSummaryPrompt(p, quests))
	release()
	if err != nil {
		return err
	}
	summary = strings.Trim(strings.TrimSpace(summary), `"`)
	if summary == "" {
		return fmt.Errorf("model returned an empty summary")
	}
	if len(summary) > maxSummaryLength {
		summary = summary[:maxSummaryLength]
	}
	return q.SaveGeneratedSummary(ctx, store.SaveGeneratedSummaryParams{
		UserID:           userID,
		Summary:          summary,
		SummaryUpdatedAt: toNullString(time.Now().UTC().Format(time.RFC3339)),
	})
}

// refreshSummaryInBackground updates the summary after a completion without making
// the person wait for the model. Failures only cost a stale summary, so they are logged.
func (s *Server) refreshSummaryInBackground(userID string) {
	if s.gen == nil {
		return
	}
	s.background.Add(1)
	go func() {
		defer s.background.Done()
		ctx, cancel := context.WithTimeout(context.Background(), generateTimeout)
		defer cancel()
		if err := s.refreshSummary(ctx, userID); err != nil {
			slog.Warn("background summary refresh", "user", userID, "err", err)
		}
	}()
}

// Wait blocks until background work (summary refreshes) has finished.
func (s *Server) Wait() { s.background.Wait() }

func buildSummaryPrompt(p store.Profile, quests []store.Quest) string {
	var b strings.Builder
	b.WriteString("You keep a short, private profile of one person for Sidequestar, an app that suggests small real-life activities away from work and school.\n")
	b.WriteString("Write 1 to 3 plain sentences about them in the third person: what they enjoy, when and for how long they are free, and anything that should shape future suggestions.\n")
	b.WriteString("Rules: only state what the information below directly supports, and say less rather than guess. ")
	b.WriteString("Never claim they finished, skipped, or dislike something unless it is listed below as completed or they said so. Never invent times, places, or numbers. ")
	b.WriteString("Do not guess about health, family, money, religion, or other personal matters they haven't mentioned. No lists, no headings, no quotation marks.\n\n")

	if p.Note != "" {
		fmt.Fprintf(&b, "In their own words (always true, never contradict this): %s\n", p.Note)
	}
	if p.Location != "" {
		fmt.Fprintf(&b, "Location: %s\n", p.Location)
	}
	if p.SummaryUserEdit != "" {
		fmt.Fprintf(&b, "They wrote this about themselves when correcting an earlier profile (keep it unless their quests clearly show it changed): %s\n", p.SummaryUserEdit)
	}

	var done, open []string
	for i, q := range quests {
		if i >= summaryHistory {
			break
		}
		line := fmt.Sprintf("- %s (%s; %s)", q.Title, durationText(q.DurationBucket), strings.Join(store.ParseTags(q.Tags), ", "))
		switch q.Status {
		case "completed":
			if q.PhotoPath.Valid {
				line += ", with a photo"
			}
			done = append(done, line)
		case "active":
			open = append(open, line)
		}
	}
	if len(done) > 0 {
		b.WriteString("\nQuests they actually completed, newest first:\n" + strings.Join(done, "\n") + "\n")
	} else {
		b.WriteString("\nThey have not completed any quests yet.\n")
	}
	if len(open) > 0 {
		b.WriteString("\nOn their to-do list (many are starter suggestions; this is NOT evidence of what they like or do):\n" + strings.Join(open, "\n") + "\n")
	}
	b.WriteString("\nWrite only the profile text.")
	return b.String()
}

func durationText(bucket string) string {
	switch xp.Bucket(bucket) {
	case xp.Under30m:
		return "under 30 minutes"
	case xp.M30to60:
		return "30 to 60 minutes"
	case xp.OneTo2h:
		return "1 to 2 hours"
	case xp.HalfDay:
		return "half a day"
	case xp.FullDay:
		return "a full day"
	}
	return bucket
}
