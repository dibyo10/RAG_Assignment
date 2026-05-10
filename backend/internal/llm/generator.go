package llm

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/dibyochakraborty/notebooklm/internal/store"
)

const (
	Model      = "gemini-3.1-pro-preview"
	systemPrompt = `You are a precise research assistant. Answer questions ONLY using the provided document context below. If the answer cannot be found in the context, say exactly: "I cannot find this information in the document." Do not use your general knowledge or make assumptions beyond the context.`
)

type Generator struct {
	client *genai.Client
}

func New(apiKey string) (*Generator, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &Generator{client: client}, nil
}

// Generate builds a grounded prompt and calls Gemini.
func (g *Generator) Generate(ctx context.Context, chunks []string, history []*store.Message, userQuery string) (string, error) {
	// Build system instruction + context block
	var sb strings.Builder
	sb.WriteString(systemPrompt)
	if len(chunks) > 0 {
		sb.WriteString("\n\nHere is the relevant context from the document:\n\n")
		for i, c := range chunks {
			fmt.Fprintf(&sb, "--- CHUNK %d ---\n%s\n\n", i+1, c)
		}
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(sb.String(), genai.RoleUser),
	}

	// Build conversation history + current query as a turn sequence
	var contents []*genai.Content
	for _, msg := range history {
		var role genai.Role
		if msg.Role == "assistant" {
			role = genai.RoleModel
		} else {
			role = genai.RoleUser
		}
		contents = append(contents, genai.NewContentFromText(msg.Content, role))
	}
	contents = append(contents, genai.NewContentFromText(userQuery, genai.RoleUser))

	result, err := g.client.Models.GenerateContent(ctx, Model, contents, config)
	if err != nil {
		return "", fmt.Errorf("gemini generate: %w", err)
	}
	return result.Text(), nil
}
