package exercise_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leonj/rep-set-lab/internal/exercise"
)

func init() { gin.SetMode(gin.TestMode) }

// mockTransport intercepts outbound HTTP calls so tests never hit a real network.
type mockTransport struct {
	callCount  int
	statusCode int
	body       []byte
	err        error
}

func (m *mockTransport) RoundTrip(*http.Request) (*http.Response, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Header:     http.Header{"Content-Type": []string{"image/gif"}},
		Body:       io.NopCloser(bytes.NewReader(m.body)),
	}, nil
}

func newProxyRouterWithClient(exerciseDBKey string, transport http.RoundTripper) *gin.Engine {
	client := &http.Client{Transport: transport}
	h := exercise.NewHandlerWithClient(nil, exerciseDBKey, client)
	r := gin.New()
	r.GET("/api/v1/exercises/image/:exerciseid", h.ProxyImage)
	return r
}

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

// TestProxyImage_CacheHit_NoSecondUpstreamCall verifies that the second request
// for the same exercise ID is served from the in-memory cache without a second
// upstream call.
func TestProxyImage_CacheHit_NoSecondUpstreamCall(t *testing.T) {
	transport := &mockTransport{
		statusCode: http.StatusOK,
		body:       []byte("GIF89a"), // minimal GIF-like payload
	}
	r := newProxyRouterWithClient("testkey", transport)

	for i, want := range []int{http.StatusOK, http.StatusOK} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
			"/api/v1/exercises/image/0027", nil))
		if w.Code != want {
			t.Errorf("request %d: want %d, got %d", i+1, want, w.Code)
		}
	}

	if transport.callCount != 1 {
		t.Errorf("upstream called %d times; want exactly 1 (cache should serve second request)", transport.callCount)
	}
}

// TestProxyImage_Upstream404_Mirrors404 verifies that a non-200 from ExerciseDB
// is mirrored back to the client unchanged.
func TestProxyImage_Upstream404_Mirrors404(t *testing.T) {
	transport := &mockTransport{statusCode: http.StatusNotFound}
	r := newProxyRouterWithClient("testkey", transport)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		"/api/v1/exercises/image/9999", nil))

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404 mirrored from upstream, got %d", w.Code)
	}
}

// TestProxyImage_UpstreamError_Returns502 verifies that a transport-level error
// (connection refused, DNS failure, etc.) returns 502 Bad Gateway.
func TestProxyImage_UpstreamError_Returns502(t *testing.T) {
	transport := &mockTransport{err: errors.New("connection refused")}
	r := newProxyRouterWithClient("testkey", transport)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		"/api/v1/exercises/image/0001", nil))

	if w.Code != http.StatusBadGateway {
		t.Errorf("want 502 on upstream error, got %d", w.Code)
	}
}
