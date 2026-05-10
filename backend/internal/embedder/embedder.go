package embedder

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

const (
	EmbedModel   = "gemini-embedding-2"
	maxBatchSize = 100
	// Output dimension for gemini-embedding-2
	EmbedDim = 3072
)

type Embedder struct {
	client *genai.Client
}

func New(apiKey string) (*Embedder, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &Embedder{client: client}, nil
}

// Embed embeds a batch of texts using Google text-embedding-004.
func (e *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	var all [][]float32
	for i := 0; i < len(texts); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]

		contents := make([]*genai.Content, len(batch))
		for j, t := range batch {
			contents[j] = genai.NewContentFromText(t, genai.RoleUser)
		}

		result, err := e.client.Models.EmbedContent(ctx, EmbedModel, contents, nil)
		if err != nil {
			return nil, fmt.Errorf("embed batch %d: %w", i/maxBatchSize, err)
		}
		for _, emb := range result.Embeddings {
			all = append(all, emb.Values)
		}
	}
	return all, nil
}
