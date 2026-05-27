package ai

import (
	"context"
	"fmt"
)

type WorkoutRequest struct {
	UserPrompt      string
	MuscleGroup     string
	DurationMinutes int
	Injuries        string
	Goals           string
	Environment     string
}

type Exercise struct {
	Name            string `json:"name"`
	Sets            int    `json:"sets,omitempty"`
	Reps            int    `json:"reps,omitempty"`
	DurationSeconds int    `json:"duration_seconds,omitempty"`
	RestSeconds     int    `json:"rest_seconds,omitempty"`
	Notes           string `json:"notes,omitempty"`
}

type WorkoutResponse struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	WarmUp      []Exercise `json:"warm_up"`
	Main        []Exercise `json:"main"`
	CoolDown    []Exercise `json:"cool_down"`
	Tips        []string   `json:"tips"`
}

type Usage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CostUSD      float64 `json:"cost_usd"`
}

type Provider interface {
	Name() string
	GenerateWorkout(ctx context.Context, req WorkoutRequest) (WorkoutResponse, Usage, error)
}

const systemPrompt = `You are an expert personal trainer. Generate a structured workout appropriate for the user's training environment and input.

Always respond with valid JSON matching this exact structure:
{
  "title": "string",
  "description": "string",
  "warm_up": [{"name":"string","sets":0,"reps":0,"duration_seconds":0,"rest_seconds":0,"notes":"string"}],
  "main":    [{"name":"string","sets":0,"reps":0,"duration_seconds":0,"rest_seconds":0,"notes":"string"}],
  "cool_down":[{"name":"string","duration_seconds":0,"notes":"string"}],
  "tips": ["string"]
}

Rules:
- Respect all injuries — never suggest exercises that could aggravate them
- Fit the workout within the time limit provided
- Use 0 for fields that don't apply (e.g. reps for timed exercises)
- Respond ONLY with the JSON object, no markdown fences, no explanation`

// environmentDescription maps the environment key to a human-readable description
// that gives the AI enough context to select appropriate exercises.
func environmentDescription(env string) string {
	switch env {
	case "home":
		return "home (bodyweight and minimal equipment only — no barbells, cables, or gym machines)"
	case "outdoor":
		return "outdoor (no equipment — use bodyweight, open space, and natural terrain)"
	default:
		return "gym (full equipment available — barbells, cables, machines, dumbbells)"
	}
}

func buildUserPrompt(req WorkoutRequest) string {
	prompt := fmt.Sprintf("Muscle group: %s\nDuration: %d minutes\n", req.MuscleGroup, req.DurationMinutes)
	prompt += fmt.Sprintf("Training environment: %s\n", environmentDescription(req.Environment))
	if req.Injuries != "" {
		prompt += fmt.Sprintf("Injuries/limitations: %s\n", req.Injuries)
	}
	if req.Goals != "" {
		prompt += fmt.Sprintf("Goals: %s\n", req.Goals)
	}
	if req.UserPrompt != "" {
		prompt += fmt.Sprintf("Additional notes: %s\n", req.UserPrompt)
	}
	return prompt
}
