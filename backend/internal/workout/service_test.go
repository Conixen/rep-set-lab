package workout_test

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/leonj/rep-set-lab/internal/ai"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/mock"
	"github.com/leonj/rep-set-lab/internal/workout"
	"github.com/leonj/rep-set-lab/internal/ws"
)

func newHub() *ws.Hub {
	return ws.NewHub(slog.New(slog.NewTextHandler(os.Stderr, nil)))
}

// --- Generate ---

func TestGenerate_HappyPath(t *testing.T) {
	svc := workout.NewService(
		&mock.WorkoutStorage{},
		&mock.UserStorage{},
		map[string]ai.Provider{"mock": &mock.AIProvider{}},
		newHub(),
		nil,
	)

	result, err := svc.Generate(context.Background(), 1, workout.GenerateRequest{
		MuscleGroup:     "chest",
		DurationMinutes: 60,
		AIProvider:      "mock",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Workout.XPEarned != 600 {
		t.Errorf("XPEarned = %d, want 600 (60 min × 10)", result.Workout.XPEarned)
	}
	if result.Response.Title == "" {
		t.Error("expected non-empty workout title in response")
	}
}

func TestGenerate_UnknownProvider(t *testing.T) {
	svc := workout.NewService(
		&mock.WorkoutStorage{},
		&mock.UserStorage{},
		map[string]ai.Provider{"claude": &mock.AIProvider{}},
		newHub(),
		nil,
	)

	_, err := svc.Generate(context.Background(), 1, workout.GenerateRequest{
		MuscleGroup:     "chest",
		DurationMinutes: 60,
		AIProvider:      "openai", // not registered
	})
	if err == nil {
		t.Fatal("expected error for unregistered provider")
	}
	if !strings.Contains(err.Error(), "claude") {
		t.Errorf("error should list registered provider 'claude', got: %v", err)
	}
}

func TestGenerate_ProviderError(t *testing.T) {
	provider := &mock.AIProvider{
		GenerateWorkoutFunc: func(_ context.Context, _ ai.WorkoutRequest) (ai.WorkoutResponse, ai.Usage, error) {
			return ai.WorkoutResponse{}, ai.Usage{}, errors.New("provider unavailable")
		},
	}
	svc := workout.NewService(
		&mock.WorkoutStorage{},
		&mock.UserStorage{},
		map[string]ai.Provider{"mock": provider},
		newHub(),
		nil,
	)

	_, err := svc.Generate(context.Background(), 1, workout.GenerateRequest{
		MuscleGroup:     "back",
		DurationMinutes: 45,
		AIProvider:      "mock",
	})
	if err == nil {
		t.Fatal("expected error propagated from provider")
	}
	if !strings.Contains(err.Error(), "provider unavailable") {
		t.Errorf("expected wrapped provider error, got: %v", err)
	}
}

func TestGenerate_XPCalculation(t *testing.T) {
	tests := []struct {
		duration int
		wantXP   int
	}{
		{30, 300},
		{45, 450},
		{60, 600},
		{90, 900},
	}
	for _, tt := range tests {
		svc := workout.NewService(
			&mock.WorkoutStorage{},
			&mock.UserStorage{},
			map[string]ai.Provider{"mock": &mock.AIProvider{}},
			newHub(),
			nil,
		)
		result, err := svc.Generate(context.Background(), 1, workout.GenerateRequest{
			MuscleGroup:     "legs",
			DurationMinutes: tt.duration,
			AIProvider:      "mock",
		})
		if err != nil {
			t.Errorf("duration %d: unexpected error: %v", tt.duration, err)
			continue
		}
		if result.Workout.XPEarned != tt.wantXP {
			t.Errorf("duration %d: XPEarned = %d, want %d", tt.duration, result.Workout.XPEarned, tt.wantXP)
		}
	}
}

// --- Complete ---

func TestComplete_HappyPath(t *testing.T) {
	// Default workout mock: XPEarned=600, not completed.
	// Default user mock: XP=0, Level=1.
	// 600 XP crosses the 500 threshold → should reach level 2.
	svc := workout.NewService(&mock.WorkoutStorage{}, &mock.UserStorage{}, nil, newHub(), nil)

	result, err := svc.Complete(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.XPEarned != 600 {
		t.Errorf("XPEarned = %d, want 600", result.XPEarned)
	}
	if result.TotalXP != 600 {
		t.Errorf("TotalXP = %d, want 600", result.TotalXP)
	}
	if result.Level != 2 {
		t.Errorf("Level = %d, want 2", result.Level)
	}
	if !result.LeveledUp {
		t.Error("LeveledUp = false, want true")
	}
}

func TestComplete_NoLevelUp(t *testing.T) {
	// User at 1000 XP (level 2). 10-min workout = 100 XP → 1100 XP, still level 2 (threshold: 1500).
	workouts := &mock.WorkoutStorage{
		GetByIDFunc: func(_ context.Context, id, userID int64) (*database.Workout, error) {
			return &database.Workout{ID: id, UserID: userID, XPEarned: 100}, nil
		},
	}
	users := &mock.UserStorage{
		GetByIDFunc: func(_ context.Context, id int64) (*database.User, error) {
			return &database.User{ID: id, XP: 1000, Level: 2}, nil
		},
	}
	svc := workout.NewService(workouts, users, nil, newHub(), nil)

	result, err := svc.Complete(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LeveledUp {
		t.Error("LeveledUp = true, want false (1000 + 100 = 1100 < 1500 threshold)")
	}
	if result.Level != 2 {
		t.Errorf("Level = %d, want 2", result.Level)
	}
}

func TestComplete_AlreadyCompleted(t *testing.T) {
	workouts := &mock.WorkoutStorage{
		GetByIDFunc: func(_ context.Context, id, userID int64) (*database.Workout, error) {
			return &database.Workout{
				ID:          id,
				UserID:      userID,
				CompletedAt: sql.NullTime{Time: time.Now(), Valid: true},
			}, nil
		},
	}
	svc := workout.NewService(workouts, &mock.UserStorage{}, nil, newHub(), nil)

	_, err := svc.Complete(context.Background(), 1, 1)
	if err == nil {
		t.Fatal("expected error for already completed workout")
	}
	if !strings.Contains(err.Error(), "already completed") {
		t.Errorf("expected 'already completed' in error, got: %v", err)
	}
}

func TestComplete_WorkoutNotFound(t *testing.T) {
	workouts := &mock.WorkoutStorage{
		GetByIDFunc: func(_ context.Context, _, _ int64) (*database.Workout, error) {
			return nil, errors.New("sql: no rows in result set")
		},
	}
	svc := workout.NewService(workouts, &mock.UserStorage{}, nil, newHub(), nil)

	_, err := svc.Complete(context.Background(), 999, 1)
	if err == nil {
		t.Fatal("expected error for non-existent workout")
	}
}

func TestComplete_TransactionError(t *testing.T) {
	// The DB transaction fails — both completion and XP award should be rolled back.
	workouts := &mock.WorkoutStorage{
		CompleteAndAwardXPFunc: func(_ context.Context, _, _, _ int64, _ int) error {
			return errors.New("deadlock detected")
		},
	}
	users := &mock.UserStorage{
		GetByIDFunc: func(_ context.Context, id int64) (*database.User, error) {
			return &database.User{ID: id, XP: 0, Level: 1}, nil
		},
	}
	svc := workout.NewService(workouts, users, nil, newHub(), nil)

	_, err := svc.Complete(context.Background(), 1, 1)
	if err == nil {
		t.Fatal("expected error from failed transaction")
	}
	if !strings.Contains(err.Error(), "deadlock") {
		t.Errorf("expected wrapped transaction error, got: %v", err)
	}
}

func TestComplete_LevelUpAtBoundary(t *testing.T) {
	// User at 499 XP (level 1). A 1-min workout (10 XP) → 509 XP, just past the 500 threshold.
	workouts := &mock.WorkoutStorage{
		GetByIDFunc: func(_ context.Context, id, userID int64) (*database.Workout, error) {
			return &database.Workout{ID: id, UserID: userID, XPEarned: 10}, nil
		},
	}
	users := &mock.UserStorage{
		GetByIDFunc: func(_ context.Context, id int64) (*database.User, error) {
			return &database.User{ID: id, XP: 499, Level: 1}, nil
		},
	}
	svc := workout.NewService(workouts, users, nil, newHub(), nil)

	result, err := svc.Complete(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.LeveledUp {
		t.Error("LeveledUp = false, want true (499 + 10 = 509 crosses 500 threshold)")
	}
}
