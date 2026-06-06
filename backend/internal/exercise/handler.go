package exercise

import (
	"fmt"
	"io"
	"log/slog"
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

type gifCacheEntry struct {
	data        []byte
	contentType string
}

type Handler struct {
	exercises     *database.ExerciseStore
	exerciseDBKey string
	imageCache    sync.Map // map[exerciseID string]gifCacheEntry — in-memory L1 cache
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

// ProxyImage serves exercise GIFs with a two-level cache: in-memory (L1) and
// PostgreSQL (L2). ExerciseDB is only contacted on the very first request for
// each image, after which the bytes are stored permanently in the DB.
// Route: GET /api/v1/exercises/image/:exerciseid
func (h *Handler) ProxyImage(c *gin.Context) {
	exerciseID := c.Param("exerciseid")
	if !validExerciseID.MatchString(exerciseID) {
		c.Status(http.StatusBadRequest)
		return
	}

	// L1: in-memory cache (warm within the current process lifetime)
	if cached, ok := h.imageCache.Load(exerciseID); ok {
		entry := cached.(gifCacheEntry)
		c.Data(http.StatusOK, entry.contentType, entry.data)
		return
	}

	// L2: database cache (survives redeploys)
	if dbData, dbCT, err := h.exercises.GetGifBytes(c.Request.Context(), exerciseID); err == nil && len(dbData) > 0 {
		h.imageCache.Store(exerciseID, gifCacheEntry{data: dbData, contentType: dbCT})
		c.Data(http.StatusOK, dbCT, dbData)
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

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/gif"
	}

	entry := gifCacheEntry{data: data, contentType: contentType}
	h.imageCache.Store(exerciseID, entry)

	if err := h.exercises.StoreGifBytes(c.Request.Context(), exerciseID, data, contentType); err != nil {
		slog.Default().Error("gif cache: failed to persist to db", "exercisedb_id", exerciseID, "error", err)
	}

	c.Data(http.StatusOK, contentType, data)
}
