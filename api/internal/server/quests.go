package server

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tony-nyagah/sidequestar/api/internal/store"
	"github.com/tony-nyagah/sidequestar/api/internal/xp"
)

const maxPhotoBytes = 10 << 20

// photoExts maps the sniffed content type of an upload to the extension we
// store it under. Anything else is rejected.
var photoExts = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var photoNameRe = regexp.MustCompile(`^[0-9a-f-]{36}\.(jpg|png|gif|webp)$`)

var errBadPhoto = errors.New("photo must be a JPEG, PNG, GIF or WebP image")

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
	if !decodeJSON(w, r, &req) {
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
		Tags:           store.MarshalTags(normalizeTags(req.Tags)),
		CreatedAt:      store.Timestamp(time.Now()),
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

	var photoName string
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, maxPhotoBytes)
		photoName, err = s.savePhoto(r)
		if err != nil {
			var tooBig *http.MaxBytesError
			switch {
			case errors.As(err, &tooBig):
				writeError(w, http.StatusRequestEntityTooLarge, "photo must be under 10 MB")
			case errors.Is(err, errBadPhoto):
				writeError(w, http.StatusBadRequest, errBadPhoto.Error())
			default:
				slog.Error("save photo", "err", err)
				writeError(w, http.StatusInternalServerError, "could not save photo")
			}
			return
		}
	}

	awarded := xp.Award(xp.Bucket(q.DurationBucket), photoName != "")
	now := store.Timestamp(time.Now())

	completed, err := s.store.Queries().CompleteQuest(r.Context(), store.CompleteQuestParams{
		ID:          id,
		CompletedAt: toNullString(now),
		PhotoPath:   toNullString(photoName),
		XpAwarded:   toNullInt(int64(awarded)),
	})
	if err != nil {
		s.removePhoto(photoName)
		if isNoRows(err) {
			// Another request completed or deleted it after our status check.
			writeError(w, http.StatusConflict, "quest is not active")
			return
		}
		slog.Error("complete quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not complete quest")
		return
	}
	writeJSON(w, http.StatusOK, toQuestDTO(completed))
}

// handleDeleteQuest removes generated or active quests. Completed quests are
// kept because total XP is derived from them.
func (s *Server) handleDeleteQuest(w http.ResponseWriter, r *http.Request) {
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
	if q.Status == "completed" {
		writeError(w, http.StatusConflict, "completed quests can't be deleted")
		return
	}
	if err := s.store.Queries().DeleteQuest(r.Context(), id); err != nil {
		slog.Error("delete quest", "err", err)
		writeError(w, http.StatusInternalServerError, "could not delete quest")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetPhoto(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !photoNameRe.MatchString(name) {
		writeError(w, http.StatusNotFound, "photo not found")
		return
	}
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, s.photoPathFor(name))
}

// savePhoto stores the uploaded "photo" field, if any, and returns its file
// name. The extension comes from the sniffed content, not the client's name.
func (s *Server) savePhoto(r *http.Request) (string, error) {
	file, _, err := r.FormFile("photo")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", errBadPhoto
	}
	ext, ok := photoExts[http.DetectContentType(head[:n])]
	if !ok {
		return "", errBadPhoto
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	name := uuid.NewString() + ext
	dst, err := os.Create(s.photoPathFor(name))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		s.removePhoto(name)
		return "", err
	}
	if err := dst.Close(); err != nil {
		s.removePhoto(name)
		return "", err
	}
	return name, nil
}

func (s *Server) removePhoto(name string) {
	if name == "" {
		return
	}
	if err := os.Remove(s.photoPathFor(name)); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("remove photo", "name", name, "err", err)
	}
}

func (s *Server) photoPathFor(name string) string {
	return filepath.Join(s.photosDir, name)
}
