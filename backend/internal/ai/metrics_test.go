package ai

import (
	"testing"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func workout(warmUp, main []Exercise, tips []string) WorkoutResponse {
	return WorkoutResponse{
		Title:       "Test Workout",
		Description: "desc",
		WarmUp:      warmUp,
		Main:        main,
		Tips:        tips,
	}
}

func ex(name string, sets, reps, durationSec, restSec int, notes string) Exercise {
	return Exercise{
		Name:            name,
		Sets:            sets,
		Reps:            reps,
		DurationSeconds: durationSec,
		RestSeconds:     restSec,
		Notes:           notes,
	}
}

// ── computeLibraryMatch ───────────────────────────────────────────────────────

func TestComputeLibraryMatch_AllMatch(t *testing.T) {
	lookup := map[string]bool{
		"barbell bench press": true,
		"bench press":         true,
		"flat bench":          true,
		"pull-up":             true,
	}
	resp := workout(
		[]Exercise{ex("Pull-Up", 3, 8, 0, 60, "")},
		[]Exercise{ex("Barbell Bench Press", 4, 8, 0, 90, "")},
		nil,
	)
	got := computeLibraryMatch(resp, lookup)
	if got.MatchCount != 2 {
		t.Errorf("match count: got %d, want 2", got.MatchCount)
	}
	if got.TotalCount != 2 {
		t.Errorf("total count: got %d, want 2", got.TotalCount)
	}
	if got.MatchRate != 1.0 {
		t.Errorf("match rate: got %.2f, want 1.0", got.MatchRate)
	}
}

func TestComputeLibraryMatch_NoMatch(t *testing.T) {
	lookup := map[string]bool{"barbell bench press": true}
	resp := workout(
		nil,
		[]Exercise{ex("Machine Chest Press", 3, 10, 0, 60, "")},
		nil,
	)
	got := computeLibraryMatch(resp, lookup)
	if got.MatchCount != 0 {
		t.Errorf("match count: got %d, want 0", got.MatchCount)
	}
	if got.MatchRate != 0.0 {
		t.Errorf("match rate: got %.2f, want 0.0", got.MatchRate)
	}
}

func TestComputeLibraryMatch_PartialMatch(t *testing.T) {
	lookup := map[string]bool{
		"push-up":  true,
		"push-ups": true,
	}
	resp := workout(
		nil,
		[]Exercise{
			ex("Push-Up", 3, 15, 0, 60, ""),
			ex("Handstand Push-Up", 3, 5, 0, 90, ""),
			ex("Pike Push-Up", 3, 10, 0, 60, ""),
		},
		nil,
	)
	got := computeLibraryMatch(resp, lookup)
	if got.MatchCount != 1 {
		t.Errorf("match count: got %d, want 1 (only exact 'push-up' matches)", got.MatchCount)
	}
	if got.TotalCount != 3 {
		t.Errorf("total count: got %d, want 3", got.TotalCount)
	}
}

func TestComputeLibraryMatch_EmptyResponse(t *testing.T) {
	lookup := map[string]bool{"barbell bench press": true}
	got := computeLibraryMatch(WorkoutResponse{}, lookup)
	if got.TotalCount != 0 || got.MatchRate != 0 {
		t.Errorf("empty response should return zero LibraryMatch, got %+v", got)
	}
}

// ── computeBehavioralMetrics — section counts & completeness ─────────────────

func TestBehavioralMetrics_Completeness(t *testing.T) {
	resp := workout(
		[]Exercise{ex("Jumping Jacks", 0, 0, 60, 0, "")},
		[]Exercise{ex("Push-Up", 3, 15, 0, 60, "")},
		[]string{"Stay hydrated"},
	)
	got := computeBehavioralMetrics(resp, "gym")
	if got.CompletenessScore != 3 {
		t.Errorf("completeness: got %d, want 3", got.CompletenessScore)
	}
	if got.WarmUpCount != 1 || got.MainCount != 1 || got.TipsCount != 1 {
		t.Errorf("section counts wrong: %+v", got)
	}
}

func TestBehavioralMetrics_PartialCompleteness(t *testing.T) {
	resp := workout(nil, []Exercise{ex("Push-Up", 3, 15, 0, 60, "")}, nil)
	got := computeBehavioralMetrics(resp, "gym")
	if got.CompletenessScore != 1 {
		t.Errorf("completeness: got %d, want 1 (main only)", got.CompletenessScore)
	}
}

// ── computeBehavioralMetrics — emoji counting ─────────────────────────────────

func TestBehavioralMetrics_EmojiCount(t *testing.T) {
	resp := WorkoutResponse{
		Title:       "💪 Chest Day 🔥",
		Description: "Let's go!",
		Main: []Exercise{
			{Name: "Push-Up 🏋️", Notes: "Keep your form"},
			{Name: "Dip", Notes: ""},
		},
		Tips: []string{"Stay focused 💯"},
	}
	got := computeBehavioralMetrics(resp, "gym")
	// 🏋️ is two code points (U+1F3CB + U+FE0F variation selector), so total = 5.
	if got.EmojiCount != 5 {
		t.Errorf("emoji count: got %d, want 5", got.EmojiCount)
	}
}

func TestBehavioralMetrics_NoEmoji(t *testing.T) {
	resp := workout(nil, []Exercise{ex("Push-Up", 3, 15, 0, 60, "Keep core tight")}, nil)
	got := computeBehavioralMetrics(resp, "gym")
	if got.EmojiCount != 0 {
		t.Errorf("emoji count: got %d, want 0", got.EmojiCount)
	}
}

// ── computeBehavioralMetrics — equipment violations ───────────────────────────

func TestBehavioralMetrics_EquipmentViolations_Home(t *testing.T) {
	resp := workout(nil, []Exercise{
		ex("Barbell Squat", 4, 8, 0, 120, ""),    // violation
		ex("Push-Up", 3, 15, 0, 60, ""),           // fine
		ex("Cable Crossover", 3, 12, 0, 60, ""),   // violation
	}, nil)
	got := computeBehavioralMetrics(resp, "home")
	if got.EquipmentViolations != 2 {
		t.Errorf("equipment violations: got %d, want 2", got.EquipmentViolations)
	}
}

func TestBehavioralMetrics_EquipmentViolations_Gym(t *testing.T) {
	resp := workout(nil, []Exercise{
		ex("Barbell Squat", 4, 8, 0, 120, ""),
		ex("Cable Crossover", 3, 12, 0, 60, ""),
	}, nil)
	got := computeBehavioralMetrics(resp, "gym")
	if got.EquipmentViolations != 0 {
		t.Errorf("gym environment should have 0 violations, got %d", got.EquipmentViolations)
	}
}

func TestBehavioralMetrics_EquipmentViolations_Outdoor(t *testing.T) {
	resp := workout(nil, []Exercise{
		ex("Dumbbell Curl", 3, 12, 0, 60, ""), // violation outdoors
		ex("Push-Up", 3, 15, 0, 60, ""),       // fine
		ex("Sprint", 5, 0, 30, 30, ""),        // fine
	}, nil)
	got := computeBehavioralMetrics(resp, "outdoor")
	if got.EquipmentViolations != 1 {
		t.Errorf("equipment violations: got %d, want 1", got.EquipmentViolations)
	}
}

// ── computeBehavioralMetrics — note metrics ───────────────────────────────────

func TestBehavioralMetrics_NoteMetrics_AllPresent(t *testing.T) {
	resp := workout(nil, []Exercise{
		ex("Push-Up", 3, 15, 0, 60, "Keep core tight"),
		ex("Dip", 3, 10, 0, 60, "Lean forward slightly"),
	}, nil)
	got := computeBehavioralMetrics(resp, "gym")
	if got.NotesPresentRate != 1.0 {
		t.Errorf("notes present rate: got %.2f, want 1.0", got.NotesPresentRate)
	}
	if got.AvgNoteLength == 0 {
		t.Error("avg note length should be > 0 when all notes present")
	}
}

func TestBehavioralMetrics_NoteMetrics_NonePresent(t *testing.T) {
	resp := workout(nil, []Exercise{
		ex("Push-Up", 3, 15, 0, 60, ""),
		ex("Dip", 3, 10, 0, 60, ""),
	}, nil)
	got := computeBehavioralMetrics(resp, "gym")
	if got.NotesPresentRate != 0.0 {
		t.Errorf("notes present rate: got %.2f, want 0.0", got.NotesPresentRate)
	}
	if got.AvgNoteLength != 0 {
		t.Errorf("avg note length: got %.2f, want 0", got.AvgNoteLength)
	}
}

func TestBehavioralMetrics_NoteMetrics_HalfPresent(t *testing.T) {
	resp := workout(nil, []Exercise{
		ex("Push-Up", 3, 15, 0, 60, "Keep core tight"),
		ex("Dip", 3, 10, 0, 60, ""),
	}, nil)
	got := computeBehavioralMetrics(resp, "gym")
	if got.NotesPresentRate != 0.5 {
		t.Errorf("notes present rate: got %.2f, want 0.5", got.NotesPresentRate)
	}
}

// ── estimateMinutes ───────────────────────────────────────────────────────────

func TestEstimateMinutes_SetsBased(t *testing.T) {
	// 3 sets × 10 reps × 3s = 90s + 60s rest = 150s = 2.5 min
	main := []Exercise{ex("Push-Up", 3, 10, 0, 60, "")}
	got := estimateMinutes(main)
	want := 2.5
	if got != want {
		t.Errorf("estimated minutes: got %.2f, want %.2f", got, want)
	}
}

func TestEstimateMinutes_DurationBased(t *testing.T) {
	// duration_seconds=60 + rest=30 = 90s = 1.5 min
	main := []Exercise{ex("Plank", 0, 0, 60, 30, "")}
	got := estimateMinutes(main)
	want := 1.5
	if got != want {
		t.Errorf("estimated minutes: got %.2f, want %.2f", got, want)
	}
}

func TestEstimateMinutes_NoRest(t *testing.T) {
	// 3 × 10 × 3s = 90s = 1.5 min, no rest
	main := []Exercise{ex("Push-Up", 3, 10, 0, 0, "")}
	got := estimateMinutes(main)
	want := 1.5
	if got != want {
		t.Errorf("estimated minutes: got %.2f, want %.2f", got, want)
	}
}

func TestEstimateMinutes_Empty(t *testing.T) {
	got := estimateMinutes(nil)
	if got != 0 {
		t.Errorf("empty main should give 0 minutes, got %.2f", got)
	}
}

// ── countEmojis ───────────────────────────────────────────────────────────────

func TestCountEmojis(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"Hello world", 0},
		{"💪 Let's go 🔥", 2},
		{"No emoji here!", 0},
		{"🏋️🏃💯", 4}, // 🏋️ = U+1F3CB + U+FE0F (two code points)
		{"Mixed 💪 text 🔥 here", 2},
	}
	for _, tc := range cases {
		got := countEmojis(tc.input)
		if got != tc.want {
			t.Errorf("countEmojis(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}
