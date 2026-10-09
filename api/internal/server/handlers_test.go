package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/tony-nyagah/sidequestar/api/internal/ollama"
)

type fakeGenerator struct {
	quest  ollama.GeneratedQuest
	err    error
	prompt string
}

func (f *fakeGenerator) Generate(_ context.Context, prompt string) (ollama.GeneratedQuest, error) {
	f.prompt = prompt
	return f.quest, f.err
}

func decode[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %s: %v", w.Body.String(), err)
	}
	return v
}

func createQuest(t *testing.T, h http.Handler) questDTO {
	t.Helper()
	w := doJSON(t, h, http.MethodPost, "/api/quests",
		`{"title":"Walk","description":"Around the block","duration_bucket":"under-30m","tags":[" Outdoors ","outdoors","Walking"]}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", w.Code, w.Body.String())
	}
	return decode[questDTO](t, w)
}

func TestCreateNormalizesTags(t *testing.T) {
	q := createQuest(t, newTestServer(t).Handler())
	if strings.Join(q.Tags, ",") != "outdoors,walking" {
		t.Fatalf("tags = %v, want [outdoors walking]", q.Tags)
	}
}

func TestRejectsNonJSONBody(t *testing.T) {
	h := newTestServer(t).Handler()
	r := httptest.NewRequest(http.MethodPost, "/api/quests",
		strings.NewReader(`{"title":"x","description":"y","duration_bucket":"under-30m"}`))
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415", w.Code)
	}
}

func TestGenerate(t *testing.T) {
	s := newTestServer(t)
	gen := &fakeGenerator{quest: ollama.GeneratedQuest{
		Title:          "Spot three birds",
		Description:    "Sit in the park and find three different birds.",
		DurationBucket: "not-a-bucket",
		Tags:           []string{"Birding", "outdoors"},
	}}
	s.SetGenerator(gen)
	h := s.Handler()

	w := doJSON(t, h, http.MethodPost, "/api/quests/generate", `{"context":"free evening"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	q := decode[questDTO](t, w)
	if q.Status != "generated" || q.DurationBucket != "under-30m" {
		t.Fatalf("unexpected quest: %+v", q)
	}
	if strings.Join(q.Tags, ",") != "birding,outdoors" {
		t.Fatalf("tags = %v", q.Tags)
	}
	if !strings.Contains(gen.prompt, "free evening") {
		t.Fatalf("prompt missing situation: %s", gen.prompt)
	}

	// The previous quest should be listed as one not to repeat.
	doJSON(t, h, http.MethodPost, "/api/quests/generate", `{"context":"again"}`)
	if !strings.Contains(gen.prompt, "Spot three birds") {
		t.Fatalf("prompt missing recent quests: %s", gen.prompt)
	}
}

func TestGenerateWithoutModel(t *testing.T) {
	h := newTestServer(t).Handler()
	w := doJSON(t, h, http.MethodPost, "/api/quests/generate", `{"context":"x"}`)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
}

func TestGenerateModelError(t *testing.T) {
	s := newTestServer(t)
	s.SetGenerator(&fakeGenerator{err: errors.New("boom")})
	w := doJSON(t, s.Handler(), http.MethodPost, "/api/quests/generate", `{"context":"x"}`)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", w.Code)
	}
}

func TestInterestsOnlyCountCompleted(t *testing.T) {
	s := newTestServer(t)
	s.SetGenerator(&fakeGenerator{quest: ollama.GeneratedQuest{
		Title: "T", Description: "D", DurationBucket: "1-2h", Tags: []string{"generated-only"},
	}})
	h := s.Handler()
	doJSON(t, h, http.MethodPost, "/api/quests/generate", `{"context":"x"}`)
	createQuest(t, h) // active, not completed

	p := decode[profileDTO](t, doJSON(t, h, http.MethodGet, "/api/profile", ""))
	if len(p.Interests) != 0 {
		t.Fatalf("interests = %v, want none", p.Interests)
	}
}

func TestAcceptAndCompleteConflicts(t *testing.T) {
	h := newTestServer(t).Handler()
	q := createQuest(t, h) // hand-made quests start active

	if w := doJSON(t, h, http.MethodPost, "/api/quests/"+q.ID+"/accept", ""); w.Code != http.StatusNotFound {
		t.Fatalf("accept active = %d, want 404", w.Code)
	}
	if w := doJSON(t, h, http.MethodPost, "/api/quests/"+q.ID+"/complete", ""); w.Code != http.StatusOK {
		t.Fatalf("complete = %d", w.Code)
	}
	if w := doJSON(t, h, http.MethodPost, "/api/quests/"+q.ID+"/complete", ""); w.Code != http.StatusConflict {
		t.Fatalf("complete twice = %d, want 409", w.Code)
	}
	if w := doJSON(t, h, http.MethodPost, "/api/quests/missing/complete", ""); w.Code != http.StatusNotFound {
		t.Fatalf("complete missing = %d, want 404", w.Code)
	}
}

func TestDelete(t *testing.T) {
	h := newTestServer(t).Handler()
	active := createQuest(t, h)
	if w := doJSON(t, h, http.MethodDelete, "/api/quests/"+active.ID, ""); w.Code != http.StatusNoContent {
		t.Fatalf("delete active = %d, want 204", w.Code)
	}
	if w := doJSON(t, h, http.MethodDelete, "/api/quests/"+active.ID, ""); w.Code != http.StatusNotFound {
		t.Fatalf("delete missing = %d, want 404", w.Code)
	}

	done := createQuest(t, h)
	doJSON(t, h, http.MethodPost, "/api/quests/"+done.ID+"/complete", "")
	if w := doJSON(t, h, http.MethodDelete, "/api/quests/"+done.ID, ""); w.Code != http.StatusConflict {
		t.Fatalf("delete completed = %d, want 409", w.Code)
	}
}

func multipartPhoto(t *testing.T, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("photo", filename)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write(content)
	mw.Close()
	return &buf, mw.FormDataContentType()
}

func pngBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestCompleteWithPhoto(t *testing.T) {
	s := newTestServer(t)
	h := s.Handler()
	q := createQuest(t, h)

	// The client-supplied extension is ignored; the stored one is sniffed.
	body, ct := multipartPhoto(t, "evil.html", pngBytes(t))
	r := httptest.NewRequest(http.MethodPost, "/api/quests/"+q.ID+"/complete", body)
	r.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	done := decode[questDTO](t, w)
	if !strings.HasSuffix(done.PhotoPath, ".png") {
		t.Fatalf("photo path = %q, want .png", done.PhotoPath)
	}
	if done.XpAwarded != 10+20+30 {
		t.Fatalf("xp = %d, want 60", done.XpAwarded)
	}

	photo := doJSON(t, h, http.MethodGet, "/api/photos/"+done.PhotoPath, "")
	if photo.Code != http.StatusOK || photo.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("get photo = %d %q", photo.Code, photo.Header().Get("Content-Type"))
	}
	if w := doJSON(t, h, http.MethodGet, "/api/photos/..%2Fsecret", ""); w.Code != http.StatusNotFound {
		t.Fatalf("traversal = %d, want 404", w.Code)
	}
}

func TestCompleteRejectsNonImage(t *testing.T) {
	s := newTestServer(t)
	h := s.Handler()
	q := createQuest(t, h)

	body, ct := multipartPhoto(t, "photo.png", []byte("<html><script>alert(1)</script></html>"))
	r := httptest.NewRequest(http.MethodPost, "/api/quests/"+q.ID+"/complete", body)
	r.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	entries, _ := os.ReadDir(s.photosDir)
	if len(entries) != 0 {
		t.Fatalf("photos dir has %d files, want 0", len(entries))
	}
}
