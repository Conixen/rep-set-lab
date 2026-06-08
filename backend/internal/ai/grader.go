package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/leonj/rep-set-lab/internal/database"
)

// GradeResult holds Groq's objective grading of a generated workout.
type GradeResult struct {
	InjuryGrade       string `json:"injury_grade"`
	InjuryFeedback    string `json:"injury_feedback"`
	EquipmentGrade    string `json:"equipment_grade"`
	EquipmentFeedback string `json:"equipment_feedback"`
	GoalGrade         string `json:"goal_grade"`
	GoalFeedback      string `json:"goal_feedback"`
}

// Grader scores a generated workout against the original request.
type Grader interface {
	GradeWorkout(ctx context.Context, req WorkoutRequest, resp WorkoutResponse) (*GradeResult, error)
}

const graderModel = "llama-3.3-70b-versatile"

const gradeSystemPrompt = `You are an objective fitness evaluator. Grade a workout plan on three dimensions using A–F.

Return only valid JSON with this exact structure:
{
  "injury_grade": "A",
  "injury_feedback": "one sentence",
  "equipment_grade": "A",
  "equipment_feedback": "one sentence",
  "goal_grade": "A",
  "goal_feedback": "one sentence"
}

Grading scale: A=Excellent, B=Good, C=Average, D=Poor, F=Fail.

Criteria:
- injury_grade: does the workout avoid exercises that could aggravate stated injuries? Grade A if no injuries stated.
- equipment_grade: do all exercises respect the environment's equipment constraints? Grade A if environment is gym.
- goal_grade: do sets, reps, and overall structure match the stated goal?`

// GroqGrader uses Groq/LLaMA as a neutral, fast, third-party workout evaluator.
type GroqGrader struct {
	client openai.Client
}

func NewGroqGrader(apiKey string) *GroqGrader {
	return &GroqGrader{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(groqBaseURL),
		),
	}
}

func (g *GroqGrader) GradeWorkout(ctx context.Context, req WorkoutRequest, resp WorkoutResponse) (*GradeResult, error) {
	completion, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: graderModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(gradeSystemPrompt),
			openai.UserMessage(buildGradePrompt(req, resp)),
		},
		MaxTokens: openai.Int(512),
	})
	if err != nil {
		return nil, fmt.Errorf("groq grader: %w", err)
	}
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("groq grader: empty response")
	}

	raw := cleanJSON(completion.Choices[0].Message.Content)

	var result GradeResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("groq grader: parse JSON: %w", err)
	}
	return &result, nil
}

// SessionAnalysis is the result of AnalyzeCompare — a written report across all recorded sessions.
type SessionAnalysis struct {
	Narrative    string           `json:"narrative"`
	Verdicts     []ProviderVerdict `json:"verdicts"`
	SessionCount int              `json:"session_count"`
}

// ProviderVerdict is Groq's overall assessment of one provider across all sessions.
type ProviderVerdict struct {
	Provider string `json:"provider"`
	Grade    string `json:"grade"`
	Summary  string `json:"summary"`
}

// AnalyzeCompare sends aggregated provider metrics to Groq and returns a written comparative analysis.
func (g *GroqGrader) AnalyzeCompare(ctx context.Context, avgs []*database.ProviderCompareAvg, sessionCount int) (*SessionAnalysis, error) {
	completion, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: graderModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are an AI research analyst. Return only valid JSON — no markdown fences, no commentary outside the JSON object."),
			openai.UserMessage(buildAnalyzePrompt(avgs, sessionCount)),
		},
		MaxTokens: openai.Int(1024),
	})
	if err != nil {
		return nil, fmt.Errorf("groq analyze: %w", err)
	}
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("groq analyze: empty response")
	}
	raw := cleanJSON(completion.Choices[0].Message.Content)

	var result SessionAnalysis
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("groq analyze: parse JSON: %w", err)
	}
	result.SessionCount = sessionCount
	for i := range result.Verdicts {
		result.Verdicts[i].Grade = normalizeGrade(result.Verdicts[i].Grade)
	}
	return &result, nil
}

func buildAnalyzePrompt(avgs []*database.ProviderCompareAvg, sessionCount int) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "You are analyzing workout generation performance across %d compare sessions.\n", sessionCount)
	if sessionCount < 15 {
		sb.WriteString("Sample size is small — treat findings as directional, not conclusive.\n")
	}
	sb.WriteString(`
Focus your analysis on these dimensions:
1. Workout structure balance — warm-up vs main exercise counts relative to the requested duration. Does the split feel appropriate?
2. Time accuracy — how close is the estimated workout duration to what was actually requested?
3. Safety & compliance — injury compliance score and equipment violations. Are providers respecting constraints?
4. Goal alignment — are the exercises and structure actually targeting the requested muscle group?
5. Instruction quality — notes present rate and response length. Are exercises explained well enough for a user to follow?
6. Cost & speed efficiency — tokens used, cost per run, latency. What does each provider cost in practice?

Provider aggregate metrics:

`)
	for _, a := range avgs {
		fmt.Fprintf(&sb, "Provider: %s (%d sessions)\n", a.Provider, a.TotalSessions)
		fmt.Fprintf(&sb, "  Warm-up exercises (avg):   %.1f\n", a.AvgWarmUpCount)
		fmt.Fprintf(&sb, "  Main exercises (avg):      %.1f\n", a.AvgMainCount)
		fmt.Fprintf(&sb, "  Tips included (avg):       %.1f\n", a.AvgTipsCount)
		fmt.Fprintf(&sb, "  Estimated duration:        %.0f min\n", a.AvgEstimatedMinutes)
		fmt.Fprintf(&sb, "  Structure completeness:    %.1f/3\n", a.AvgCompletenessScore)
		fmt.Fprintf(&sb, "  Library match rate:        %.0f%%\n", a.AvgLibraryMatchRate*100)
		fmt.Fprintf(&sb, "  Exercises with notes:      %.0f%%\n", a.AvgNotesPresentRate*100)
		fmt.Fprintf(&sb, "  Avg response length:       %.0f chars\n", a.AvgCharCount)
		fmt.Fprintf(&sb, "  Equipment violations (avg):%.1f\n", a.AvgEquipmentViolations)
		fmt.Fprintf(&sb, "  Emoji count (avg):         %.1f\n", a.AvgEmojiCount)
		fmt.Fprintf(&sb, "  Avg input tokens:          %.0f\n", a.AvgInputTokens)
		fmt.Fprintf(&sb, "  Avg output tokens:         %.0f\n", a.AvgOutputTokens)
		fmt.Fprintf(&sb, "  Avg cost per run (USD):    $%.4f\n", a.AvgCostUSD)
		fmt.Fprintf(&sb, "  Avg latency:               %.0f ms\n", a.AvgLatencyMs)
		if a.AvgGroqInjuryScore > 0 {
			fmt.Fprintf(&sb, "  Injury compliance:         %.1f/5\n", a.AvgGroqInjuryScore)
			fmt.Fprintf(&sb, "  Equipment compliance:      %.1f/5\n", a.AvgGroqEquipmentScore)
			fmt.Fprintf(&sb, "  Goal alignment:            %.1f/5\n", a.AvgGroqGoalScore)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("Return only valid JSON:\n")
	sb.WriteString(`{"narrative":"3-4 paragraph analysis covering structure balance, safety/compliance, instruction quality, and cost-efficiency. Be specific — reference actual numbers. Note where sample size limits confidence.","verdicts":[{"provider":"name","grade":"A","summary":"one sentence covering the provider's defining strength and weakness"}]}`)
	sb.WriteString("\nGrade A=Excellent B=Good C=Average D=Poor F=Fail. No E grade.")
	return sb.String()
}

// returns {A,B,C,D,F}, or "" for anything else.
func normalizeGrade(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	c := strings.ToUpper(string(s[0]))
	switch c {
	case "A", "B", "C", "D", "F":
		return c
	}
	return ""
}

func buildGradePrompt(req WorkoutRequest, resp WorkoutResponse) string {
	var sb strings.Builder
	sb.WriteString("=== ORIGINAL REQUEST ===\n")
	fmt.Fprintf(&sb, "Muscle group: %s\n", req.MuscleGroup)
	fmt.Fprintf(&sb, "Duration: %d minutes\n", req.DurationMinutes)
	fmt.Fprintf(&sb, "Environment: %s\n", environmentDescription(req.Environment))
	if req.Injuries != "" {
		fmt.Fprintf(&sb, "Injuries/limitations: %s\n", req.Injuries)
	}
	if req.Goals != "" {
		fmt.Fprintf(&sb, "Goals: %s\n", req.Goals)
	}

	sb.WriteString("\n=== GENERATED WORKOUT ===\n")
	fmt.Fprintf(&sb, "Title: %s\n", resp.Title)

	sb.WriteString("Warm-up:\n")
	for _, ex := range resp.WarmUp {
		fmt.Fprintf(&sb, "  - %s", ex.Name)
		if ex.Sets > 0 && ex.Reps > 0 {
			fmt.Fprintf(&sb, " %dx%d", ex.Sets, ex.Reps)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Main:\n")
	for _, ex := range resp.Main {
		fmt.Fprintf(&sb, "  - %s", ex.Name)
		if ex.Sets > 0 && ex.Reps > 0 {
			fmt.Fprintf(&sb, " %dx%d", ex.Sets, ex.Reps)
		}
		if ex.RestSeconds > 0 {
			fmt.Fprintf(&sb, ", %ds rest", ex.RestSeconds)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
