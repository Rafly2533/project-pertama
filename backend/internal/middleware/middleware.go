package middleware

import (
	"net/http"
	"strings"

	"intan-florist-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func CORS(origins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(origins))
	for _, origin := range origins {
		allowed[origin] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (allowed[origin] || allowed["*"]) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			if origin != "" && !allowed[origin] && !allowed["*"] {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abort(c, http.StatusUnauthorized, "authentication required")
			return
		}
		claims, err := utils.ParseToken(secret, parts[1])
		if err != nil {
			abort(c, http.StatusUnauthorized, err.Error())
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func Authorize(resource string) gin.HandlerFunc {
	staffResources := map[string]bool{"products": true, "categories": true, "testimonials": true, "banners": true}
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "super_admin" {
			c.Next()
			return
		}
		if role == "staff" && staffResources[resource] && (c.Request.Method == http.MethodGet || c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut) {
			c.Next()
			return
		}
		abort(c, http.StatusForbidden, "insufficient permissions")
	}
}

func abort(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message, "message": message})
}
