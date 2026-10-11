package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/tony-nyagah/sidequestar/api/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionTTL        = 30 * 24 * time.Hour
	minPasswordLength = 8
	maxPasswordLength = 72 // bcrypt ignores anything longer
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

// dummyHash keeps a failed login for an unknown username as slow as a wrong password,
// so response times don't reveal which usernames exist.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)

type userKey struct{}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type sessionDTO struct {
	Token     string  `json:"token"`
	ExpiresAt string  `json:"expires_at"`
	User      userDTO `json:"user"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	c.Username = strings.TrimSpace(c.Username)
	if !usernamePattern.MatchString(c.Username) {
		writeError(w, http.StatusBadRequest, "Usernames are 3 to 32 letters, numbers, dashes or underscores.")
		return
	}
	if len(c.Password) < minPasswordLength || len(c.Password) > maxPasswordLength {
		writeError(w, http.StatusBadRequest, "Passwords need 8 to 72 characters.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash password", "err", err)
		writeError(w, http.StatusInternalServerError, "could not create account")
		return
	}
	user, err := s.store.RegisterUser(r.Context(), c.Username, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			writeError(w, http.StatusConflict, "That username is taken.")
			return
		}
		slog.Error("register", "err", err)
		writeError(w, http.StatusInternalServerError, "could not create account")
		return
	}
	s.startSession(w, r.Context(), user, http.StatusCreated)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	user, err := s.store.Queries().GetUserByUsername(r.Context(), strings.TrimSpace(c.Username))
	hash := dummyHash
	if err == nil {
		hash = []byte(user.PasswordHash)
	} else if !isNoRows(err) {
		slog.Error("login lookup", "err", err)
		writeError(w, http.StatusInternalServerError, "could not sign in")
		return
	}
	if bcrypt.CompareHashAndPassword(hash, []byte(c.Password)) != nil || err != nil {
		writeError(w, http.StatusUnauthorized, "Wrong username or password.")
		return
	}
	s.startSession(w, r.Context(), user, http.StatusOK)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if token := bearerToken(r); token != "" {
		if err := s.store.Queries().DeleteSession(r.Context(), hashToken(token)); err != nil {
			slog.Error("logout", "err", err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	writeJSON(w, http.StatusOK, userDTO{ID: u.ID, Username: u.Username})
}

func (s *Server) startSession(w http.ResponseWriter, ctx context.Context, user store.User, status int) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		slog.Error("session token", "err", err)
		writeError(w, http.StatusInternalServerError, "could not start session")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	now := time.Now().UTC()
	expires := now.Add(sessionTTL).Format(time.RFC3339)

	q := s.store.Queries()
	_ = q.DeleteExpiredSessions(ctx, now.Format(time.RFC3339))
	if err := q.CreateSession(ctx, store.CreateSessionParams{
		TokenHash: hashToken(token),
		UserID:    user.ID,
		ExpiresAt: expires,
	}); err != nil {
		slog.Error("create session", "err", err)
		writeError(w, http.StatusInternalServerError, "could not start session")
		return
	}
	writeJSON(w, status, sessionDTO{
		Token:     token,
		ExpiresAt: expires,
		User:      userDTO{ID: user.ID, Username: user.Username},
	})
}

// requireUser rejects requests without a live session and puts the user in the context.
func (s *Server) requireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "Sign in first.")
			return
		}
		user, err := s.store.Queries().GetSessionUser(r.Context(), store.GetSessionUserParams{
			TokenHash: hashToken(token),
			Now:       time.Now().UTC().Format(time.RFC3339),
		})
		if err != nil {
			if !isNoRows(err) {
				slog.Error("session lookup", "err", err)
			}
			writeError(w, http.StatusUnauthorized, "Your session has expired. Sign in again.")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	}
}

func currentUser(r *http.Request) store.User {
	u, _ := r.Context().Value(userKey{}).(store.User)
	return u
}

func bearerToken(r *http.Request) string {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		return ""
	}
	return strings.TrimSpace(token)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
