package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWKS wraps a keyfunc.JWKS and provides a gin middleware for JWT verification.
type JWKS struct {
	kf keyfunc.Keyfunc
}

// NewJWKS fetches the Supabase JWKS from the well-known endpoint and returns
// a JWKS ready to use as middleware. The projectURL is the Supabase project
// base URL, e.g. https://<ref>.supabase.co
func NewJWKS(ctx context.Context, projectURL string) (*JWKS, error) {
	jwksURL := strings.TrimRight(projectURL, "/") + "/auth/v1/.well-known/jwks.json"
	kf, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("auth: fetch JWKS from %s: %w", jwksURL, err)
	}
	return &JWKS{kf: kf}, nil
}

// RequireAuth returns a gin middleware that validates a Supabase-issued JWT.
// On success it sets "user_id" in the gin context to the token's sub claim.
func (j *JWKS) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}

		token, err := jwt.Parse(raw, j.kf.Keyfunc)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		sub, _ := claims["sub"].(string)
		if sub == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing subject claim"})
			return
		}

		c.Set("user_id", sub)

		// Extract GitHub username from user_metadata for human-readable publisher name.
		if meta, ok := claims["user_metadata"].(map[string]any); ok {
			if name, _ := meta["preferred_username"].(string); name != "" {
				c.Set("publisher_name", name)
			} else if name, _ := meta["user_name"].(string); name != "" {
				c.Set("publisher_name", name)
			}
		}

		c.Next()
	}
}

// UserID retrieves the authenticated user's ID from the gin context.
func UserID(c *gin.Context) string {
	v, _ := c.Get("user_id")
	id, _ := v.(string)
	return id
}

// PublisherName retrieves the human-readable publisher name (GitHub username) from the gin context.
func PublisherName(c *gin.Context) string {
	v, _ := c.Get("publisher_name")
	name, _ := v.(string)
	return name
}
