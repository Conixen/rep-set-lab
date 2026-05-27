// test-provider is a standalone diagnostic tool that verifies each configured
// AI provider key works, using a dead-simple prompt that completely bypasses
// the workout template, system prompt, and JSON schema.
//
// Anthropic and OpenAI sections are commented out by default to avoid burning
// tokens on every run. Uncomment the relevant section (imports + main block +
// helper function) when you need to verify one of those keys.
//
// Usage:
//
//	cd backend && go run ./cmd/test-provider
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"

	// ── Uncomment to test Anthropic ───────────────────────────────────────────
	// "github.com/anthropics/anthropic-sdk-go"
	// anthropicoption "github.com/anthropics/anthropic-sdk-go/option"

	// ── Uncomment to test OpenAI ──────────────────────────────────────────────
	// "github.com/openai/openai-go"
	// openaioption "github.com/openai/openai-go/option"
)

const testPrompt = "Say hello in exactly one word."

func main() {
	_ = godotenv.Load() // loads backend/.env when run from backend/

	ctx := context.Background()
	passed, total := 0, 0

	fmt.Println("=== AI Provider Key Test ===")
	fmt.Println()

	// ── Gemini ────────────────────────────────────────────────────────────────
	if key := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); key != "" {
		fmt.Printf("Gemini  key: %s\n", maskKey(key))
		client, err := genai.NewClient(ctx, option.WithAPIKey(key))
		if err != nil {
			fmt.Printf("  ✗  failed to create client: %v\n\n", err)
		} else {
			defer client.Close()
			for _, modelName := range []string{
				"gemini-2.5-flash",
				"gemini-2.5-pro",
			} {
				total++
				ok, ms, text, callErr := testGemini(ctx, client, modelName)
				printResult(modelName, ok, ms, text, callErr)
				if ok {
					passed++
				}
			}
		}
		fmt.Println()
	} else {
		fmt.Println("Gemini  — skipped (GEMINI_API_KEY not set)")
		fmt.Println()
	}

	// ── Anthropic ─────────────────────────────────────────────────────────────
	// Commented out — don't want to burn tokens on every test run.
	// Uncomment this block AND the anthropic imports above to test the key.
	/*
		if key := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")); key != "" {
			fmt.Printf("Anthropic  key: %s\n", maskKey(key))
			total++
			ok, ms, text, callErr := testAnthropic(ctx, key)
			printResult("claude-haiku-4-5", ok, ms, text, callErr)
			if ok {
				passed++
			}
			fmt.Println()
		} else {
			fmt.Println("Anthropic  — skipped (ANTHROPIC_API_KEY not set)")
			fmt.Println()
		}
	*/

	// ── OpenAI ────────────────────────────────────────────────────────────────
	// Commented out — don't want to burn tokens on every test run.
	// Uncomment this block AND the openai imports above to test the key.
	/*
		if key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); key != "" {
			fmt.Printf("OpenAI  key: %s\n", maskKey(key))
			total++
			ok, ms, text, callErr := testOpenAI(ctx, key)
			printResult("gpt-4o-mini", ok, ms, text, callErr)
			if ok {
				passed++
			}
			fmt.Println()
		} else {
			fmt.Println("OpenAI  — skipped (OPENAI_API_KEY not set)")
			fmt.Println()
		}
	*/

	if total == 0 {
		fmt.Fprintln(os.Stderr, "No provider keys found in .env — nothing to test.")
		os.Exit(1)
	}

	fmt.Printf("Results: %d/%d passed\n", passed, total)
	if passed < total {
		fmt.Println("\nIf all calls to a provider fail → billing not active, wrong key, or account suspended.")
		fmt.Println("If only some Gemini models fail → those variants may not be enabled on your project.")
		os.Exit(1)
	}
}

// ── per-provider test helpers ─────────────────────────────────────────────────

func testGemini(ctx context.Context, client *genai.Client, modelName string) (ok bool, ms int64, text string, err error) {
	start := time.Now()
	model := client.GenerativeModel(modelName)
	resp, callErr := model.GenerateContent(ctx, genai.Text(testPrompt))
	ms = time.Since(start).Milliseconds()
	if callErr != nil {
		return false, ms, "", callErr
	}
	if len(resp.Candidates) == 0 ||
		resp.Candidates[0].Content == nil ||
		len(resp.Candidates[0].Content.Parts) == 0 {
		return false, ms, "", fmt.Errorf("empty response")
	}
	text = strings.TrimSpace(fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]))
	return true, ms, text, nil
}

// ── Anthropic helper ──────────────────────────────────────────────────────────
// Commented out — uncomment together with the Anthropic imports and main block.
/*
func testAnthropic(ctx context.Context, apiKey string) (ok bool, ms int64, text string, err error) {
	client := anthropic.NewClient(anthropicoption.WithAPIKey(apiKey))
	start := time.Now()
	msg, callErr := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model("claude-haiku-4-5-20251001"),
		MaxTokens: 10,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(testPrompt)),
		},
	})
	ms = time.Since(start).Milliseconds()
	if callErr != nil {
		return false, ms, "", callErr
	}
	if len(msg.Content) == 0 {
		return false, ms, "", fmt.Errorf("empty response")
	}
	text = strings.TrimSpace(msg.Content[0].Text)
	return true, ms, text, nil
}
*/

// ── OpenAI helper ─────────────────────────────────────────────────────────────
// Commented out — uncomment together with the OpenAI imports and main block.
/*
func testOpenAI(ctx context.Context, apiKey string) (ok bool, ms int64, text string, err error) {
	client := openai.NewClient(openaioption.WithAPIKey(apiKey))
	start := time.Now()
	completion, callErr := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(testPrompt),
		},
		MaxTokens: openai.Int(10),
	})
	ms = time.Since(start).Milliseconds()
	if callErr != nil {
		return false, ms, "", callErr
	}
	if len(completion.Choices) == 0 {
		return false, ms, "", fmt.Errorf("empty response")
	}
	text = strings.TrimSpace(completion.Choices[0].Message.Content)
	return true, ms, text, nil
}
*/

// ── shared helpers ────────────────────────────────────────────────────────────

func printResult(label string, ok bool, ms int64, text string, err error) {
	if ok {
		if len(text) > 60 {
			text = text[:57] + "..."
		}
		fmt.Printf("  ✓  %-30s OK    (%dms)  %q\n", label, ms, text)
	} else {
		fmt.Printf("  ✗  %-30s FAIL  (%dms)  %v\n", label, ms, err)
	}
}

// maskKey shows the first 4 and last 4 characters with stars in between
// so you can confirm it's the right key without exposing it.
func maskKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}
