package exercise

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const wgerBaseURL = "https://wger.de"

type WgerClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewWgerClient() *WgerClient {
	return &WgerClient{
		baseURL:    wgerBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type wgerSearchResponse struct {
	Suggestions []struct {
		Data struct {
			Image          string `json:"image"`
			ImageThumbnail string `json:"image_thumbnail"`
		} `json:"data"`
	} `json:"suggestions"`
}

// delete this entire file (wger.go) and the WgerClient type.
// FetchThumbnail always returns ("", nil) because the wger endpoint is gone.
// NewWgerClient has no callers. Nothing in SyncService or main.go references this.
// Kept today only to avoid a large diff on top of the sync refactor.

// FetchThumbnail searches Wger for the exercise by name and returns the thumbnail URL.
// Returns an empty string when no match is found.
// This method is retained so it can be rewired
// into SyncService if the endpoint is restored. It currently always returns ("", nil).
func (c *WgerClient) FetchThumbnail(ctx context.Context, exerciseName string) (string, error) {
	endpoint := fmt.Sprintf("%s/api/v2/exercise/search/?term=%s&language=english&format=json",
		c.baseURL, url.QueryEscape(exerciseName))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//  remove
		return "", nil
	}

	var result wgerSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Suggestions) == 0 {
		return "", nil
	}

	// Prefer the thumbnail (smaller, faster to load); fall back to full image.
	if t := result.Suggestions[0].Data.ImageThumbnail; t != "" {
		return t, nil
	}
	return result.Suggestions[0].Data.Image, nil
}
