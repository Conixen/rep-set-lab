package auth_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/leonj/rep-set-lab/internal/auth"
	"github.com/leonj/rep-set-lab/internal/database"
)

type stubAuthStore struct {
	users      []*database.User
	nextID     int64
	failCreate bool
}

func (s *stubAuthStore) Create(_ context.Context, email, username, passwordHash string) (*database.User, error) {
	if s.failCreate {
		return nil, &pq.Error{Code: "23505", Message: "unique constraint violated"}
	}
	s.nextID++
	u := &database.User{
		ID:           s.nextID,
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         auth.RoleUser,
		TokenVersion: 1,
	}
	s.users = append(s.users, u)
	return u, nil
}

func (s *stubAuthStore) GetByEmail(_ context.Context, email string) (*database.User, error) {
	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, sql.ErrNoRows
}

func authRouter(store *stubAuthStore) *gin.Engine {
	r := gin.New()
	h := auth.NewHandler(store, testSecret)
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	return r
}

func newLoginStore() *stubAuthStore {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	return &stubAuthStore{
		users: []*database.User{{
			ID:           10,
			Email:        "carol@example.com",
			Username:     "carol",
			PasswordHash: string(hash),
			Role:         auth.RoleUser,
			TokenVersion: 1,
		}},
	}
}

// --- Register ---

func TestRegister_Success(t *testing.T) {
	body, _ := json.Marshal(map[string]string{
		"email": "alice@example.com", "username": "alice", "password": "secret123",
	})
	w := httptest.NewRecorder()
	authRouter(&stubAuthStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == nil {
		t.Error("expected token in response")
	}
	user := resp["user"].(map[string]any)
	if user["username"] != "alice" {
		t.Errorf("username = %v, want alice", user["username"])
	}
	if user["role"] != auth.RoleUser {
		t.Errorf("role = %v, want %q", user["role"], auth.RoleUser)
	}
}

func TestRegister_MissingEmail(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"username": "bob", "password": "secret123"})
	w := httptest.NewRecorder()
	authRouter(&stubAuthStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"email": "not-an-email", "username": "bob", "password": "secret123"})
	w := httptest.NewRecorder()
	authRouter(&stubAuthStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestRegister_MissingUsername(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"email": "bob@example.com", "password": "secret123"})
	w := httptest.NewRecorder()
	authRouter(&stubAuthStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestRegister_MissingPassword(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"email": "bob@example.com", "username": "bob"})
	w := httptest.NewRecorder()
	authRouter(&stubAuthStore{}).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"email": "dup@example.com", "username": "dup", "password": "secret123"})
	w := httptest.NewRecorder()
	authRouter(&stubAuthStore{failCreate: true}).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body)))
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w.Code)
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"email": "carol@example.com", "password": "password123"})
	w := httptest.NewRecorder()
	authRouter(newLoginStore()).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == nil {
		t.Error("expected token in response")
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"email": "nobody@example.com", "password": "password123"})
	w := httptest.NewRecorder()
	authRouter(newLoginStore()).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"email": "carol@example.com", "password": "wrongpassword"})
	w := httptest.NewRecorder()
	authRouter(newLoginStore()).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestLogin_MissingEmail(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"password": "password123"})
	w := httptest.NewRecorder()
	authRouter(newLoginStore()).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"email": "carol@example.com"})
	w := httptest.NewRecorder()
	authRouter(newLoginStore()).ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
