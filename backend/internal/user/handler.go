package user

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/xp"
)

type userReader interface {
	GetByID(ctx context.Context, id int64) (*database.User, error)
}

type workoutLister interface {
	ListByUser(ctx context.Context, userID int64) ([]*database.Workout, error)
}

type Handler struct {
	users    userReader
	workouts workoutLister
}

func NewHandler(users userReader, workouts workoutLister) *Handler {
	return &Handler{users: users, workouts: workouts}
}

func (h *Handler) Stats(c *gin.Context) {
	claims := auth.GetClaims(c)

	user, err := h.users.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	workouts, err := h.workouts.ListByUser(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load workouts"})
		return
	}

	completed := 0
	for _, w := range workouts {
		if w.CompletedAt.Valid {
			completed++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"username":           user.Username,
		"total_xp":           user.XP,
		"level":              user.Level,
		"current_level_xp":   xp.CurrentThreshold(user.Level),
		"next_level_xp":      xp.NextThreshold(user.Level),
		"workouts_total":     len(workouts),
		"workouts_completed": completed,
	})
}
