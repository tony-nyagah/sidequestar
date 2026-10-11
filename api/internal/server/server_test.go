package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tony-nyagah/sidequestar/api/internal/ollama"
	"github.com/tony-nyagah/sidequestar/api/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	s := New(st, filepath.Join(t.TempDir(), "photos"))
	t.Cleanup(func() {
		s.Wait()
		st.Close()
	})
	return s
}

// client sends requests as one signed-in user (or anonymously when token is "").
type client struct {
	t     *testing.T
	h     http.Handler
	token string
}

func (c client) do(method, path, body string) *httptest.ResponseRecorder {
	c.t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		r.Header.Set("Authorization", "Bearer "+c.token)
	}
	w := httptest.NewRecorder()
	c.h.ServeHTTP(w, r)
	return w
}

func (c client) decode(w *httptest.ResponseRecorder, v any) {
	c.t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		c.t.Fatalf("decode %s: %v", w.Body.String(), err)
	}
}

func register(t *testing.T, h http.Handler, username string) client {
	t.Helper()
	anon := client{t: t, h: h}
	w := anon.do(http.MethodPost, "/api/auth/register", `{"username":"`+username+`","password":"correct horse"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("register %s = %d %s", username, w.Code, w.Body.String())
	}
	var s sessionDTO
	anon.decode(w, &s)
	return client{t: t, h: h, token: s.Token}
}

func (c client) createQuest(title string, tags ...string) questDTO {
	c.t.Helper()
	body, _ := json.Marshal(map[string]any{"title": title, "description": "Do it", "duration_bucket": "30-60m", "tags": tags})
	w := c.do(http.MethodPost, "/api/quests", string(body))
	if w.Code != http.StatusCreated {
		c.t.Fatalf("create quest = %d %s", w.Code, w.Body.String())
	}
	var q questDTO
	c.decode(w, &q)
	return q
}

func (c client) profile() profileDTO {
	c.t.Helper()
	var p profileDTO
	c.decode(c.do(http.MethodGet, "/api/profile", ""), &p)
	return p
}

func TestWalkingSkeleton(t *testing.T) {
	c := register(t, newTestServer(t).Handler(), "tony")

	var seeded []questDTO
	c.decode(c.do(http.MethodGet, "/api/quests", ""), &seeded)
	if len(seeded) != 10 {
		t.Fatalf("new account has %d quests, want the 10 seeds", len(seeded))
	}

	created := c.createQuest("Run 5k", "fitness", "outdoors")
	if created.Status != "active" || created.Source != "hand" {
		t.Fatalf("unexpected quest: %+v", created)
	}

	complete := c.do(http.MethodPost, "/api/quests/"+created.ID+"/complete", "")
	if complete.Code != http.StatusOK {
		t.Fatalf("complete status = %d, body = %s", complete.Code, complete.Body.String())
	}
	var done questDTO
	c.decode(complete, &done)
	wantXP := 20 + 20 // 30-60m base + completion bonus, no photo
	if done.XpAwarded != wantXP {
		t.Fatalf("xp = %d, want %d", done.XpAwarded, wantXP)
	}

	p := c.profile()
	if p.Username != "tony" || p.XpTotal != wantXP || p.Completed != 1 {
		t.Fatalf("profile = %+v", p)
	}
}

func TestAuth(t *testing.T) {
	h := newTestServer(t).Handler()
	anon := client{t: t, h: h}

	if w := anon.do(http.MethodGet, "/api/quests", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous list = %d, want 401", w.Code)
	}
	if w := anon.do(http.MethodPost, "/api/auth/register", `{"username":"a","password":"correct horse"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("short username = %d, want 400", w.Code)
	}
	if w := anon.do(http.MethodPost, "/api/auth/register", `{"username":"amina","password":"short"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("short password = %d, want 400", w.Code)
	}

	register(t, h, "amina")
	if w := anon.do(http.MethodPost, "/api/auth/register", `{"username":"AMINA","password":"correct horse"}`); w.Code != http.StatusConflict {
		t.Fatalf("duplicate username (any case) = %d, want 409", w.Code)
	}
	if w := anon.do(http.MethodPost, "/api/auth/login", `{"username":"amina","password":"wrong password"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password = %d, want 401", w.Code)
	}
	if w := anon.do(http.MethodPost, "/api/auth/login", `{"username":"nobody","password":"correct horse"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("unknown user = %d, want 401", w.Code)
	}

	w := anon.do(http.MethodPost, "/api/auth/login", `{"username":"amina","password":"correct horse"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login = %d %s", w.Code, w.Body.String())
	}
	var s sessionDTO
	anon.decode(w, &s)
	c := client{t: t, h: h, token: s.Token}
	if w := c.do(http.MethodGet, "/api/me", ""); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "amina") {
		t.Fatalf("me = %d %s", w.Code, w.Body.String())
	}

	c.do(http.MethodPost, "/api/auth/logout", "")
	if w := c.do(http.MethodGet, "/api/me", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("after logout = %d, want 401", w.Code)
	}
}

func TestUsersOnlySeeTheirOwnData(t *testing.T) {
	h := newTestServer(t).Handler()
	amina := register(t, h, "amina")
	brian := register(t, h, "brian")

	quest := amina.createQuest("Paint a mural", "art")
	if w := brian.do(http.MethodPost, "/api/quests/"+quest.ID+"/complete", ""); w.Code != http.StatusNotFound {
		t.Fatalf("brian completing amina's quest = %d, want 404", w.Code)
	}
	brian.do(http.MethodDelete, "/api/quests/"+quest.ID, "")
	if w := amina.do(http.MethodPost, "/api/quests/"+quest.ID+"/complete", ""); w.Code != http.StatusOK {
		t.Fatalf("amina's quest should survive brian's delete; complete = %d", w.Code)
	}

	var brianQuests []questDTO
	brian.decode(brian.do(http.MethodGet, "/api/quests", ""), &brianQuests)
	for _, q := range brianQuests {
		if q.ID == quest.ID {
			t.Fatal("brian can list amina's quest")
		}
	}

	amina.do(http.MethodPatch, "/api/profile", `{"note":"I love murals"}`)
	if got := brian.profile().Note; got != "" {
		t.Fatalf("brian sees note %q", got)
	}
	if amina.profile().XpTotal == 0 || brian.profile().XpTotal != 0 {
		t.Fatal("XP leaked between accounts")
	}
}

type fakeGen struct{ err error }

func (f fakeGen) Generate(context.Context, string) (ollama.GeneratedQuest, error) {
	return ollama.GeneratedQuest{}, f.err
}

func (f fakeGen) Complete(context.Context, string) (string, error) { return "", f.err }

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
		c := register(t, s.Handler(), "tony")
		for _, path := range []string{"/api/quests/generate", "/api/profile/summary"} {
			w := c.do(http.MethodPost, path, `{"context":"free evening"}`)
			if w.Code != tc.wantCode || !strings.Contains(w.Body.String(), tc.wantMsg) {
				t.Errorf("%s %v: got %d %s, want %d containing %q", path, tc.err, w.Code, w.Body.String(), tc.wantCode, tc.wantMsg)
			}
		}
	}
}

func (c client) completeWithPhoto(id, filename string) *httptest.ResponseRecorder {
	c.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("photo", filename)
	fw.Write([]byte("img"))
	mw.Close()
	r := httptest.NewRequest(http.MethodPost, "/api/quests/"+id+"/complete", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.Header.Set("Authorization", "Bearer "+c.token)
	w := httptest.NewRecorder()
	c.h.ServeHTTP(w, r)
	return w
}

func TestPhotoUploadAndServe(t *testing.T) {
	h := newTestServer(t).Handler()
	c := register(t, h, "tony")

	if w := c.completeWithPhoto(c.createQuest("Sketch a tree").ID, "page.html"); w.Code != http.StatusBadRequest {
		t.Fatalf("html upload status = %d, want 400", w.Code)
	}

	w := c.completeWithPhoto(c.createQuest("Sketch a tree").ID, "tree.JPG")
	if w.Code != http.StatusOK {
		t.Fatalf("complete status = %d, body = %s", w.Code, w.Body.String())
	}
	var done questDTO
	c.decode(w, &done)
	if done.XpAwarded != 20+20+30 {
		t.Fatalf("xp = %d, want 70", done.XpAwarded)
	}

	if got := c.do(http.MethodGet, "/api/photos/"+done.PhotoPath, ""); got.Code != http.StatusOK || got.Body.String() != "img" {
		t.Fatalf("serve photo = %d %q", got.Code, got.Body.String())
	}
	if got := c.do(http.MethodGet, "/api/photos/..%2Ftest.db", ""); got.Code != http.StatusNotFound {
		t.Fatalf("traversal status = %d, want 404", got.Code)
	}
	other := register(t, h, "someone")
	if got := other.do(http.MethodGet, "/api/photos/"+done.PhotoPath, ""); got.Code != http.StatusNotFound {
		t.Fatalf("another user's photo = %d, want 404", got.Code)
	}
}

// captureGen records prompts and answers with fixed text.
type captureGen struct {
	prompt        string
	summaryPrompt string
	summary       string
}

func (c *captureGen) Generate(_ context.Context, prompt string) (ollama.GeneratedQuest, error) {
	c.prompt = prompt
	return ollama.GeneratedQuest{Title: "Walk", Description: "Go", DurationBucket: "under-30m"}, nil
}

func (c *captureGen) Complete(_ context.Context, prompt string) (string, error) {
	c.summaryPrompt = prompt
	return c.summary, nil
}

func (c *captureGen) Model() string { return "test-model" }

func TestProfileShowsWhatTheModelIsTold(t *testing.T) {
	s := newTestServer(t)
	gen := &captureGen{}
	s.SetGenerator(gen)
	c := register(t, s.Handler(), "tony")

	c.do(http.MethodPatch, "/api/profile", `{"location":"Nairobi","note":"Recovering from a cold","summary":"Likes slow walks."}`)

	p := c.profile()
	if p.Model != "test-model" {
		t.Fatalf("model = %q", p.Model)
	}
	want := []string{
		"Their location: Nairobi.",
		"In their own words: Recovering from a cold.",
		"What you have learned about them: Likes slow walks.",
	}
	if strings.Join(p.KnownFacts, "|") != strings.Join(want, "|") {
		t.Fatalf("known_facts = %q, want %q", p.KnownFacts, want)
	}

	c.do(http.MethodPost, "/api/quests/generate", `{"context":"free evening"}`)
	for _, fact := range p.KnownFacts {
		if !strings.Contains(gen.prompt, fact) {
			t.Errorf("prompt is missing %q", fact)
		}
	}
}

func TestSummaryUpdatesAfterCompletionAndOnDemand(t *testing.T) {
	s := newTestServer(t)
	gen := &captureGen{summary: "Tony finishes short art quests."}
	s.SetGenerator(gen)
	c := register(t, s.Handler(), "tony")
	c.do(http.MethodPatch, "/api/profile", `{"note":"I hate running"}`)

	quest := c.createQuest("Sketch the view", "art")
	c.do(http.MethodPost, "/api/quests/"+quest.ID+"/complete", "")
	s.Wait()

	p := c.profile()
	if p.Summary != gen.summary || p.SummaryUpdatedAt == "" {
		t.Fatalf("summary after completion = %q (updated %q)", p.Summary, p.SummaryUpdatedAt)
	}
	for _, want := range []string{"I hate running", "Sketch the view"} {
		if !strings.Contains(gen.summaryPrompt, want) {
			t.Errorf("summary prompt is missing %q", want)
		}
	}

	// A hand edit sticks, and the next refresh is told to keep it.
	c.do(http.MethodPatch, "/api/profile", `{"summary":"Tony prefers mornings."}`)
	if got := c.profile().Summary; got != "Tony prefers mornings." {
		t.Fatalf("edited summary = %q", got)
	}
	gen.summary = "Tony prefers mornings and art."
	if w := c.do(http.MethodPost, "/api/profile/summary", ""); w.Code != http.StatusOK {
		t.Fatalf("refresh = %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(gen.summaryPrompt, "Tony prefers mornings.") {
		t.Error("refresh prompt doesn't include the hand-edited summary")
	}
	if got := c.profile().Summary; got != gen.summary {
		t.Fatalf("summary after refresh = %q", got)
	}

	// The model's own output is never fed back, only the person's edit,
	// so a detail the model invented can't keep reinforcing itself.
	if w := c.do(http.MethodPost, "/api/profile/summary", ""); w.Code != http.StatusOK {
		t.Fatalf("second refresh = %d", w.Code)
	}
	if strings.Contains(gen.summaryPrompt, "Tony prefers mornings and art.") {
		t.Error("refresh prompt includes the model's previous output")
	}
}

func TestQuestPromptForbidsInventedPlaces(t *testing.T) {
	s := newTestServer(t)
	gen := &captureGen{}
	s.SetGenerator(gen)
	c := register(t, s.Handler(), "tony")
	c.do(http.MethodPost, "/api/quests/generate", `{"context":"free evening"}`)
	if !strings.Contains(gen.prompt, "Never name a specific place") {
		t.Fatal("quest prompt no longer tells the model not to invent places")
	}
}

func (c client) doFrom(ip, method, path, body string) *httptest.ResponseRecorder {
	c.t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Client-IP", ip)
	if c.token != "" {
		r.Header.Set("Authorization", "Bearer "+c.token)
	}
	w := httptest.NewRecorder()
	c.h.ServeHTTP(w, r)
	return w
}

func TestSignInLockout(t *testing.T) {
	h := newTestServer(t).Handler()
	register(t, h, "amina")
	anon := client{t: t, h: h}
	for i := 0; i < 5; i++ {
		anon.doFrom("10.0.0.1", http.MethodPost, "/api/auth/login", `{"username":"amina","password":"wrong password"}`)
	}
	// Locked by username, even from another address and with the right password.
	if w := anon.doFrom("10.0.0.2", http.MethodPost, "/api/auth/login", `{"username":"Amina","password":"correct horse"}`); w.Code != http.StatusTooManyRequests {
		t.Fatalf("after 5 failures = %d, want 429", w.Code)
	}
	register(t, h, "brian")
	if w := anon.doFrom("10.0.0.2", http.MethodPost, "/api/auth/login", `{"username":"brian","password":"correct horse"}`); w.Code != http.StatusOK {
		t.Fatalf("another user = %d, want 200", w.Code)
	}
}

func TestRegistrationLimitPerAddress(t *testing.T) {
	anon := client{t: t, h: newTestServer(t).Handler()}
	for i := 0; i < 5; i++ {
		if w := anon.doFrom("10.0.0.9", http.MethodPost, "/api/auth/register", fmt.Sprintf(`{"username":"user%d","password":"correct horse"}`, i)); w.Code != http.StatusCreated {
			t.Fatalf("register %d = %d", i, w.Code)
		}
	}
	if w := anon.doFrom("10.0.0.9", http.MethodPost, "/api/auth/register", `{"username":"user5","password":"correct horse"}`); w.Code != http.StatusTooManyRequests {
		t.Fatalf("6th account = %d, want 429", w.Code)
	}
	if w := anon.doFrom("10.0.0.10", http.MethodPost, "/api/auth/register", `{"username":"user5","password":"correct horse"}`); w.Code != http.StatusCreated {
		t.Fatalf("other address = %d, want 201", w.Code)
	}
}

func TestChangePasswordKeepsOnlyThisSession(t *testing.T) {
	h := newTestServer(t).Handler()
	phone := register(t, h, "amina")
	var s sessionDTO
	anon := client{t: t, h: h}
	anon.decode(anon.do(http.MethodPost, "/api/auth/login", `{"username":"amina","password":"correct horse"}`), &s)
	laptop := client{t: t, h: h, token: s.Token}

	if w := laptop.do(http.MethodPost, "/api/me/password", `{"current_password":"wrong","new_password":"new password 1"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong current password = %d, want 401", w.Code)
	}
	if w := laptop.do(http.MethodPost, "/api/me/password", `{"current_password":"correct horse","new_password":"new password 1"}`); w.Code != http.StatusNoContent {
		t.Fatalf("change password = %d %s", w.Code, w.Body.String())
	}
	if w := laptop.do(http.MethodGet, "/api/me", ""); w.Code != http.StatusOK {
		t.Fatalf("this session after change = %d, want 200", w.Code)
	}
	if w := phone.do(http.MethodGet, "/api/me", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("other session after change = %d, want 401", w.Code)
	}
	if w := anon.do(http.MethodPost, "/api/auth/login", `{"username":"amina","password":"new password 1"}`); w.Code != http.StatusOK {
		t.Fatalf("login with new password = %d", w.Code)
	}
}

func TestDeleteAccountRemovesEverything(t *testing.T) {
	s := newTestServer(t)
	h := s.Handler()
	amina := register(t, h, "amina")
	brian := register(t, h, "brian")
	done := amina.completeWithPhoto(amina.createQuest("Paint").ID, "mural.jpg")
	var q questDTO
	amina.decode(done, &q)
	photo := s.photoPathFor(q.PhotoPath)
	if _, err := os.Stat(photo); err != nil {
		t.Fatalf("photo not saved: %v", err)
	}

	if w := amina.do(http.MethodDelete, "/api/me", `{"password":"wrong"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("delete with wrong password = %d, want 401", w.Code)
	}
	if w := amina.do(http.MethodDelete, "/api/me", `{"password":"correct horse"}`); w.Code != http.StatusNoContent {
		t.Fatalf("delete = %d %s", w.Code, w.Body.String())
	}
	if w := amina.do(http.MethodGet, "/api/me", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("session after delete = %d, want 401", w.Code)
	}
	if _, err := os.Stat(photo); !os.IsNotExist(err) {
		t.Fatalf("photo file still there: %v", err)
	}
	anon := client{t: t, h: h}
	if w := anon.do(http.MethodPost, "/api/auth/login", `{"username":"amina","password":"correct horse"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("login after delete = %d, want 401", w.Code)
	}
	if w := brian.do(http.MethodGet, "/api/me", ""); w.Code != http.StatusOK {
		t.Fatalf("other account after delete = %d, want 200", w.Code)
	}
	// The username is free again.
	register(t, h, "amina")
}

func TestResetPassword(t *testing.T) {
	s := newTestServer(t)
	h := s.Handler()
	amina := register(t, h, "amina")
	password, err := ResetPassword(context.Background(), s.store, "amina")
	if err != nil || len(password) < 12 {
		t.Fatalf("reset = %q, %v", password, err)
	}
	if w := amina.do(http.MethodGet, "/api/me", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("old session after reset = %d, want 401", w.Code)
	}
	anon := client{t: t, h: h}
	body, _ := json.Marshal(map[string]string{"username": "amina", "password": password})
	if w := anon.do(http.MethodPost, "/api/auth/login", string(body)); w.Code != http.StatusOK {
		t.Fatalf("login with reset password = %d", w.Code)
	}
}

func TestQuestsPerHourCap(t *testing.T) {
	s := newTestServer(t)
	s.SetGenerator(&captureGen{})
	c := register(t, s.Handler(), "amina")
	for i := 0; i < 30; i++ {
		if w := c.do(http.MethodPost, "/api/quests/generate", `{"context":"free evening"}`); w.Code != http.StatusCreated {
			t.Fatalf("quest %d = %d", i, w.Code)
		}
	}
	if w := c.do(http.MethodPost, "/api/quests/generate", `{"context":"free evening"}`); w.Code != http.StatusTooManyRequests {
		t.Fatalf("31st quest = %d, want 429", w.Code)
	}
}

func TestPhotoLimits(t *testing.T) {
	s := newTestServer(t)
	c := register(t, s.Handler(), "amina")
	for i := 0; i < 20; i++ {
		if w := c.completeWithPhoto(c.createQuest("Walk").ID, "walk.jpg"); w.Code != http.StatusOK {
			t.Fatalf("photo %d = %d %s", i, w.Code, w.Body.String())
		}
	}
	if w := c.completeWithPhoto(c.createQuest("Walk").ID, "walk.jpg"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("21st photo in an hour = %d, want 429", w.Code)
	}
	// Completing without a photo still works.
	if w := c.do(http.MethodPost, "/api/quests/"+c.createQuest("Walk").ID+"/complete", ""); w.Code != http.StatusOK {
		t.Fatalf("complete without photo = %d", w.Code)
	}

	// Quota: each test photo is 3 bytes, so a 5-byte quota fits one and refuses the next.
	s2 := newTestServer(t)
	s2.limits.photoQuota = 5
	c2 := register(t, s2.Handler(), "brian")
	if w := c2.completeWithPhoto(c2.createQuest("Sketch").ID, "a.jpg"); w.Code != http.StatusOK {
		t.Fatalf("first photo = %d", w.Code)
	}
	quest := c2.createQuest("Sketch")
	if w := c2.completeWithPhoto(quest.ID, "b.jpg"); w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("over quota = %d, want 413", w.Code)
	}
	var q questDTO
	if w := c2.do(http.MethodPost, "/api/quests/"+quest.ID+"/complete", ""); w.Code != http.StatusOK {
		t.Fatalf("refused photo should leave the quest completable, got %d", w.Code)
	} else {
		c2.decode(w, &q)
		if q.PhotoPath != "" {
			t.Fatal("refused photo was attached anyway")
		}
	}
}

// blockingGen holds every call until release is closed.
type blockingGen struct{ release chan struct{} }

func (b blockingGen) Generate(ctx context.Context, _ string) (ollama.GeneratedQuest, error) {
	select {
	case <-b.release:
		return ollama.GeneratedQuest{Title: "Walk", Description: "Go", DurationBucket: "under-30m"}, nil
	case <-ctx.Done():
		return ollama.GeneratedQuest{}, ctx.Err()
	}
}

func (b blockingGen) Complete(context.Context, string) (string, error) { return "", nil }

func (b blockingGen) Model() string { return "test-model" }

func TestBusyModelFailsFast(t *testing.T) {
	s := newTestServer(t)
	gen := blockingGen{release: make(chan struct{})}
	s.SetGenerator(gen)
	c := register(t, s.Handler(), "amina")

	// One request on the model, three queued behind it.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.do(http.MethodPost, "/api/quests/generate", `{"context":"free evening"}`)
		}()
	}
	deadline := time.Now().Add(5 * time.Second)
	for s.limits.waiting.Load() < 3 {
		if time.Now().After(deadline) {
			t.Fatalf("queue never filled (waiting = %d)", s.limits.waiting.Load())
		}
		time.Sleep(10 * time.Millisecond)
	}

	start := time.Now()
	w := c.do(http.MethodPost, "/api/quests/generate", `{"context":"free evening"}`)
	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "busy") {
		t.Fatalf("5th request = %d %s, want 503 busy", w.Code, w.Body.String())
	}
	if time.Since(start) > time.Second {
		t.Fatalf("busy answer took %v; it should be immediate", time.Since(start))
	}

	close(gen.release)
	wg.Wait()
}
