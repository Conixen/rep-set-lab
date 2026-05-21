package workout_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/leonj/rep-set-lab/internal/ai"
	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/mock"
	"github.com/leonj/rep-set-lab/internal/workout"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// handlerRouter wires a workout.Handler into a Gin router with fake auth claims.
func handlerRouter(h *workout.Handler) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("claims", &auth.Claims{
			UserID:   1,
			Username: "testuser",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		})
		c.Next()
	})
	r.POST("/workouts/generate", h.Generate)
	r.POST("/workouts/:id/complete", h.Complete)
	r.GET("/workouts", h.List)
	r.GET("/workouts/:id", h.Get)
	return r
}

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

// --- Generate ---

func TestHandleGenerate_OK(t *testing.T) {
	svc := workout.NewService(
		&mock.WorkoutStorage{},
		&mock.UserStorage{},
		map[string]ai.Provider{"mock": &mock.AIProvider{}},
		newHub(),
		nil,
	)
	r := handlerRouter(workout.NewHandler(svc, &mock.WorkoutStorage{}))

	req := httptest.NewRequest(http.MethodPost, "/workouts/generate", jsonBody(map[string]any{
		"muscle_group":     "chest",
		"duration_minutes": 60,
		"ai_provider":      "mock",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
}

func TestHandleGenerate_MissingMuscleGroup(t *testing.T) {
	svc := workout.NewService(&mock.WorkoutStorage{}, &mock.UserStorage{}, nil, newHub(), nil)
	r := handlerRouter(workout.NewHandler(svc, &mock.WorkoutStorage{}))

	req := httptest.NewRequest(http.MethodPost, "/workouts/generate", jsonBody(map[string]any{
		"duration_minutes": 60,
		"ai_provider":      "mock",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleGenerate_MissingDuration(t *testing.T) {
	svc := workout.NewService(&mock.WorkoutStorage{}, &mock.UserStorage{}, nil, newHub(), nil)
	r := handlerRouter(workout.NewHandler(svc, &mock.WorkoutStorage{}))

	req := httptest.NewRequest(http.MethodPost, "/workouts/generate", jsonBody(map[string]any{
		"muscle_group": "chest",
		"ai_provider":  "mock",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleGenerate_UnknownProvider(t *testing.T) {
	svc := workout.NewService(
		&mock.WorkoutStorage{},
		&mock.UserStorage{},
		map[string]ai.Provider{"claude": &mock.AIProvider{}},
		newHub(),
		nil,
	)
	r := handlerRouter(workout.NewHandler(svc, &mock.WorkoutStorage{}))

	req := httptest.NewRequest(http.MethodPost, "/workouts/generate", jsonBody(map[string]any{
		"muscle_group":     "back",
		"duration_minutes": 45,
		"ai_provider":      "nonexistent",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for unknown provider", w.Code)
	}
}

func TestHandleGenerate_ProviderError_Returns500(t *testing.T) {
	provider := &mock.AIProvider{
		GenerateWorkoutFunc: func(_ context.Context, _ ai.WorkoutRequest) (ai.WorkoutResponse, ai.Usage, error) {
			return ai.WorkoutResponse{}, ai.Usage{}, errors.New("upstream timeout")
		},
	}
	svc := workout.NewService(
		&mock.WorkoutStorage{},
		&mock.UserStorage{},
		map[string]ai.Provider{"mock": provider},
		newHub(),
		nil,
	)
	r := handlerRouter(workout.NewHandler(svc, &mock.WorkoutStorage{}))

	req := httptest.NewRequest(http.MethodPost, "/workouts/generate", jsonBody(map[string]any{
		"muscle_group":     "legs",
		"duration_minutes": 30,
		"ai_provider":      "mock",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for provider error", w.Code)
	}
}

func TestHandleGenerate_InvalidJSON(t *testing.T) {
	svc := workout.NewService(&mock.WorkoutStorage{}, &mock.UserStorage{}, nil, newHub(), nil)
	r := handlerRouter(workout.NewHandler(svc, &mock.WorkoutStorage{}))

	req := httptest.NewRequest(http.MethodPost, "/workouts/generate", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for invalid JSON", w.Code)
	}
}

// --- Complete ---

func TestHandleComplete_OK(t *testing.T) {
	svc := workout.NewService(&mock.WorkoutStorage{}, &mock.UserStorage{}, nil, newHub(), nil)
	r := handlerRouter(workout.NewHandler(svc, &mock.WorkoutStorage{}))

	req := httptest.NewRequest(http.MethodPost, "/workouts/1/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
}

func TestHandleComplete_NotFound_Returns404(t *testing.T) {
	workouts := &mock.WorkoutStorage{
		GetByIDFunc: func(_ context.Context, _, _ int64) (*database.Workout, error) {
			return nil, errors.New("sql: no rows in result set")
		},
	}
	svc := workout.NewService(workouts, &mock.UserStorage{}, nil, newHub(), nil)
	r := handlerRouter(workout.NewHandler(svc, workouts))

	req := httptest.NewRequest(http.MethodPost, "/workouts/999/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandleComplete_AlreadyCompleted_Returns409(t *testing.T) {
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
	r := handlerRouter(workout.NewHandler(svc, workouts))

	req := httptest.NewRequest(http.MethodPost, "/workouts/1/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w.Code)
	}
}

func TestHandleComplete_InvalidID_Returns400(t *testing.T) {
	svc := workout.NewService(&mock.WorkoutStorage{}, &mock.UserStorage{}, nil, newHub(), nil)
	r := handlerRouter(workout.NewHandler(svc, &mock.WorkoutStorage{}))

	req := httptest.NewRequest(http.MethodPost, "/workouts/notanumber/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for non-integer ID", w.Code)
	}
}

func TestHandleComplete_DBError_Returns500(t *testing.T) {
	workouts := &mock.WorkoutStorage{
		CompleteAndAwardXPFunc: func(_ context.Context, _, _, _ int64, _ int) error {
			return errors.New("connection reset")
		},
	}
	svc := workout.NewService(workouts, &mock.UserStorage{}, nil, newHub(), nil)
	r := handlerRouter(workout.NewHandler(svc, workouts))

	req := httptest.NewRequest(http.MethodPost, "/workouts/1/complete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for DB error", w.Code)
	}
}
