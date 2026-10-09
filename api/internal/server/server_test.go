package server

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tony-nyagah/sidequestar/api/internal/ollama"
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

type fakeGen struct{ err error }

func (f fakeGen) Generate(context.Context, string) (ollama.GeneratedQuest, error) {
	return ollama.GeneratedQuest{}, f.err
}

func (f fakeGen) Model() string { return "test-model" }

func TestGenerateExplainsOllamaFailures(t *testing.T) {
	cases := []struct {
		err      error
		wantCode int
		wantMsg  string
	}{
		{ollama.ErrUnreachable, http.StatusServiceUnavailable, "ollama serve"},
		{ollama.ErrModelMissing, http.StatusServiceUnavailable, "ollama pull test-model"},
	}
	for _, tc := range cases {
		s := newTestServer(t)
		s.SetGenerator(fakeGen{err: tc.err})
		w := doJSON(t, s.Handler(), http.MethodPost, "/api/quests/generate", `{"context":"free evening"}`)
		if w.Code != tc.wantCode || !strings.Contains(w.Body.String(), tc.wantMsg) {
			t.Errorf("%v: got %d %s, want %d containing %q", tc.err, w.Code, w.Body.String(), tc.wantCode, tc.wantMsg)
		}
	}
}

func completeWithPhoto(t *testing.T, h http.Handler, id, filename string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("photo", filename)
	fw.Write([]byte("img"))
	mw.Close()
	r := httptest.NewRequest(http.MethodPost, "/api/quests/"+id+"/complete", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestPhotoUploadAndServe(t *testing.T) {
	h := newTestServer(t).Handler()
	newQuest := func() string {
		w := doJSON(t, h, http.MethodPost, "/api/quests",
			`{"title":"Sketch a tree","description":"Draw one","duration_bucket":"under-30m"}`)
		var q questDTO
		json.Unmarshal(w.Body.Bytes(), &q)
		return q.ID
	}

	if w := completeWithPhoto(t, h, newQuest(), "page.html"); w.Code != http.StatusBadRequest {
		t.Fatalf("html upload status = %d, want 400", w.Code)
	}

	w := completeWithPhoto(t, h, newQuest(), "tree.JPG")
	if w.Code != http.StatusOK {
		t.Fatalf("complete status = %d, body = %s", w.Code, w.Body.String())
	}
	var done questDTO
	json.Unmarshal(w.Body.Bytes(), &done)
	if done.XpAwarded != 10+20+30 {
		t.Fatalf("xp = %d, want 60", done.XpAwarded)
	}

	if got := doJSON(t, h, http.MethodGet, "/api/photos/"+done.PhotoPath, ""); got.Code != http.StatusOK || got.Body.String() != "img" {
		t.Fatalf("serve photo = %d %q", got.Code, got.Body.String())
	}
	if got := doJSON(t, h, http.MethodGet, "/api/photos/..%2Ftest.db", ""); got.Code != http.StatusNotFound {
		t.Fatalf("traversal status = %d, want 404", got.Code)
	}
}
