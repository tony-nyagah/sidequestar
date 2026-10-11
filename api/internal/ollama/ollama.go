package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/ollama/ollama/api"
)

const DefaultModel = "llama3.2:3b"

var (
	// ErrUnreachable means the Ollama server isn't running or can't be reached.
	ErrUnreachable = errors.New("ollama unreachable")
	// ErrModelMissing means Ollama is running but the model hasn't been pulled.
	ErrModelMissing = errors.New("model not pulled")
)

type GeneratedQuest struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	DurationBucket string   `json:"duration_bucket"`
	Tags           []string `json:"tags"`
}

type Generator interface {
	Generate(ctx context.Context, prompt string) (GeneratedQuest, error)
	Complete(ctx context.Context, prompt string) (string, error)
	Model() string
}

type Client struct {
	api   *api.Client
	model string
}

// New builds a client without contacting Ollama, so the server can start
// before Ollama does. Reachability is checked on each Generate call.
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

func (c *Client) Model() string {
	return c.model
}

func (c *Client) Generate(ctx context.Context, prompt string) (GeneratedQuest, error) {
	stream := false
	req := &api.ChatRequest{
		Model:    c.model,
		Format:   json.RawMessage(`"json"`),
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
		return GeneratedQuest{}, classify(err)
	}
	if out.Title == "" {
		return GeneratedQuest{}, errors.New("model returned no quest")
	}
	return out, nil
}

// Complete returns the model's plain-text answer to a single prompt. It samples
// at a low temperature: it writes about real people, so literal beats creative.
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	stream := false
	req := &api.ChatRequest{
		Model:    c.model,
		Stream:   &stream,
		Messages: []api.Message{{Role: "user", Content: prompt}},
		Options:  map[string]any{"temperature": 0.2},
	}
	var out strings.Builder
	err := c.api.Chat(ctx, req, func(resp api.ChatResponse) error {
		out.WriteString(resp.Message.Content)
		return nil
	})
	if err != nil {
		return "", classify(err)
	}
	return strings.TrimSpace(out.String()), nil
}

func classify(err error) error {
	var statusErr api.StatusError
	var netErr *net.OpError
	switch {
	case errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound:
		return fmt.Errorf("%w: %w", ErrModelMissing, err)
	case errors.As(err, &netErr):
		return fmt.Errorf("%w: %w", ErrUnreachable, err)
	}
	return fmt.Errorf("ollama chat: %w", err)
}
