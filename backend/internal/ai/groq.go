package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	groqBaseURL      = "https://api.groq.com/openai/v1"
	groqModel        = "llama-3.3-70b-versatile"
	groqInputCost    = 0.00000059 // $0.59 per 1M input tokens
	groqOutputCost   = 0.00000079 // $0.79 per 1M output tokens
)

type GroqProvider struct {
	client openai.Client
}

func NewGroq(apiKey string) *GroqProvider {
	return &GroqProvider{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(groqBaseURL),
		),
	}
}

func (p *GroqProvider) Name() string { return "groq" }

func (p *GroqProvider) GenerateWorkout(ctx context.Context, req WorkoutRequest) (WorkoutResponse, Usage, error) {
	completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: groqModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(buildUserPrompt(req)),
		},
		MaxTokens: openai.Int(2048),
	})
	if err != nil {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("groq: %w", err)
	}

	if len(completion.Choices) == 0 {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("groq: empty response")
	}

	raw := completion.Choices[0].Message.Content
	// Strip markdown fences defensively — some models wrap output despite instructions.
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var response WorkoutResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("groq: parse JSON: %w", err)
	}

	usage := Usage{
		InputTokens:  int(completion.Usage.PromptTokens),
		OutputTokens: int(completion.Usage.CompletionTokens),
		CostUSD:      float64(completion.Usage.PromptTokens)*groqInputCost + float64(completion.Usage.CompletionTokens)*groqOutputCost,
	}
	return response, usage, nil
}
