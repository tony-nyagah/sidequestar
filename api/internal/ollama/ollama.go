package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

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
		var statusErr api.StatusError
		var netErr *net.OpError
		switch {
		case errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound:
			return GeneratedQuest{}, fmt.Errorf("%w: %w", ErrModelMissing, err)
		case errors.As(err, &netErr):
			return GeneratedQuest{}, fmt.Errorf("%w: %w", ErrUnreachable, err)
		}
		return GeneratedQuest{}, fmt.Errorf("ollama chat: %w", err)
	}
	if out.Title == "" {
		return GeneratedQuest{}, errors.New("model returned no quest")
	}
	return out, nil
}
