package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	geminiInputCost  = 0.00000125
	geminiOutputCost = 0.000005
)

type GeminiProvider struct {
	client *genai.Client
	model  string
	name   string
}

// NewGemini creates a Gemini provider for the given model name.
func NewGemini(apiKey, model, providerName string) (*GeminiProvider, error) {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini: create client: %w", err)
	}
	return &GeminiProvider{client: client, model: model, name: providerName}, nil
}

func (p *GeminiProvider) Name() string { return p.name }

func (p *GeminiProvider) GenerateWorkout(ctx context.Context, req WorkoutRequest) (WorkoutResponse, Usage, error) {
	model := p.client.GenerativeModel(p.model)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(effectiveSystemPrompt(req))},
	}

	resp, err := model.GenerateContent(ctx, genai.Text(buildUserPrompt(req)))
	if err != nil {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("gemini: %w", err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("gemini: empty response")
	}

	raw := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	// Strip markdown code fences if the model wraps output despite instructions.
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var response WorkoutResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("gemini: parse JSON: %w", err)
	}

	var inputTokens, outputTokens int
	if resp.UsageMetadata != nil {
		inputTokens = int(resp.UsageMetadata.PromptTokenCount)
		outputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	usage := Usage{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		CostUSD:      float64(inputTokens)*geminiInputCost + float64(outputTokens)*geminiOutputCost,
	}
	return response, usage, nil
}
