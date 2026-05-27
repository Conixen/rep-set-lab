package exercise_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leonj/rep-set-lab/internal/exercise"
)

func init() { gin.SetMode(gin.TestMode) }

// newProxyRouter wires a minimal Gin router with ProxyImage using the same
// path used in production. Passing an empty exerciseDBKey simulates the
// "EXERCISEDB_API_KEY not configured" state.
func newProxyRouter(exerciseDBKey string) *gin.Engine {
	h := exercise.NewHandler(nil, exerciseDBKey)
	r := gin.New()
	r.GET("/api/v1/exercises/image/:exerciseid", h.ProxyImage)
	return r
}

// TestProxyImage_InvalidID_Returns400 verifies that IDs failing the
// alphanumeric-only regex are rejected before any upstream call is made.
func TestProxyImage_InvalidID_Returns400(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{"ampersand injection", "0001%26resolution=9999"},
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, // 33 chars
		{"dot in id", "0001.gif"}, // dots not in allowed charset
	}

	r := newProxyRouter("somekey")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
				"/api/v1/exercises/image/"+tc.id, nil))
			if w.Code != http.StatusBadRequest {
				t.Errorf("id=%q: want 400, got %d", tc.id, w.Code)
			}
		})
	}
}

// TestProxyImage_NoAPIKey_Returns503 verifies that a valid ID returns 503
// when no ExerciseDB API key is configured.
func TestProxyImage_NoAPIKey_Returns503(t *testing.T) {
	r := newProxyRouter("") // empty key = not configured

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		"/api/v1/exercises/image/0027", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("want 503, got %d", w.Code)
	}
}

// TestProxyImage_ValidIDFormat_PassesValidation verifies that a well-formed ID
// moves past the validation guard (it will then 503 because there is no key).
func TestProxyImage_ValidIDFormat_PassesValidation(t *testing.T) {
	validIDs := []string{"0027", "1269", "abc-123", "A1_B2"}

	r := newProxyRouter("") // no key keeps us from hitting real ExerciseDB

	for _, id := range validIDs {
		t.Run(id, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
				"/api/v1/exercises/image/"+id, nil))
			// 503 = passed validation, then stopped at "no key configured"
			if w.Code != http.StatusServiceUnavailable {
				t.Errorf("id=%q: want 503 (valid ID, no key), got %d", id, w.Code)
			}
		})
	}
}
