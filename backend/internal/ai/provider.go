package ai

import (
	"context"
	"fmt"
	"strings"
)

type WorkoutRequest struct {
	UserPrompt           string
	MuscleGroup          string
	DurationMinutes      int
	Injuries             string
	Goals                string
	Environment          string
	Language             string   // "sv" = Swedish output; exercise names always stay in English
	AvailableExercises   []string // exercise names from our library; AI prefers these for GIF previews
	SystemPromptOverride string   // if set, replaces systemPrompt (used by compare to expose natural model behavior)
}

// effectiveSystemPrompt returns the override if set, otherwise the default structured prompt.
func effectiveSystemPrompt(req WorkoutRequest) string {
	if req.SystemPromptOverride != "" {
		return req.SystemPromptOverride
	}
	return systemPrompt
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

// compareSystemPrompt is intentionally looser than systemPrompt so models reveal
// natural differences in verbosity, emoji use, injury handling, and equipment awareness.
const compareSystemPrompt = `You are a personal trainer. Generate a workout plan for the user.

Return only valid JSON with these fields: title (string), description (string), warm_up (array of exercises), main (array of exercises), tips (array of strings).

Each exercise object contains: name (string), sets (integer), reps (integer), duration_seconds (integer), rest_seconds (integer), notes (string).`

const systemPrompt = `You are an expert personal trainer. Generate a structured workout appropriate for the user's training environment and input.

Always respond with valid JSON matching this exact structure:
{
  "title": "string",
  "description": "string",
  "warm_up": [{"name":"string","sets":0,"reps":0,"duration_seconds":0,"rest_seconds":0,"notes":"string"}],
  "main":    [{"name":"string","sets":0,"reps":0,"duration_seconds":0,"rest_seconds":0,"notes":"string"}],
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

// ── Metric types ─────────────────────────────────────────────────────────────

type BehavioralMetrics struct {
	CharCount           int     `json:"char_count"`
	EmojiCount          int     `json:"emoji_count"`
	EquipmentViolations int     `json:"equipment_violations"`
	CompletenessScore   int     `json:"completeness_score"`
	WarmUpCount         int     `json:"warm_up_count"`
	MainCount           int     `json:"main_count"`
	TipsCount           int     `json:"tips_count"`
	AvgNoteLength       float64 `json:"avg_note_length"`
	NotesPresentRate    float64 `json:"notes_present_rate"`
	EstimatedMinutes    float64 `json:"estimated_minutes"`
}

type LibraryMatch struct {
	MatchRate  float64 `json:"match_rate"`
	MatchCount int     `json:"match_count"`
	TotalCount int     `json:"total_count"`
}

// ── Metric helpers ────────────────────────────────────────────────────────────

var homeViolationKeywords    = []string{"barbell", "cable", "smith machine", "leg press"}
var outdoorViolationKeywords = []string{"barbell", "cable", "dumbbell", "machine"}

const secondsPerRep = 3

func computeBehavioralMetrics(resp WorkoutResponse, environment string) BehavioralMetrics {
	all := collectExercises(resp)
	text := collectAllText(resp)

	completeness := 0
	if len(resp.WarmUp) > 0 { completeness++ }
	if len(resp.Main) > 0   { completeness++ }
	if len(resp.Tips) > 0   { completeness++ }

	avgNoteLen, notesPresentRate := computeNoteMetrics(all)

	return BehavioralMetrics{
		CharCount:           len(text),
		EmojiCount:          countEmojis(text),
		EquipmentViolations: countEquipmentViolations(all, environment),
		CompletenessScore:   completeness,
		WarmUpCount:         len(resp.WarmUp),
		MainCount:           len(resp.Main),
		TipsCount:           len(resp.Tips),
		AvgNoteLength:       avgNoteLen,
		NotesPresentRate:    notesPresentRate,
		EstimatedMinutes:    estimateMinutes(resp.Main),
	}
}

// computeLibraryMatch takes a pre-built lowercase lookup set of seed names + aliases.
func computeLibraryMatch(resp WorkoutResponse, lookup map[string]bool) LibraryMatch {
	all := collectExercises(resp)
	if len(all) == 0 {
		return LibraryMatch{}
	}
	matched := 0
	for _, ex := range all {
		if lookup[strings.ToLower(strings.TrimSpace(ex.Name))] {
			matched++
		}
	}
	return LibraryMatch{
		MatchRate:  float64(matched) / float64(len(all)),
		MatchCount: matched,
		TotalCount: len(all),
	}
}

func collectExercises(resp WorkoutResponse) []Exercise {
	out := make([]Exercise, 0, len(resp.WarmUp)+len(resp.Main))
	out = append(out, resp.WarmUp...)
	out = append(out, resp.Main...)
	return out
}

// collectAllText concatenates all human-readable text from a response into one string.
// Used for char count and emoji count so both measures share a single allocation.
func collectAllText(resp WorkoutResponse) string {
	var sb strings.Builder
	sb.WriteString(resp.Title)
	sb.WriteString(resp.Description)
	for _, ex := range collectExercises(resp) {
		sb.WriteString(ex.Name)
		sb.WriteString(ex.Notes)
	}
	for _, tip := range resp.Tips {
		sb.WriteString(tip)
	}
	return sb.String()
}

func countEmojis(s string) int {
	n := 0
	for _, r := range s {
		if isEmojiRune(r) {
			n++
		}
	}
	return n
}

func isEmojiRune(r rune) bool {
	return (r >= 0x1F300 && r <= 0x1F9FF) ||
		(r >= 0x2600 && r <= 0x27BF) ||
		(r >= 0xFE00 && r <= 0xFE0F) ||
		(r >= 0x1FA00 && r <= 0x1FA9F)
}

func countEquipmentViolations(exercises []Exercise, environment string) int {
	var keywords []string
	switch environment {
	case "home":
		keywords = homeViolationKeywords
	case "outdoor":
		keywords = outdoorViolationKeywords
	default:
		return 0
	}
	count := 0
	for _, ex := range exercises {
		name := strings.ToLower(ex.Name)
		for _, kw := range keywords {
			if strings.Contains(name, kw) {
				count++
				break
			}
		}
	}
	return count
}

func computeNoteMetrics(exercises []Exercise) (avgLen float64, presentRate float64) {
	if len(exercises) == 0 {
		return 0, 0
	}
	totalLen, withNotes := 0, 0
	for _, ex := range exercises {
		note := strings.TrimSpace(ex.Notes)
		totalLen += len(note)
		if note != "" {
			withNotes++
		}
	}
	return float64(totalLen) / float64(len(exercises)),
		float64(withNotes) / float64(len(exercises))
}

func estimateMinutes(main []Exercise) float64 {
	total := 0
	for _, ex := range main {
		var t int
		if ex.DurationSeconds > 0 {
			t = ex.DurationSeconds
		} else {
			t = ex.Sets * ex.Reps * secondsPerRep
		}
		total += t + ex.RestSeconds
	}
	return float64(total) / 60.0
}

// cleanJSON strips markdown fences, extracts the first {...} block, and
// escapes literal newlines inside JSON string values so json.Unmarshal accepts it.
func cleanJSON(raw string) string {
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	if start := strings.Index(raw, "{"); start != -1 {
		if end := strings.LastIndex(raw, "}"); end > start {
			raw = raw[start : end+1]
		}
	}

	var b strings.Builder
	inString, escaped := false, false
	for _, r := range raw {
		switch {
		case escaped:
			b.WriteRune(r)
			escaped = false
		case r == '\\' && inString:
			b.WriteRune(r)
			escaped = true
		case r == '"':
			inString = !inString
			b.WriteRune(r)
		case inString && r == '\n':
			b.WriteString(`\n`)
		case inString && r == '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ── Prompt builders ───────────────────────────────────────────────────────────

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
	if len(req.AvailableExercises) > 0 {
		prompt += "\nExercises in our library (prefer these where they fit — users get image previews for library exercises):\n"
		for _, name := range req.AvailableExercises {
			prompt += "- " + name + "\n"
		}
	}
	if req.Language == "sv" {
		prompt += "\nIMPORTANT: Write the title, description, exercise notes, and tips in Swedish. Keep all exercise names in English.\n"
	}
	return prompt
}
