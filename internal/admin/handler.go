package admin

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/exercise"
	"github.com/leonj/rep-set-lab/internal/validate"
)

type UserStore interface {
	ListAll(ctx context.Context) ([]*database.User, error)
	GetByID(ctx context.Context, id int64) (*database.User, error)
	Delete(ctx context.Context, id int64) error
	UpdateRole(ctx context.Context, id int64, role string) error
	IncrementTokenVersion(ctx context.Context, id int64) error
}

type WorkoutStore interface {
	ListAll(ctx context.Context) ([]*database.Workout, error)
	GetByIDAdmin(ctx context.Context, id int64) (*database.Workout, error)
	AdminSetCompleted(ctx context.Context, id int64, completed bool) error
}

type ExerciseSyncer interface {
	Sync(ctx context.Context) (exercise.SyncResult, error)
}

type Handler struct {
	users    UserStore
	workouts WorkoutStore
	syncer   ExerciseSyncer
}

func NewHandler(users UserStore, workouts WorkoutStore, syncer ExerciseSyncer) *Handler {
	return &Handler{users: users, workouts: workouts, syncer: syncer}
}

// --- Users ---

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.users.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

type updateUserRequest struct {
	Role string `json:"role"`
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.OneOf("role", req.Role, auth.RoleUser, auth.RoleAdmin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.users.UpdateRole(c.Request.Context(), id, req.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	// Bump token version so any existing JWT for this user is invalidated at the admin gate.
	if err := h.users.IncrementTokenVersion(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	claims := auth.GetClaims(c)
	if claims != nil && id == claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete your own account"})
		return
	}

	if err := h.users.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Workouts ---

func (h *Handler) ListWorkouts(c *gin.Context) {
	workouts, err := h.workouts.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workouts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"workouts": workouts})
}

func (h *Handler) GetWorkout(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
		return
	}
	workout, err := h.workouts.GetByIDAdmin(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workout"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"workout": workout})
}

type updateWorkoutRequest struct {
	Completed bool `json:"completed"`
}

func (h *Handler) UpdateWorkout(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
		return
	}

	var req updateWorkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.workouts.AdminSetCompleted(c.Request.Context(), id, req.Completed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workout"})
		return
	}

	workout, err := h.workouts.GetByIDAdmin(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workout"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"workout": workout})
}

// --- Exercises ---

// SyncExercises triggers a media sync for all exercises, fetching thumbnail and GIF
// URLs from Wger and ExerciseDB. Runs synchronously; capped at 5 minutes.
func (h *Handler) SyncExercises(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()
	result, err := h.syncer.Sync(ctx)
	if err != nil {
		slog.Default().Error("exercise sync failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}
