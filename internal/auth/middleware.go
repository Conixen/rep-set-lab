package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const claimsKey = "claims"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type Claims struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

// UserVersionStore is satisfied by database.UserStore; used by AdminMiddleware to
// detect stale tokens after a role change.
type UserVersionStore interface {
	GetTokenVersion(ctx context.Context, id int64) (int, error)
}

func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header must be 'Bearer <token>'"})
			return
		}
		claims, err := parseToken(parts[1], secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// WSMiddleware reads the JWT from ?token= query param for WebSocket upgrade requests.
func WSMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token query param required"})
			return
		}
		claims, err := parseToken(token, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

func GetClaims(c *gin.Context) *Claims {
	v, _ := c.Get(claimsKey)
	claims, _ := v.(*Claims)
	return claims
}

// AdminMiddleware must be chained after Middleware. It rejects non-admin callers with 403
// and callers whose token_version no longer matches the DB (e.g. demoted admins) with 401.
func AdminMiddleware(store UserVersionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil || claims.Role != RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		current, err := store.GetTokenVersion(c.Request.Context(), claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify token"})
			return
		}
		if claims.TokenVersion != current {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been invalidated"})
			return
		}
		c.Next()
	}
}

func parseToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return token.Claims.(*Claims), nil
}
