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
	{"Watch the sunset somewhere new", "Find a spot with a clear view west and watch the sun go all the way down.", "30-60m", []string{"outdoors", "calm", "sky"}},
	{"Identify three trees", "Walk around your block and learn the names of three trees you pass every day.", "30-60m", []string{"outdoors", "nature", "learning"}},
	{"Eat a meal outside", "Take lunch or dinner to a park, balcony or bench instead of eating at a screen.", "30-60m", []string{"outdoors", "food", "calm"}},
	{"Visit a local green space", "Go to a park or garden and just sit and notice for a while.", "30-60m", []string{"outdoors", "nature", "calm"}},
	{"Photograph five colours", "Go outside and photograph five different colours you find in nature.", "30-60m", []string{"outdoors", "creative", "photography"}},
	{"Walk with a friend", "Invite someone you miss for a walk instead of a call.", "1-2h", []string{"outdoors", "social", "walking"}},
	{"Pick up litter in your street", "Take a bag and gloves and leave your street cleaner than you found it.", "30-60m", []string{"outdoors", "community", "giving"}},
	{"Hike a nearby trail", "Find the closest trail or nature reserve and walk it end to end.", "half-day", []string{"outdoors", "hiking", "explore"}},
	{"Plan a day trip outdoors", "Pick a lake, hill or forest you've never been to and go spend the day there.", "full-day", []string{"outdoors", "explore", "adventure"}},
}

func (s *Store) SeedIfEmpty(ctx context.Context) error {
	count, err := s.q.CountQuests(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now()
	for i, sq := range seeds {
		_, err := s.q.CreateQuest(ctx, CreateQuestParams{
			ID:             uuid.NewString(),
			Title:          sq.Title,
			Description:    sq.Description,
			DurationBucket: sq.DurationBucket,
			Source:         "seed",
			Status:         "active",
			Tags:           MarshalTags(sq.Tags),
			// Stagger timestamps so the list keeps the order above (newest first).
			CreatedAt: Timestamp(now.Add(-time.Duration(i) * time.Millisecond)),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
