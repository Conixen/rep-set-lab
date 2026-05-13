package exercise

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const exerciseDBHost = "exercisedb.p.rapidapi.com"

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

type exerciseDBEntry struct {
	GifURL string `json:"gifUrl"`
}

// FetchGIF searches ExerciseDB for the exercise by name and returns the animated GIF URL.
// Returns an empty string when no match is found.
func (c *ExerciseDBClient) FetchGIF(ctx context.Context, exerciseName string) (string, error) {
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

	if len(entries) == 0 {
		return "", nil
	}
	return entries[0].GifURL, nil
}
