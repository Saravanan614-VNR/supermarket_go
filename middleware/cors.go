// Contract ID: CTR-008 (CORS Middleware)
// Service: SupermarketService
// Target Concern: middleware

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// NewCORSMiddleware builds a Gin middleware that applies CORS headers for the
// given list of allowed origins. If allowedOrigins is empty, the middleware
// still runs (so preflight OPTIONS requests get a clean 204) but never emits
// an Access-Control-Allow-Origin header, meaning no browser origin is trusted.
func NewCORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			allowed[o] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}