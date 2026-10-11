package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/tony-nyagah/sidequestar/api/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// handleChangePassword keeps the current session and signs out every other one.
func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	user := currentUser(r)
	if !s.checkPassword(w, r, user, req.CurrentPassword) {
		return
	}
	if len(req.NewPassword) < minPasswordLength || len(req.NewPassword) > maxPasswordLength {
		writeError(w, http.StatusBadRequest, "Passwords need 8 to 72 characters.")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("hash password", "err", err)
		writeError(w, http.StatusInternalServerError, "could not change password")
		return
	}
	q := s.store.Queries()
	if err := q.UpdatePasswordHash(r.Context(), store.UpdatePasswordHashParams{ID: user.ID, PasswordHash: string(hash)}); err != nil {
		slog.Error("update password", "err", err)
		writeError(w, http.StatusInternalServerError, "could not change password")
		return
	}
	_ = q.DeleteOtherSessions(r.Context(), store.DeleteOtherSessionsParams{UserID: user.ID, TokenHash: hashToken(bearerToken(r))})
	w.WriteHeader(http.StatusNoContent)
}

type deleteAccountRequest struct {
	Password string `json:"password"`
}

// handleDeleteAccount removes the account, its quests, profile, summary, sessions and photos.
func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	var req deleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	user := currentUser(r)
	if !s.checkPassword(w, r, user, req.Password) {
		return
	}
	photos, err := s.store.DeleteUser(r.Context(), user.ID)
	if err != nil {
		slog.Error("delete account", "err", err)
		writeError(w, http.StatusInternalServerError, "could not delete account")
		return
	}
	for _, name := range photos {
		if err := os.Remove(s.photoPathFor(name)); err != nil && !os.IsNotExist(err) {
			slog.Warn("delete photo", "name", name, "err", err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// checkPassword confirms the signed-in user's password before a sensitive change,
// counting failures against the same limits as sign-in.
func (s *Server) checkPassword(w http.ResponseWriter, r *http.Request, user store.User, password string) bool {
	name, ip := user.Username, clientIP(r)
	if s.limits.loginByUser.full(name) || s.limits.loginByIP.full(ip) {
		tooMany(w, "Too many wrong passwords. Wait 15 minutes and try again.")
		return false
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		s.limits.loginByUser.add(name)
		s.limits.loginByIP.add(ip)
		writeError(w, http.StatusUnauthorized, "That password is wrong.")
		return false
	}
	return true
}

// ResetPassword gives the account a new random password and signs it out everywhere.
// It's for the server operator: there's no email, so this is the recovery path.
func ResetPassword(ctx context.Context, st *store.Store, username string) (string, error) {
	q := st.Queries()
	user, err := q.GetUserByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	raw := make([]byte, 12)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	password := base64.RawURLEncoding.EncodeToString(raw)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	if err := q.UpdatePasswordHash(ctx, store.UpdatePasswordHashParams{ID: user.ID, PasswordHash: string(hash)}); err != nil {
		return "", err
	}
	if err := q.DeleteSessionsForUser(ctx, user.ID); err != nil {
		return "", err
	}
	return password, nil
}
