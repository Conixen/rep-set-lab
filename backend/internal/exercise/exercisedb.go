package exercise

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const exerciseDBHost = "exercisedb.p.rapidapi.com"

// exerciseIDOverrides maps our seed exercise names (lowercase) to known
// ExerciseDB IDs. Used when ExerciseDB's name search returns no results
// because their naming convention differs from ours.
// IDs verified against https://exercisedb.p.rapidapi.com.
var exerciseIDOverrides = map[string]string{
	"barbell row":            "0027", // barbell bent over row
	"seated cable row":       "0180", // cable low seated row
	"cable crossover":        "1269", // cable standing up straight crossovers
	"dumbbell flyes":         "0308", // dumbbell fly
	"incline dumbbell press": "0314", // dumbbell incline bench press
	"ab wheel rollout":       "0857", // wheel rollout
	"cable crunch":           "0175", // cable kneeling crunch
	"lunges":                 "0336", // dumbbell lunge
	"rear delt fly":          "0154", // cable cross-over reverse fly
	"face pull":              "0203", // cable rear delt row (with rope) — closest ExerciseDB match
}

type ExerciseDBClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewExerciseDBClient(apiKey string) *ExerciseDBClient {
	return &ExerciseDBClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// exerciseDBEntry holds the fields we need from the ExerciseDB v2 API.
// Note: gifUrl was removed in the API v2 redesign. Images are now served via
// GET /image?exerciseId={id}&resolution=180 (returns binary GIF data).
// We store the id and return a backend proxy URL so the API key stays server-side.
type exerciseDBEntry struct {
	ID string `json:"id"`
	// gifUrl string `json:"gifUrl"` — removed in ExerciseDB API v2
}

// FetchGIF searches ExerciseDB for the exercise by name and returns a backend
// proxy URL (/api/v1/exercises/image/{id}) that serves the animated GIF.
// Returns an empty string when no match is found.
func (c *ExerciseDBClient) FetchGIF(ctx context.Context, exerciseName string) (string, error) {
	// Check the override map first — avoids an API round-trip for known mismatches.
	if id, ok := exerciseIDOverrides[strings.ToLower(exerciseName)]; ok {
		return "/api/v1/exercises/image/" + id, nil
	}

	endpoint := fmt.Sprintf("https://%s/exercises/name/%s?limit=1&offset=0",
		exerciseDBHost, url.PathEscape(exerciseName))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-RapidAPI-Key", c.apiKey)
	req.Header.Set("X-RapidAPI-Host", exerciseDBHost)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("exercisedb returned %d", resp.StatusCode)
	}

	var entries []exerciseDBEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return "", err
	}

	if len(entries) == 0 || entries[0].ID == "" {
		return "", nil
	}

	// Return the backend proxy URL; the handler fetches the binary GIF from
	// ExerciseDB with the API key and caches it in memory.
	return "/api/v1/exercises/image/" + entries[0].ID, nil
}
