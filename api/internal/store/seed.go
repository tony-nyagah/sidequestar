package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type seedQuest struct {
	Title          string
	Description    string
	DurationBucket string
	Tags           []string
}

var seeds = []seedQuest{
	{"Find a new walking route", "Walk a route you've never taken before, right from your front door.", "under-30m", []string{"outdoors", "walking", "explore"}},
	{"Cook a new meal from scratch", "Pick a recipe you've never tried and cook the whole thing yourself.", "1-2h", []string{"cooking", "creative"}},
	{"Read for 30 minutes", "Sit somewhere comfortable and read a book or a long article.", "under-30m", []string{"reading", "calm"}},
	{"Call a friend you miss", "Call or send a voice note to someone you haven't spoken to in a while.", "under-30m", []string{"social", "connection"}},
	{"Do a 20-minute stretch", "Stretch or do gentle movement for your whole body.", "under-30m", []string{"body", "health"}},
	{"Visit a local green space", "Go to a park or garden and just sit and notice for a while.", "30-60m", []string{"outdoors", "nature", "calm"}},
	{"Sketch or photograph something", "Capture one thing you find beautiful today, in any medium.", "30-60m", []string{"creative", "art"}},
	{"Deep clean one corner", "Tidy and clean one small space you've been ignoring.", "1-2h", []string{"home", "organize"}},
	{"Plan a day trip", "Research and plan a day out somewhere you've never been.", "1-2h", []string{"planning", "explore"}},
	{"Offer an hour of help", "Give an hour to someone nearby or a local cause.", "1-2h", []string{"giving", "community"}},
}

func (s *Store) SeedIfEmpty(ctx context.Context) error {
	count, err := s.q.CountQuests(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, sq := range seeds {
		_, err := s.q.CreateQuest(ctx, CreateQuestParams{
			ID:             uuid.NewString(),
			Title:          sq.Title,
			Description:    sq.Description,
			DurationBucket: sq.DurationBucket,
			Source:         "seed",
			Status:         "active",
			Tags:           MarshalTags(sq.Tags),
			CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
