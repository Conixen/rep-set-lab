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

	"github.com/leonj/rep-set-lab/internal/ai"
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
	ApproveUser(ctx context.Context, id int64) error
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

type AIRequestStore interface {
	ListAdmin(ctx context.Context, limit, offset int) ([]*database.AIRequestRow, error)
	CountAll(ctx context.Context) (int, error)
	ProviderStats(ctx context.Context) ([]*database.AIProviderStat, error)
}

type CompareMetricsStore interface {
	ProviderAverages(ctx context.Context) ([]*database.ProviderCompareAvg, error)
	LatestSession(ctx context.Context) ([]*database.CompareMetric, error)
	SessionCount(ctx context.Context) (int, error)
}

// Narrator generates a written analysis of historical compare sessions.
type Narrator interface {
	AnalyzeCompare(ctx context.Context, avgs []*database.ProviderCompareAvg, sessionCount int) (*ai.SessionAnalysis, error)
}

type Handler struct {
	users          UserStore
	workouts       WorkoutStore
	syncer         ExerciseSyncer
	aiRequests     AIRequestStore
	compareMetrics CompareMetricsStore
	narrator       Narrator
}

func NewHandler(users UserStore, workouts WorkoutStore, syncer ExerciseSyncer, aiRequests AIRequestStore, compareMetrics CompareMetricsStore, narrator Narrator) *Handler {
	return &Handler{users: users, workouts: workouts, syncer: syncer, aiRequests: aiRequests, compareMetrics: compareMetrics, narrator: narrator}
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

func (h *Handler) ApproveUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.users.ApproveUser(c.Request.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve user"})
		return
	}

	user, err := h.users.GetByID(c.Request.Context(), id)
	if err != nil {
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

// --- AI Requests ---

func (h *Handler) ListAIRequests(c *gin.Context) {
	const pageSize = 10
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	rows, err := h.aiRequests.ListAdmin(c.Request.Context(), pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list ai requests"})
		return
	}
	total, err := h.aiRequests.CountAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count ai requests"})
		return
	}
	stats, err := h.aiRequests.ProviderStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get provider stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"requests":      rows,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
		"total_pages":   (total + pageSize - 1) / pageSize,
		"provider_stats": stats,
	})
}

// --- Compare Analytics ---

func (h *Handler) CompareStats(c *gin.Context) {
	avgs, err := h.compareMetrics.ProviderAverages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get compare stats"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"provider_averages": avgs})
}

// LatestSession returns all compare_metrics rows from the most recent compare run.
func (h *Handler) LatestSession(c *gin.Context) {
	rows, err := h.compareMetrics.LatestSession(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get latest session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": toSessionRows(rows)})
}

const narrativeTimeout = 30 * time.Second

// NarrativeAnalysis calls Groq to produce a written comparative analysis of all recorded sessions.
func (h *Handler) NarrativeAnalysis(c *gin.Context) {
	if h.narrator == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "narrator not configured (GROQ_KEY not set)"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), narrativeTimeout)
	defer cancel()
	avgs, err := h.compareMetrics.ProviderAverages(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load compare averages"})
		return
	}
	if len(avgs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no compare sessions recorded yet"})
		return
	}
	count, err := h.compareMetrics.SessionCount(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count sessions"})
		return
	}
	analysis, err := h.narrator.AnalyzeCompare(ctx, avgs, count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "analysis failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, analysis)
}

// sessionRow is a JSON-friendly view of database.CompareMetric (sql.NullString → *string).
type sessionRow struct {
	Provider            string   `json:"provider"`
	MuscleGroup         string   `json:"muscle_group"`
	DurationMinutes     int      `json:"duration_minutes"`
	Environment         string   `json:"environment"`
	HasInjuries         bool     `json:"has_injuries"`
	LibraryMatchRate    float64  `json:"library_match_rate"`
	LibraryMatchCount   int      `json:"library_match_count"`
	LibraryTotalCount   int      `json:"library_total_count"`
	CharCount           int      `json:"char_count"`
	EmojiCount          int      `json:"emoji_count"`
	EquipmentViolations int      `json:"equipment_violations"`
	CompletenessScore   int      `json:"completeness_score"`
	WarmUpCount         int      `json:"warm_up_count"`
	MainCount           int      `json:"main_count"`
	CoolDownCount       int      `json:"cool_down_count"`
	TipsCount           int      `json:"tips_count"`
	NotesPresentRate    float64  `json:"notes_present_rate"`
	EstimatedMinutes    float64  `json:"estimated_minutes"`
	GroqInjuryGrade     *string  `json:"groq_injury_grade,omitempty"`
	GroqEquipmentGrade  *string  `json:"groq_equipment_grade,omitempty"`
	GroqGoalGrade       *string  `json:"groq_goal_grade,omitempty"`
	GroqFeedback        *string  `json:"groq_feedback,omitempty"`
}

func toSessionRows(rows []*database.CompareMetric) []sessionRow {
	out := make([]sessionRow, len(rows))
	for i, r := range rows {
		out[i] = sessionRow{
			Provider:            r.Provider,
			MuscleGroup:         r.MuscleGroup,
			DurationMinutes:     r.DurationMinutes,
			Environment:         r.Environment,
			HasInjuries:         r.HasInjuries,
			LibraryMatchRate:    r.LibraryMatchRate,
			LibraryMatchCount:   r.LibraryMatchCount,
			LibraryTotalCount:   r.LibraryTotalCount,
			CharCount:           r.CharCount,
			EmojiCount:          r.EmojiCount,
			EquipmentViolations: r.EquipmentViolations,
			CompletenessScore:   r.CompletenessScore,
			WarmUpCount:         r.WarmUpCount,
			MainCount:           r.MainCount,
			CoolDownCount:       r.CoolDownCount,
			TipsCount:           r.TipsCount,
			NotesPresentRate:    r.NotesPresentRate,
			EstimatedMinutes:    r.EstimatedMinutes,
		}
		if r.GroqInjuryGrade.Valid {
			s := r.GroqInjuryGrade.String
			out[i].GroqInjuryGrade = &s
		}
		if r.GroqEquipmentGrade.Valid {
			s := r.GroqEquipmentGrade.String
			out[i].GroqEquipmentGrade = &s
		}
		if r.GroqGoalGrade.Valid {
			s := r.GroqGoalGrade.String
			out[i].GroqGoalGrade = &s
		}
		if r.GroqFeedback.Valid {
			s := r.GroqFeedback.String
			out[i].GroqFeedback = &s
		}
	}
	return out
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
