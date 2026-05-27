package admin_test

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

	"github.com/leonj/rep-set-lab/internal/admin"
	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/exercise"
)

const testSecret = "test-secret-key-for-admin-tests"

func init() {
	gin.SetMode(gin.TestMode)
}

// --- stub stores ---

type stubUserStore struct {
	users            []*database.User
	deleted          []int64
	roleUpdate       map[int64]string
	versionIncrement []int64
}

func (s *stubUserStore) ListAll(_ context.Context) ([]*database.User, error) {
	return s.users, nil
}

func (s *stubUserStore) GetByID(_ context.Context, id int64) (*database.User, error) {
	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (s *stubUserStore) Delete(_ context.Context, id int64) error {
	for _, u := range s.users {
		if u.ID == id {
			s.deleted = append(s.deleted, id)
			return nil
		}
	}
	return sql.ErrNoRows
}

func (s *stubUserStore) UpdateRole(_ context.Context, id int64, role string) error {
	for _, u := range s.users {
		if u.ID == id {
			if s.roleUpdate == nil {
				s.roleUpdate = make(map[int64]string)
			}
			s.roleUpdate[id] = role
			u.Role = role
			return nil
		}
	}
	return sql.ErrNoRows
}

func (s *stubUserStore) IncrementTokenVersion(_ context.Context, id int64) error {
	s.versionIncrement = append(s.versionIncrement, id)
	return nil
}

// GetTokenVersion satisfies auth.UserVersionStore so the stub can be passed to AdminMiddleware.
func (s *stubUserStore) GetTokenVersion(_ context.Context, _ int64) (int, error) {
	return 1, nil
}

type stubWorkoutStore struct {
	workouts  []*database.Workout
	completed map[int64]bool
}

func (s *stubWorkoutStore) ListAll(_ context.Context) ([]*database.Workout, error) {
	return s.workouts, nil
}

func (s *stubWorkoutStore) GetByIDAdmin(_ context.Context, id int64) (*database.Workout, error) {
	for _, w := range s.workouts {
		if w.ID == id {
			return w, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (s *stubWorkoutStore) AdminSetCompleted(_ context.Context, id int64, completed bool) error {
	for _, w := range s.workouts {
		if w.ID == id {
			if s.completed == nil {
				s.completed = make(map[int64]bool)
			}
			s.completed[id] = completed
			if completed {
				w.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
			} else {
				w.CompletedAt = sql.NullTime{}
			}
			return nil
		}
	}
	return sql.ErrNoRows
}

type stubSyncer struct {
	result exercise.SyncResult
	err    error
}

func (s *stubSyncer) Sync(_ context.Context) (exercise.SyncResult, error) {
	return s.result, s.err
}

type stubAIRequestStore struct{}

func (s *stubAIRequestStore) ListAdmin(_ context.Context, _, _ int) ([]*database.AIRequestRow, error) {
	return nil, nil
}
func (s *stubAIRequestStore) CountAll(_ context.Context) (int, error) { return 0, nil }
func (s *stubAIRequestStore) ProviderStats(_ context.Context) ([]*database.AIProviderStat, error) {
	return nil, nil
}

// adminTestRouter injects admin claims (userID=1, version=1) and wires AdminMiddleware.
func adminTestRouter(users *stubUserStore, workouts *stubWorkoutStore) *gin.Engine {
	return adminTestRouterWithSyncer(users, workouts, &stubSyncer{})
}

func adminTestRouterWithSyncer(users *stubUserStore, workouts *stubWorkoutStore, syncer *stubSyncer) *gin.Engine {
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("claims", &auth.Claims{
			UserID:       1,
			Username:     "admin",
			Role:         auth.RoleAdmin,
			TokenVersion: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		})
		c.Next()
	})
	r.Use(auth.AdminMiddleware(users))

	h := admin.NewHandler(users, workouts, syncer, &stubAIRequestStore{})
	r.GET("/admin/users", h.ListUsers)
	r.GET("/admin/users/:id", h.GetUser)
	r.PUT("/admin/users/:id", h.UpdateUser)
	r.DELETE("/admin/users/:id", h.DeleteUser)
	r.GET("/admin/workouts", h.ListWorkouts)
	r.GET("/admin/workouts/:id", h.GetWorkout)
	r.PUT("/admin/workouts/:id", h.UpdateWorkout)
	r.POST("/admin/exercises/sync", h.SyncExercises)
	return r
}

// --- user tests ---

func TestListUsers(t *testing.T) {
	store := &stubUserStore{users: []*database.User{
		{ID: 1, Username: "alice", Role: auth.RoleUser},
		{ID: 2, Username: "bob", Role: auth.RoleAdmin},
	}}
	r := adminTestRouter(store, &stubWorkoutStore{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/users", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	users := body["users"].([]any)
	if len(users) != 2 {
		t.Errorf("len(users) = %d, want 2", len(users))
	}
}

func TestGetUser_Found(t *testing.T) {
	store := &stubUserStore{users: []*database.User{{ID: 5, Username: "carol", Role: auth.RoleUser}}}
	r := adminTestRouter(store, &stubWorkoutStore{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/users/5", nil))

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	r := adminTestRouter(&stubUserStore{}, &stubWorkoutStore{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/users/99", nil))

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestUpdateUser_PromoteToAdmin(t *testing.T) {
	store := &stubUserStore{users: []*database.User{{ID: 3, Username: "dave", Role: auth.RoleUser}}}
	r := adminTestRouter(store, &stubWorkoutStore{})

	body, _ := json.Marshal(map[string]string{"role": auth.RoleAdmin})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/admin/users/3", bytes.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if store.roleUpdate[3] != auth.RoleAdmin {
		t.Errorf("role not updated: got %q", store.roleUpdate[3])
	}
	if len(store.versionIncrement) != 1 || store.versionIncrement[0] != 3 {
		t.Errorf("expected token version bumped for user 3, got %v", store.versionIncrement)
	}
}

func TestUpdateUser_InvalidRole(t *testing.T) {
	store := &stubUserStore{users: []*database.User{{ID: 3, Username: "dave", Role: auth.RoleUser}}}
	r := adminTestRouter(store, &stubWorkoutStore{})

	body, _ := json.Marshal(map[string]string{"role": "superuser"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/admin/users/3", bytes.NewReader(body)))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	r := adminTestRouter(&stubUserStore{}, &stubWorkoutStore{})

	body, _ := json.Marshal(map[string]string{"role": auth.RoleAdmin})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/admin/users/99", bytes.NewReader(body)))

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestDeleteUser(t *testing.T) {
	store := &stubUserStore{users: []*database.User{{ID: 7, Username: "eve", Role: auth.RoleUser}}}
	r := adminTestRouter(store, &stubWorkoutStore{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/admin/users/7", nil))

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
	if len(store.deleted) != 1 || store.deleted[0] != 7 {
		t.Errorf("expected user 7 deleted, got %v", store.deleted)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	r := adminTestRouter(&stubUserStore{}, &stubWorkoutStore{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/admin/users/99", nil))

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestDeleteUser_SelfForbidden(t *testing.T) {
	// adminTestRouter injects claims with UserID=1; deleting ID 1 should be blocked
	store := &stubUserStore{users: []*database.User{{ID: 1, Username: "admin", Role: auth.RoleAdmin}}}
	r := adminTestRouter(store, &stubWorkoutStore{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/admin/users/1", nil))

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
	if len(store.deleted) != 0 {
		t.Error("expected no deletion to occur")
	}
}

// --- workout tests ---

func TestListWorkouts(t *testing.T) {
	wStore := &stubWorkoutStore{workouts: []*database.Workout{
		{ID: 1, UserID: 1, MuscleGroup: "chest"},
		{ID: 2, UserID: 2, MuscleGroup: "legs"},
	}}
	r := adminTestRouter(&stubUserStore{}, wStore)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/workouts", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	workouts := body["workouts"].([]any)
	if len(workouts) != 2 {
		t.Errorf("len(workouts) = %d, want 2", len(workouts))
	}
}

func TestGetWorkout_Found(t *testing.T) {
	wStore := &stubWorkoutStore{workouts: []*database.Workout{{ID: 10, UserID: 3, MuscleGroup: "back"}}}
	r := adminTestRouter(&stubUserStore{}, wStore)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/workouts/10", nil))

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGetWorkout_NotFound(t *testing.T) {
	r := adminTestRouter(&stubUserStore{}, &stubWorkoutStore{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/workouts/999", nil))

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestUpdateWorkout_MarkComplete(t *testing.T) {
	wStore := &stubWorkoutStore{workouts: []*database.Workout{{ID: 4, UserID: 1, MuscleGroup: "arms"}}}
	r := adminTestRouter(&stubUserStore{}, wStore)

	body, _ := json.Marshal(map[string]bool{"completed": true})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/admin/workouts/4", bytes.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !wStore.completed[4] {
		t.Error("expected workout 4 marked complete")
	}
}

func TestUpdateWorkout_MarkIncomplete(t *testing.T) {
	now := time.Now()
	wStore := &stubWorkoutStore{workouts: []*database.Workout{
		{ID: 5, UserID: 1, MuscleGroup: "shoulders", CompletedAt: sql.NullTime{Time: now, Valid: true}},
	}}
	r := adminTestRouter(&stubUserStore{}, wStore)

	body, _ := json.Marshal(map[string]bool{"completed": false})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/admin/workouts/5", bytes.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if wStore.completed[5] {
		t.Error("expected workout 5 marked incomplete")
	}
}

func TestUpdateWorkout_NotFound(t *testing.T) {
	r := adminTestRouter(&stubUserStore{}, &stubWorkoutStore{})

	body, _ := json.Marshal(map[string]bool{"completed": true})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/admin/workouts/999", bytes.NewReader(body)))

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// --- sync tests ---

func TestSyncExercises_OK(t *testing.T) {
	syncer := &stubSyncer{result: exercise.SyncResult{Total: 10, GIFs: 4}}
	r := adminTestRouterWithSyncer(&stubUserStore{}, &stubWorkoutStore{}, syncer)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/admin/exercises/sync", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if int(body["total"].(float64)) != 10 {
		t.Errorf("total = %v, want 10", body["total"])
	}
}

func TestSyncExercises_Error(t *testing.T) {
	syncer := &stubSyncer{err: errors.New("db failure")}
	r := adminTestRouterWithSyncer(&stubUserStore{}, &stubWorkoutStore{}, syncer)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/admin/exercises/sync", nil))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

// --- middleware tests ---

func TestAdminMiddleware_UserForbidden(t *testing.T) {
	store := &stubUserStore{}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("claims", &auth.Claims{
			UserID: 2, Username: "bob", Role: auth.RoleUser, TokenVersion: 1,
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
		})
		c.Next()
	})
	r.Use(auth.AdminMiddleware(store))
	r.GET("/admin/users", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/users", nil))

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}
