package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	claudeModel      = "claude-sonnet-4-6"
	claudeInputCost  = 0.000003  // $3.00 per 1M input tokens
	claudeOutputCost = 0.000015  // $15.00 per 1M output tokens
)

type ClaudeProvider struct {
	client anthropic.Client
}

func NewClaude(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{client: anthropic.NewClient(option.WithAPIKey(apiKey))}
}

func (p *ClaudeProvider) Name() string { return "claude" }

func (p *ClaudeProvider) GenerateWorkout(ctx context.Context, req WorkoutRequest) (WorkoutResponse, Usage, error) {
	msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(claudeModel),
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: effectiveSystemPrompt(req)},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(buildUserPrompt(req))),
		},
	})
	if err != nil {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("claude: %w", err)
	}

	if len(msg.Content) == 0 {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("claude: empty response")
	}

	var response WorkoutResponse
	if err := json.Unmarshal([]byte(msg.Content[0].Text), &response); err != nil {
		return WorkoutResponse{}, Usage{}, fmt.Errorf("claude: parse JSON: %w", err)
	}

	usage := Usage{
		InputTokens:  int(msg.Usage.InputTokens),
		OutputTokens: int(msg.Usage.OutputTokens),
		CostUSD:      float64(msg.Usage.InputTokens)*claudeInputCost + float64(msg.Usage.OutputTokens)*claudeOutputCost,
	}
	return response, usage, nil
}
