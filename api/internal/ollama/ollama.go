package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

type Client struct {
	api   *api.Client
	model string
}

func New(model string) (*Client, error) {
	c, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	if model == "" {
		model = DefaultModel
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := c.Version(ctx); err != nil {
		return nil, fmt.Errorf("ollama unreachable: %w", err)
	}

	return &Client{api: c, model: model}, nil
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
		return GeneratedQuest{}, fmt.Errorf("ollama chat: %w", err)
	}
	if out.Title == "" {
		return GeneratedQuest{}, errors.New("model returned no quest")
	}
	return out, nil
}
