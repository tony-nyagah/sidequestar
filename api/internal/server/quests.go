package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tony-nyagah/sidequestar/api/internal/store"
	"github.com/tony-nyagah/sidequestar/api/internal/xp"
)

const maxPhotoBytes = 20 << 20

var (
	errBadPhotoType  = errors.New("unsupported photo type")
	errTooManyPhotos = errors.New("photo upload limit reached")
	errPhotoQuota    = errors.New("photo storage quota reached")
)

var allowedPhotoExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true, ".heic": true,
}

func (s *Server) handleListQuests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	userID := currentUser(r).ID

	var list []store.Quest
	var err error
	if status == "" {
		list, err = s.store.Queries().ListQuests(r.Context(), userID)
	} else {
		list, err = s.store.Queries().ListQuestsByStatus(r.Context(), store.ListQuestsByStatusParams{UserID: userID, Status: status})
	}
	if err != nil {
		slog.Error("list quests", "err", err)
		writeError(w, http.StatusInternalServerError, "could not list quests")
		return
	}

	out := make([]questDTO, 0, len(list))
	for _, q := range list {
		out = append(out, toQuestDTO(q))
	}
	writeJSON(w, http.StatusOK, out)
}

type createQuestRequest struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	DurationBucket string   `json:"duration_bucket"`
	Tags           []string `json:"tags"`
}

func (s *Server) handleCreateQuest(w http.ResponseWriter, r *http.Request) {
	var req createQuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if strings.TrimSpace(req.Description) == "" {
		writeError(w, http.StatusBadRequest, "description is required")
		return
	}
	if !xp.ValidBucket(req.DurationBucket) {
		writeError(w, http.StatusBadRequest, "invalid duration_bucket")
		return
	}

	q, err := s.store.Queries().CreateQuest(r.Context(), store.CreateQuestParams{
		ID:             uuid.NewString(),
		UserID:         currentUser(r).ID,
		Title:          strings.TrimSpace(req.Title),
		Description:    strings.TrimSpace(req.Description),
		DurationBucket: req.DurationBucket,
		Source:         "hand",
		Status:         "active",
		Tags:           store.MarshalTags(req.Tags),
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		slog.Error("create quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not create quest")
		return
	}
	writeJSON(w, http.StatusCreated, toQuestDTO(q))
}

func (s *Server) handleAcceptQuest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	q, err := s.store.Queries().AcceptQuest(r.Context(), store.AcceptQuestParams{ID: id, UserID: currentUser(r).ID})
	if err != nil {
		if isNoRows(err) {
			writeError(w, http.StatusNotFound, "quest not found or not generated")
			return
		}
		slog.Error("accept quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not accept quest")
		return
	}
	writeJSON(w, http.StatusOK, toQuestDTO(q))
}

func (s *Server) handleCompleteQuest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := currentUser(r).ID
	q, err := s.store.Queries().GetQuest(r.Context(), store.GetQuestParams{ID: id, UserID: userID})
	if err != nil {
		if isNoRows(err) {
			writeError(w, http.StatusNotFound, "quest not found")
			return
		}
		slog.Error("get quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not read quest")
		return
	}
	if q.Status != "active" {
		writeError(w, http.StatusConflict, "quest is not active")
		return
	}

	var photoPath string
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, maxPhotoBytes)
		photoPath, err = s.savePhoto(r, userID)
		if err != nil {
			var tooBig *http.MaxBytesError
			switch {
			case errors.As(err, &tooBig):
				writeError(w, http.StatusRequestEntityTooLarge, "photo is too large (max 20 MB)")
			case errors.Is(err, errBadPhotoType):
				writeError(w, http.StatusBadRequest, "photo must be a JPEG, PNG, WebP, GIF, or HEIC image")
			case errors.Is(err, errTooManyPhotos):
				tooMany(w, "That's 20 photos in the last hour. Mark this one done without a photo, or try again later.")
			case errors.Is(err, errPhotoQuota):
				writeError(w, http.StatusRequestEntityTooLarge, "Your photos are using all 200 MB of your space. Mark this one done without a photo.")
			default:
				slog.Error("save photo", "err", err)
				writeError(w, http.StatusInternalServerError, "could not save photo")
			}
			return
		}
	}

	awarded := xp.Award(xp.Bucket(q.DurationBucket), photoPath != "")
	now := time.Now().UTC().Format(time.RFC3339)

	completed, err := s.store.Queries().CompleteQuest(r.Context(), store.CompleteQuestParams{
		ID:          id,
		UserID:      userID,
		CompletedAt: toNullString(now),
		PhotoPath:   toNullString(photoPath),
		XpAwarded:   toNullInt(int64(awarded)),
	})
	if err != nil {
		if photoPath != "" {
			_ = os.Remove(s.photoPathFor(photoPath))
		}
		slog.Error("complete quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not complete quest")
		return
	}
	s.refreshSummaryInBackground(userID)
	writeJSON(w, http.StatusOK, toQuestDTO(completed))
}

func (s *Server) handleDeleteQuest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.Queries().DeleteQuest(r.Context(), store.DeleteQuestParams{ID: id, UserID: currentUser(r).ID}); err != nil {
		slog.Error("delete quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not delete quest")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// savePhoto stores the uploaded photo, after checking the account's hourly upload
// limit and storage quota so a script can't fill the server's disk.
func (s *Server) savePhoto(r *http.Request, userID string) (string, error) {
	file, header, err := r.FormFile("photo")
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedPhotoExts[ext] {
		return "", errBadPhotoType
	}
	if s.limits.photos.full(userID) {
		return "", errTooManyPhotos
	}
	used, err := s.photoBytes(r.Context(), userID)
	if err != nil {
		return "", err
	}
	if used+header.Size > s.limits.photoQuota {
		return "", errPhotoQuota
	}
	name := uuid.NewString() + ext
	dst, err := os.Create(s.photoPathFor(name))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		_ = os.Remove(dst.Name())
		return "", err
	}
	s.limits.photos.add(userID)
	return name, nil
}

// photoBytes is how much disk the account's photos take up now.
func (s *Server) photoBytes(ctx context.Context, userID string) (int64, error) {
	names, err := s.store.Queries().ListPhotoPaths(ctx, userID)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, n := range names {
		if info, err := os.Stat(s.photoPathFor(n.String)); err == nil {
			total += info.Size()
		}
	}
	return total, nil
}

func (s *Server) photoPathFor(name string) string {
	return filepath.Join(s.photosDir, name)
}
