package server

import (
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

var errBadPhotoType = errors.New("unsupported photo type")

var allowedPhotoExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true, ".heic": true,
}

func (s *Server) handleListQuests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	var list []store.Quest
	var err error
	if status == "" {
		list, err = s.store.Queries().ListQuests(r.Context())
	} else {
		list, err = s.store.Queries().ListQuestsByStatus(r.Context(), status)
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
	q, err := s.store.Queries().AcceptQuest(r.Context(), id)
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
	q, err := s.store.Queries().GetQuest(r.Context(), id)
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
		photoPath, err = s.savePhoto(r)
		if err != nil {
			var tooBig *http.MaxBytesError
			switch {
			case errors.As(err, &tooBig):
				writeError(w, http.StatusRequestEntityTooLarge, "photo is too large (max 20 MB)")
			case errors.Is(err, errBadPhotoType):
				writeError(w, http.StatusBadRequest, "photo must be a JPEG, PNG, WebP, GIF, or HEIC image")
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
		CompletedAt: toNullString(now),
		PhotoPath:   toNullString(photoPath),
		XpAwarded:   toNullInt(int64(awarded)),
	})
	if err != nil {
		slog.Error("complete quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not complete quest")
		return
	}
	writeJSON(w, http.StatusOK, toQuestDTO(completed))
}

func (s *Server) handleDeleteQuest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.Queries().DeleteQuest(r.Context(), id); err != nil {
		slog.Error("delete quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not delete quest")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) savePhoto(r *http.Request) (string, error) {
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
	name := uuid.NewString() + ext
	dst, err := os.Create(s.photoPathFor(name))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return name, nil
}

func (s *Server) photoPathFor(name string) string {
	return filepath.Join(s.photosDir, name)
}
