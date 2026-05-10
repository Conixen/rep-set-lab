package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/xp"
)

type Handler struct {
	users    *database.UserStore
	workouts *database.WorkoutStore
}

func NewHandler(users *database.UserStore, workouts *database.WorkoutStore) *Handler {
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
		"username":            user.Username,
		"xp":                  user.XP,
		"level":               user.Level,
		"next_level_xp":       xp.NextThreshold(user.Level),
		"workouts_total":      len(workouts),
		"workouts_completed":  completed,
	})
}
