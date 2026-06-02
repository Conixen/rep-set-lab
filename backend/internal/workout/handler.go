package workout

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/validate"
)

const aiGenerateTimeout = 60 * time.Second

type Handler struct {
	service  *Service
	workouts Storage
}

func NewHandler(service *Service, workouts Storage) *Handler {
	return &Handler{service: service, workouts: workouts}
}

type generateRequest struct {
	Prompt          string `json:"prompt"`
	MuscleGroup     string `json:"muscle_group"`
	DurationMinutes int    `json:"duration_minutes"`
	Injuries        string `json:"injuries"`
	Goals           string `json:"goals"`
	AIProvider      string `json:"ai_provider"`
	Environment     string `json:"environment"` // optional; omitting defaults to "gym" in the AI layer
	Language        string `json:"language"`    // optional; "sv" = Swedish output, default English
}

func (h *Handler) Generate(c *gin.Context) {
	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.RequiredString("muscle_group", req.MuscleGroup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.PositiveInt("duration_minutes", req.DurationMinutes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.RequiredString("ai_provider", req.AIProvider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), aiGenerateTimeout)
	defer cancel()

	claims := auth.GetClaims(c)
	result, err := h.service.Generate(ctx, claims.UserID, GenerateRequest{
		UserPrompt:      req.Prompt,
		MuscleGroup:     req.MuscleGroup,
		DurationMinutes: req.DurationMinutes,
		Injuries:        req.Injuries,
		Goals:           req.Goals,
		AIProvider:      req.AIProvider,
		Environment:     req.Environment,
		Language:        req.Language,
	})
	if err != nil {
		if errors.Is(err, ErrUnknownProvider) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			slog.Error("workout generation failed", "provider", req.AIProvider, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate workout"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workout":  result.Workout,
		"response": result.Response,
		"usage": gin.H{
			"provider":      req.AIProvider,
			"input_tokens":  result.Usage.InputTokens,
			"output_tokens": result.Usage.OutputTokens,
			"cost_usd":      result.Usage.CostUSD,
		},
	})
}

func (h *Handler) List(c *gin.Context) {
	claims := auth.GetClaims(c)
	workouts, err := h.workouts.ListByUser(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workouts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"workouts": workouts})
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
		return
	}
	claims := auth.GetClaims(c)
	workout, err := h.workouts.GetByID(c.Request.Context(), id, claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"workout": workout})
}

func (h *Handler) Complete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workout id"})
		return
	}
	claims := auth.GetClaims(c)
	result, err := h.service.Complete(c.Request.Context(), id, claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "workout not found"})
		case errors.Is(err, ErrAlreadyCompleted):
			c.JSON(http.StatusConflict, gin.H{"error": "workout already completed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete workout"})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}
