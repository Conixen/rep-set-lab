package workout

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/leonj/rep-set-lab/internal/ai"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/ws"
	"github.com/leonj/rep-set-lab/internal/xp"
)

var (
	ErrNotFound         = errors.New("workout not found")
	ErrAlreadyCompleted = errors.New("workout already completed")
	ErrUnknownProvider  = errors.New("unknown provider")
)

// Storage and UserStorage are defined here so the service depends on interfaces,
// not concrete types — makes it testable without a real database.

type Storage interface {
	Create(ctx context.Context, w *database.Workout) (*database.Workout, error)
	GetByID(ctx context.Context, id, userID int64) (*database.Workout, error)
	ListByUser(ctx context.Context, userID int64) ([]*database.Workout, error)
	// CompleteAndAwardXP runs both the workout completion and the XP update in one
	// transaction — never call them separately.
	CompleteAndAwardXP(ctx context.Context, workoutID, userID, xpAmount int64, newLevel int) error
}

type UserStorage interface {
	GetByID(ctx context.Context, id int64) (*database.User, error)
}

type AILogger interface {
	Log(ctx context.Context, req *database.AIRequest) error
}

// ExerciseLister is satisfied by *database.ExerciseStore. When wired, the service
// injects available exercise names into the AI prompt so the AI prefers exercises
// that have GIF previews in the library.
type ExerciseLister interface {
	List(ctx context.Context, muscleGroup string) ([]*database.Exercise, error)
}

type Service struct {
	workouts   Storage
	users      UserStorage
	providers  map[string]ai.Provider
	hub        *ws.Hub
	aiRequests AILogger
	exercises  ExerciseLister // nil = skip library injection
}

func NewService(workouts Storage, users UserStorage, providers map[string]ai.Provider, hub *ws.Hub, aiRequests AILogger) *Service {
	return &Service{workouts: workouts, users: users, providers: providers, hub: hub, aiRequests: aiRequests}
}

// WithExerciseLister wires the exercise library into the service so available
// exercise names are injected into the AI prompt at generation time.
func (s *Service) WithExerciseLister(el ExerciseLister) *Service {
	s.exercises = el
	return s
}

type GenerateRequest struct {
	UserPrompt      string
	MuscleGroup     string
	DurationMinutes int
	Injuries        string
	Goals           string
	AIProvider      string
	Environment     string // Environment is where the user is training: "gym" (default), "home", or "outdoor".
	Language        string // "sv" for Swedish output; exercise names always stay in English.
}

type GenerateResult struct {
	Workout  *database.Workout
	Response ai.WorkoutResponse
	Usage    ai.Usage
}

func (s *Service) Generate(ctx context.Context, userID int64, req GenerateRequest) (*GenerateResult, error) {
	provider, ok := s.providers[req.AIProvider]
	if !ok {
		return nil, fmt.Errorf("%w '%s': valid values are %s", ErrUnknownProvider, req.AIProvider, s.validProviders())
	}

	aiReq := ai.WorkoutRequest{
		UserPrompt:      req.UserPrompt,
		MuscleGroup:     req.MuscleGroup,
		DurationMinutes: req.DurationMinutes,
		Injuries:        req.Injuries,
		Goals:           req.Goals,
		Environment:     req.Environment,
		Language:        req.Language,
	}

	if s.exercises != nil {
		if exList, err := s.exercises.List(ctx, ""); err == nil {
			names := make([]string, len(exList))
			for i, ex := range exList {
				names[i] = ex.Name
			}
			aiReq.AvailableExercises = names
		}
		// non-fatal: if the library fetch fails, generate without the hint
	}

	start := time.Now()
	response, usage, err := provider.GenerateWorkout(ctx, aiReq)
	latencyMs := time.Since(start).Milliseconds()

	if s.aiRequests != nil {
		logEntry := &database.AIRequest{
			UserID:       userID,
			Provider:     req.AIProvider,
			InputTokens:  usage.InputTokens,
			OutputTokens: usage.OutputTokens,
			CostUSD:      usage.CostUSD,
			LatencyMs:    latencyMs,
			ValidJSON:    err == nil,
		}
		if err != nil {
			logEntry.ErrorMessage = sql.NullString{String: err.Error(), Valid: true}
		}
		if logErr := s.aiRequests.Log(ctx, logEntry); logErr != nil {
			slog.Default().Warn("failed to log ai request", "error", logErr)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("generate workout: %w", err)
	}

	rawJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}

	w := &database.Workout{
		UserID:          userID,
		Prompt:          req.UserPrompt,
		MuscleGroup:     req.MuscleGroup,
		DurationMinutes: req.DurationMinutes,
		Injuries:        sql.NullString{String: req.Injuries, Valid: req.Injuries != ""},
		Goals:           sql.NullString{String: req.Goals, Valid: req.Goals != ""},
		AIProvider:      req.AIProvider,
		AIResponse:      rawJSON,
		XPEarned:        int(xp.Earned(req.DurationMinutes)),
	}

	saved, err := s.workouts.Create(ctx, w)
	if err != nil {
		return nil, fmt.Errorf("save workout: %w", err)
	}
	return &GenerateResult{Workout: saved, Response: response, Usage: usage}, nil
}

type CompleteResult struct {
	XPEarned  int   `json:"xp_earned"`
	TotalXP   int64 `json:"total_xp"`
	Level     int   `json:"level"`
	LeveledUp bool  `json:"leveled_up"`
}

func (s *Service) Complete(ctx context.Context, workoutID, userID int64) (*CompleteResult, error) {
	workout, err := s.workouts.GetByID(ctx, workoutID, userID)
	if err != nil {
		return nil, ErrNotFound
	}
	if workout.CompletedAt.Valid {
		return nil, ErrAlreadyCompleted
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	earned := int64(workout.XPEarned)
	newTotalXP := user.XP + earned
	oldLevel := user.Level
	newLevel := xp.LevelForXP(newTotalXP)

	if err := s.workouts.CompleteAndAwardXP(ctx, workoutID, userID, earned, newLevel); err != nil {
		return nil, fmt.Errorf("complete workout: %w", err)
	}

	result := &CompleteResult{
		XPEarned:  workout.XPEarned,
		TotalXP:   newTotalXP,
		Level:     newLevel,
		LeveledUp: newLevel > oldLevel,
	}

	s.hub.Broadcast(userID, ws.Event{Type: ws.EventXPUpdate, Data: result})
	return result, nil
}

func (s *Service) validProviders() string {
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
