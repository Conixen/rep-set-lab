package user_test

import (
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

	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/user"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- stubs ---

type stubUsers struct {
	u   *database.User
	err error
}

func (s *stubUsers) GetByID(_ context.Context, _ int64) (*database.User, error) {
	return s.u, s.err
}

type stubWorkouts struct {
	workouts []*database.Workout
	err      error
}

func (s *stubWorkouts) ListByUser(_ context.Context, _ int64) ([]*database.Workout, error) {
	return s.workouts, s.err
}

// statsRouter wires the Stats handler with injected auth claims.
func statsRouter(users *stubUsers, workouts *stubWorkouts) *gin.Engine {
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
	r.GET("/users/me/stats", user.NewHandler(users, workouts).Stats)
	return r
}

// --- tests ---

func TestStats_OK(t *testing.T) {
	users := &stubUsers{u: &database.User{ID: 1, Username: "testuser", XP: 800, Level: 2}}
	workouts := &stubWorkouts{workouts: []*database.Workout{
		{ID: 1, CompletedAt: sql.NullTime{Time: time.Now(), Valid: true}},
		{ID: 2}, // not completed
	}}
	r := statsRouter(users, workouts)

	req := httptest.NewRequest(http.MethodGet, "/users/me/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["username"] != "testuser" {
		t.Errorf("username = %v, want testuser", body["username"])
	}
	if body["total_xp"].(float64) != 800 {
		t.Errorf("total_xp = %v, want 800", body["total_xp"])
	}
	if body["level"].(float64) != 2 {
		t.Errorf("level = %v, want 2", body["level"])
	}
	if body["workouts_completed"].(float64) != 1 {
		t.Errorf("workouts_completed = %v, want 1", body["workouts_completed"])
	}
	if body["workouts_total"].(float64) != 2 {
		t.Errorf("workouts_total = %v, want 2", body["workouts_total"])
	}
	if _, ok := body["next_level_xp"]; !ok {
		t.Error("response missing next_level_xp")
	}
	if _, ok := body["current_level_xp"]; !ok {
		t.Error("response missing current_level_xp")
	}
}

func TestStats_UserNotFound_Returns404(t *testing.T) {
	r := statsRouter(&stubUsers{err: sql.ErrNoRows}, &stubWorkouts{})

	req := httptest.NewRequest(http.MethodGet, "/users/me/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestStats_WorkoutsError_Returns500(t *testing.T) {
	users := &stubUsers{u: &database.User{ID: 1, Username: "testuser", XP: 0, Level: 1}}
	r := statsRouter(users, &stubWorkouts{err: errors.New("db unavailable")})

	req := httptest.NewRequest(http.MethodGet, "/users/me/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestStats_MaxLevel_XPFieldsPresent(t *testing.T) {
	// Level 10 is max — next_level_xp should be 0.
	users := &stubUsers{u: &database.User{ID: 1, Username: "testuser", XP: 99999, Level: 10}}
	r := statsRouter(users, &stubWorkouts{})

	req := httptest.NewRequest(http.MethodGet, "/users/me/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["next_level_xp"].(float64) != 0 {
		t.Errorf("next_level_xp = %v, want 0 at max level", body["next_level_xp"])
	}
}
