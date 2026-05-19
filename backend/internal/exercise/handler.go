package exercise

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonj/rep-set-lab/internal/database"
)

type Handler struct {
	exercises *database.ExerciseStore
}

func NewHandler(exercises *database.ExerciseStore) *Handler {
	return &Handler{exercises: exercises}
}

func (h *Handler) List(c *gin.Context) {
	muscleGroup := c.Query("muscle_group")
	exercises, err := h.exercises.List(c.Request.Context(), muscleGroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list exercises"})
		return
	}
	if exercises == nil {
		exercises = []*database.Exercise{}
	}
	c.JSON(http.StatusOK, exercises)
}
