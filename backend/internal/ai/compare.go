package ai

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/database"
	"github.com/leonj/rep-set-lab/internal/validate"
)

const compareTimeout = 90 * time.Second

type ExerciseLister interface {
	List(ctx context.Context, muscleGroup string) ([]*database.Exercise, error)
}

type CompareMetricsLogger interface {
	Log(ctx context.Context, m *database.CompareMetric) error
}

type ProviderResult struct {
	Provider     string             `json:"provider"`
	Response     *WorkoutResponse   `json:"response,omitempty"`
	Usage        *Usage             `json:"usage,omitempty"`
	Error        string             `json:"error,omitempty"`
	LatencyMs    int64              `json:"latency_ms"`
	Behavioral   *BehavioralMetrics `json:"behavioral,omitempty"`
	LibraryMatch *LibraryMatch      `json:"library_match,omitempty"`
}

type CompareHandler struct {
	providers      map[string]Provider
	grader         Grader
	exercises      ExerciseLister
	aiRequests     *database.AIRequestStore
	compareMetrics CompareMetricsLogger
}

func NewCompareHandler(
	providers map[string]Provider,
	grader Grader,
	exercises ExerciseLister,
	aiRequests *database.AIRequestStore,
	compareMetrics CompareMetricsLogger,
) *CompareHandler {
	return &CompareHandler{
		providers:      providers,
		grader:         grader,
		exercises:      exercises,
		aiRequests:     aiRequests,
		compareMetrics: compareMetrics,
	}
}

type compareRequest struct {
	MuscleGroup     string `json:"muscle_group"`
	DurationMinutes int    `json:"duration_minutes"`
	Injuries        string `json:"injuries"`
	Goals           string `json:"goals"`
	Prompt          string `json:"prompt"`
	Environment     string `json:"environment"`
	Language        string `json:"language"` // optional; "sv" = Swedish output, default English
}

func (h *CompareHandler) Compare(c *gin.Context) {
	claims := auth.GetClaims(c)

	var req compareRequest
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
	if req.Environment == "" {
		req.Environment = "gym"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), compareTimeout)
	defer cancel()

	// Build library lookup once — all exercises, names + aliases, lowercase.
	libraryLookup := h.buildLibraryLookup(ctx)

	aiReq := WorkoutRequest{
		UserPrompt:           req.Prompt,
		MuscleGroup:          req.MuscleGroup,
		DurationMinutes:      req.DurationMinutes,
		Injuries:             req.Injuries,
		Goals:                req.Goals,
		Environment:          req.Environment,
		Language:             req.Language,
		SystemPromptOverride: compareSystemPrompt,
	}

	sessionID := newSessionID()

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results []ProviderResult
	)

	for _, p := range h.providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()

			start := time.Now()
			resp, usage, err := p.GenerateWorkout(ctx, aiReq)
			latencyMs := time.Since(start).Milliseconds()

			res := ProviderResult{
				Provider:  p.Name(),
				LatencyMs: latencyMs,
			}

			var gradeResult *GradeResult

			if err != nil {
				res.Error = err.Error()
			} else {
				res.Response = &resp
				res.Usage = &usage

				bm := computeBehavioralMetrics(resp, req.Environment)
				res.Behavioral = &bm

				lm := computeLibraryMatch(resp, libraryLookup)
				res.LibraryMatch = &lm

				// Grade this workout with Groq
				if h.grader != nil {
					var gradeErr error
					gradeResult, gradeErr = h.grader.GradeWorkout(ctx, aiReq, resp)
					if gradeErr != nil {
						slog.Warn("groq grader failed", "provider", p.Name(), "error", gradeErr)
					}
				}
			}

			// Log to ai_requests for the existing admin usage view.
			if h.aiRequests != nil {
				logEntry := &database.AIRequest{
					UserID:    claims.UserID,
					Provider:  p.Name(),
					LatencyMs: latencyMs,
					ValidJSON: err == nil,
				}
				if usage != (Usage{}) {
					logEntry.InputTokens  = usage.InputTokens
					logEntry.OutputTokens = usage.OutputTokens
					logEntry.CostUSD      = usage.CostUSD
				}
				_ = h.aiRequests.Log(ctx, logEntry)
			}

			// Persist compare metrics to DB.
			if h.compareMetrics != nil && res.Response != nil {
				bm := res.Behavioral
				lm := res.LibraryMatch
				m := &database.CompareMetric{
					SessionID:           sessionID,
					UserID:              claims.UserID,
					Provider:            p.Name(),
					MuscleGroup:         req.MuscleGroup,
					DurationMinutes:     req.DurationMinutes,
					Environment:         req.Environment,
					HasInjuries:         strings.TrimSpace(req.Injuries) != "",
					LibraryMatchRate:    lm.MatchRate,
					LibraryMatchCount:   lm.MatchCount,
					LibraryTotalCount:   lm.TotalCount,
					CharCount:           bm.CharCount,
					EmojiCount:          bm.EmojiCount,
					EquipmentViolations: bm.EquipmentViolations,
					CompletenessScore:   bm.CompletenessScore,
					WarmUpCount:         bm.WarmUpCount,
					MainCount:           bm.MainCount,
					TipsCount:           bm.TipsCount,
					AvgNoteLength:       bm.AvgNoteLength,
					NotesPresentRate:    bm.NotesPresentRate,
					EstimatedMinutes:    bm.EstimatedMinutes,
					InputTokens:         usage.InputTokens,
					OutputTokens:        usage.OutputTokens,
					CostUSD:             usage.CostUSD,
					LatencyMs:           int(latencyMs),
				}
				if gradeResult != nil {
					if g := normalizeGrade(gradeResult.InjuryGrade); g != "" {
						m.GroqInjuryGrade = sql.NullString{String: g, Valid: true}
					}
					if g := normalizeGrade(gradeResult.EquipmentGrade); g != "" {
						m.GroqEquipmentGrade = sql.NullString{String: g, Valid: true}
					}
					if g := normalizeGrade(gradeResult.GoalGrade); g != "" {
						m.GroqGoalGrade = sql.NullString{String: g, Valid: true}
					}
					feedback, _ := json.Marshal(map[string]string{
						"injury":    gradeResult.InjuryFeedback,
						"equipment": gradeResult.EquipmentFeedback,
						"goal":      gradeResult.GoalFeedback,
					})
					m.GroqFeedback = sql.NullString{String: string(feedback), Valid: true}
				}
				_ = h.compareMetrics.Log(ctx, m)
			}

			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(p)
	}
	wg.Wait()

	c.JSON(http.StatusOK, gin.H{"results": results})
}

func (h *CompareHandler) buildLibraryLookup(ctx context.Context) map[string]bool {
	lookup := make(map[string]bool)
	if h.exercises == nil {
		return lookup
	}
	exercises, err := h.exercises.List(ctx, "")
	if err != nil {
		return lookup
	}
	for _, ex := range exercises {
		lookup[strings.ToLower(strings.TrimSpace(ex.Name))] = true
		for _, alias := range ex.Aliases {
			lookup[strings.ToLower(strings.TrimSpace(alias))] = true
		}
	}
	return lookup
}

func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

