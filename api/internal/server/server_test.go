package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tony-nyagah/sidequestar/api/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return New(st, filepath.Join(t.TempDir(), "photos"))
}

func doJSON(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestWalkingSkeleton(t *testing.T) {
	h := newTestServer(t).Handler()

	create := doJSON(t, h, http.MethodPost, "/api/quests",
		`{"title":"Run 5k","description":"Jog around the park","duration_bucket":"30-60m","tags":["fitness","outdoors"]}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", create.Code, create.Body.String())
	}
	var created questDTO
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.Status != "active" || created.Source != "hand" {
		t.Fatalf("unexpected quest: %+v", created)
	}

	list := doJSON(t, h, http.MethodGet, "/api/quests", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d", list.Code)
	}

	complete := doJSON(t, h, http.MethodPost, "/api/quests/"+created.ID+"/complete", "")
	if complete.Code != http.StatusOK {
		t.Fatalf("complete status = %d, body = %s", complete.Code, complete.Body.String())
	}
	var done questDTO
	if err := json.Unmarshal(complete.Body.Bytes(), &done); err != nil {
		t.Fatalf("decode completed: %v", err)
	}
	wantXP := 20 + 20 // 30-60m base + completion bonus, no photo
	if done.XpAwarded != wantXP {
		t.Fatalf("xp = %d, want %d", done.XpAwarded, wantXP)
	}

	profile := doJSON(t, h, http.MethodGet, "/api/profile", "")
	if profile.Code != http.StatusOK {
		t.Fatalf("profile status = %d", profile.Code)
	}
	var p profileDTO
	if err := json.Unmarshal(profile.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if p.XpTotal != wantXP {
		t.Fatalf("profile xp = %d, want %d", p.XpTotal, wantXP)
	}
	if p.Completed != 1 {
		t.Fatalf("completed count = %d, want 1", p.Completed)
	}
	if len(p.Interests) != 2 {
		t.Fatalf("interests = %v, want 2", p.Interests)
	}
}
