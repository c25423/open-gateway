package middleware

import (
	"net/http"
	"strings"

	"github.com/c25423/open-gateway/internal/config"
	"github.com/gin-gonic/gin"
)

func BearerAuthMiddleware() gin.HandlerFunc {
	// Load tokens
	tokens, err := config.GetTokens()
	if err != nil {
		tokens = []string{}
	}
	// Convert tokens to a map for faster lookups
	tokenMap := make(map[string]bool)
	for _, token := range tokens {
		tokenMap[token] = true
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			return
		}

		// Validate format, expecting "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Malformed authorization token"})
			return
		}

		// Validate token using the converted token map
		token := parts[1]
		if tokenMap[token] {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		}
	}
}
