package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/leonj/rep-set-lab/internal/auth"
)

const testSecret = "test-secret-key-for-unit-tests"

func init() {
	gin.SetMode(gin.TestMode)
}

func makeToken(secret string, userID int64, username string, expiry time.Duration) string {
	claims := &auth.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token
}

// claimsRouter wires Middleware and returns user_id in response for assertion.
func claimsRouter(secret string) *gin.Engine {
	r := gin.New()
	r.Use(auth.Middleware(secret))
	r.GET("/ping", func(c *gin.Context) {
		claims := auth.GetClaims(c)
		c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID, "username": claims.Username})
	})
	return r
}

func TestMiddleware_ValidToken(t *testing.T) {
	token := makeToken(testSecret, 42, "leon", time.Hour)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	claimsRouter(testSecret).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if int64(body["user_id"].(float64)) != 42 {
		t.Errorf("user_id = %v, want 42", body["user_id"])
	}
}

func TestMiddleware_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	claimsRouter(testSecret).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestMiddleware_WrongScheme(t *testing.T) {
	token := makeToken(testSecret, 1, "leon", time.Hour)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Token "+token) // "Token" instead of "Bearer"
	claimsRouter(testSecret).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestMiddleware_ExpiredToken(t *testing.T) {
	token := makeToken(testSecret, 1, "leon", -time.Hour) // already expired

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	claimsRouter(testSecret).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestMiddleware_WrongSecret(t *testing.T) {
	token := makeToken("other-secret", 1, "leon", time.Hour)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	claimsRouter(testSecret).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestMiddleware_MalformedToken(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer not.a.jwt")
	claimsRouter(testSecret).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// --- WSMiddleware ---

func wsRouter(secret string) *gin.Engine {
	r := gin.New()
	r.Use(auth.WSMiddleware(secret))
	r.GET("/ws", func(c *gin.Context) {
		claims := auth.GetClaims(c)
		c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID})
	})
	return r
}

func TestWSMiddleware_ValidToken(t *testing.T) {
	token := makeToken(testSecret, 7, "leon", time.Hour)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws?token="+token, nil)
	wsRouter(testSecret).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestWSMiddleware_MissingToken(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	wsRouter(testSecret).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestWSMiddleware_ExpiredToken(t *testing.T) {
	token := makeToken(testSecret, 1, "leon", -time.Minute)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ws?token="+token, nil)
	wsRouter(testSecret).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}
