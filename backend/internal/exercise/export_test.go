package exercise

import (
	"net/http"

	"github.com/leonj/rep-set-lab/internal/database"
)

// NewHandlerWithClient creates a Handler with an injectable HTTP client.
// Used in tests to intercept upstream ExerciseDB calls without a real network.
func NewHandlerWithClient(exercises *database.ExerciseStore, exerciseDBKey string, client *http.Client) *Handler {
	h := NewHandler(exercises, exerciseDBKey)
	h.httpClient = client
	return h
}
