package server

import (
	"github.com/tony-nyagah/sidequestar/api/internal/store"
)

type questDTO struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	DurationBucket string   `json:"duration_bucket"`
	Source         string   `json:"source"`
	Status         string   `json:"status"`
	Tags           []string `json:"tags"`
	PhotoPath      string   `json:"photo_path,omitempty"`
	XpAwarded      int      `json:"xp_awarded,omitempty"`
	CreatedAt      string   `json:"created_at"`
	CompletedAt    string   `json:"completed_at,omitempty"`
}

func toQuestDTO(q store.Quest) questDTO {
	dto := questDTO{
		ID:             q.ID,
		Title:          q.Title,
		Description:    q.Description,
		DurationBucket: q.DurationBucket,
		Source:         q.Source,
		Status:         q.Status,
		Tags:           store.ParseTags(q.Tags),
		CreatedAt:      q.CreatedAt,
	}
	if q.PhotoPath.Valid {
		dto.PhotoPath = q.PhotoPath.String
	}
	if q.XpAwarded.Valid {
		dto.XpAwarded = int(q.XpAwarded.Int64)
	}
	if q.CompletedAt.Valid {
		dto.CompletedAt = q.CompletedAt.String
	}
	return dto
}
