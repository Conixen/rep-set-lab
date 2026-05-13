package ai

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leonj/rep-set-lab/internal/validate"
)

// compareTimeout gives all providers enough time to respond.
// It is longer than the single-provider timeout because the critical path
// is the slowest provider, not any individual one.
const compareTimeout = 90 * time.Second

type ProviderResult struct {
	Provider  string           `json:"provider"`
	Response  *WorkoutResponse `json:"response,omitempty"`
	Usage     *Usage           `json:"usage,omitempty"`
	Error     string           `json:"error,omitempty"`
	LatencyMs int64            `json:"latency_ms"`
}

type CompareHandler struct {
	providers map[string]Provider
}

func NewCompareHandler(providers map[string]Provider) *CompareHandler {
	return &CompareHandler{providers: providers}
}

type compareRequest struct {
	MuscleGroup     string `json:"muscle_group"`
	DurationMinutes int    `json:"duration_minutes"`
	Injuries        string `json:"injuries"`
	Goals           string `json:"goals"`
	Prompt          string `json:"prompt"`
}

func (h *CompareHandler) Compare(c *gin.Context) {
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

	aiReq := WorkoutRequest{
		UserPrompt:      req.Prompt,
		MuscleGroup:     req.MuscleGroup,
		DurationMinutes: req.DurationMinutes,
		Injuries:        req.Injuries,
		Goals:           req.Goals,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), compareTimeout)
	defer cancel()

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
			res := ProviderResult{
				Provider:  p.Name(),
				LatencyMs: time.Since(start).Milliseconds(),
			}
			if err != nil {
				res.Error = err.Error()
			} else {
				res.Response = &resp
				res.Usage = &usage
			}
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(p)
	}
	wg.Wait()

	c.JSON(http.StatusOK, gin.H{"results": results})
}
