package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jlhal/parejas/config"
	"github.com/jlhal/parejas/pkg/utils"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		// 1. Try Authorization header first (regular HTTP requests)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
				c.Abort()
				return
			}
			tokenStr = parts[1]
		} else {
			// 2. Fallback to ?token= query param (WebSocket connections)
			tokenStr = c.Query("token")
			if tokenStr == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
				c.Abort()
				return
			}
		}

		claims, err := utils.ValidateToken(tokenStr, cfg)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
