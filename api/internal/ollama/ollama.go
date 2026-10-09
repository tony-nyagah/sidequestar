package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
)

const DefaultModel = "llama3.2:3b"

type GeneratedQuest struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	DurationBucket string   `json:"duration_bucket"`
	Tags           []string `json:"tags"`
}

type Generator interface {
	Generate(ctx context.Context, prompt string) (GeneratedQuest, error)
}

// questSchema is passed as Ollama's structured-output format so the model
// is constrained to the exact shape (and bucket values) we accept.
var questSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "title": {"type": "string"},
    "description": {"type": "string"},
    "duration_bucket": {"type": "string", "enum": ["under-30m", "30-60m", "1-2h", "half-day", "full-day"]},
    "tags": {"type": "array", "items": {"type": "string"}, "minItems": 2, "maxItems": 4}
  },
  "required": ["title", "description", "duration_bucket", "tags"]
}`)

type Client struct {
	api   *api.Client
	model string
}

// New builds a client without contacting Ollama, so the API keeps working
// if Ollama is started (or restarted) after the server.
func New(model string) (*Client, error) {
	c, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	if model == "" {
		model = DefaultModel
	}
	return &Client{api: c, model: model}, nil
}

func (c *Client) Model() string { return c.model }

// Ping reports whether Ollama is reachable.
func (c *Client) Ping(ctx context.Context) error {
	if _, err := c.api.Version(ctx); err != nil {
		return fmt.Errorf("ollama unreachable: %w", err)
	}
	return nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (GeneratedQuest, error) {
	stream := false
	req := &api.ChatRequest{
		Model:    c.model,
		Format:   questSchema,
		Stream:   &stream,
		Messages: []api.Message{{Role: "user", Content: prompt}},
	}

	var out GeneratedQuest
	err := c.api.Chat(ctx, req, func(resp api.ChatResponse) error {
		if resp.Message.Content == "" {
			return nil
		}
		return json.Unmarshal([]byte(resp.Message.Content), &out)
	})
	if err != nil {
		return GeneratedQuest{}, fmt.Errorf("ollama chat: %w", err)
	}
	if strings.TrimSpace(out.Title) == "" || strings.TrimSpace(out.Description) == "" {
		return GeneratedQuest{}, errors.New("model returned an incomplete quest")
	}
	return out, nil
}
