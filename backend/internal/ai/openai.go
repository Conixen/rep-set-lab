package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	openaiModel      = openai.ChatModelGPT4o
	openaiInputCost  = 0.0000025 // $2.50 per 1M input tokens
	openaiOutputCost = 0.000010  // $10.00 per 1M output tokens
)

type OpenAIProvider struct {
	client openai.Client
}

func NewOpenAI(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{client: openai.NewClient(option.WithAPIKey(apiKey))}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) GenerateWorkout(ctx context.Context, req WorkoutRequest) (WorkoutResponse, Usage, error) {
	completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openaiModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(effectiveSystemPrompt(req)),
			openai.UserMessage(buildUserPrompt(req)),
		},
		MaxTokens: openai.Int(2048),
	})
	if err != nil {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("openai: %w", err)
	}

	if len(completion.Choices) == 0 {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("openai: empty response")
	}

	var response WorkoutResponse
	if err := json.Unmarshal([]byte(cleanJSON(completion.Choices[0].Message.Content)), &response); err != nil {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("openai: parse JSON: %w", err)
	}

	usage := Usage{
		InputTokens:  int(completion.Usage.PromptTokens),
		OutputTokens: int(completion.Usage.CompletionTokens),
		CostUSD:      float64(completion.Usage.PromptTokens)*openaiInputCost + float64(completion.Usage.CompletionTokens)*openaiOutputCost,
	}
	return response, usage, nil
}
