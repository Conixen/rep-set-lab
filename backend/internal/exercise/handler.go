package exercise

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leonj/rep-set-lab/internal/database"
)

// validExerciseID allows only alphanumeric characters, hyphens, and underscores
// with a max length of 32. ExerciseDB IDs are short numeric strings (e.g. "0027").
var validExerciseID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)

type Handler struct {
	exercises     *database.ExerciseStore
	exerciseDBKey string
	imageCache    sync.Map // map[exerciseID string][]byte — in-memory GIF cache
	httpClient    *http.Client
}

func NewHandler(exercises *database.ExerciseStore, exerciseDBKey string) *Handler {
	return &Handler{
		exercises:     exercises,
		exerciseDBKey: exerciseDBKey,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}
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

// ProxyImage fetches an exercise GIF from ExerciseDB and caches it in memory.
// The RapidAPI key is never exposed to the client.
// Route: GET /api/v1/exercises/image/:exerciseid
func (h *Handler) ProxyImage(c *gin.Context) {
	exerciseID := c.Param("exerciseid")
	if !validExerciseID.MatchString(exerciseID) {
		c.Status(http.StatusBadRequest)
		return
	}

	// Serve from cache on subsequent requests
	if cached, ok := h.imageCache.Load(exerciseID); ok {
		c.Data(http.StatusOK, "image/gif", cached.([]byte))
		return
	}

	if h.exerciseDBKey == "" {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	endpoint := fmt.Sprintf("https://%s/image?exerciseId=%s&resolution=180",
		exerciseDBHost, exerciseID)

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	req.Header.Set("X-RapidAPI-Key", h.exerciseDBKey)
	req.Header.Set("X-RapidAPI-Host", exerciseDBHost)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.Status(resp.StatusCode)
		return
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20)) // 5 MB max
	if err != nil {
		c.Status(http.StatusBadGateway)
		return
	}

	// Cache and return
	h.imageCache.Store(exerciseID, data)

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/gif"
	}
	c.Data(http.StatusOK, contentType, data)
}
