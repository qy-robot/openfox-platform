package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func DesktopReleaseInternalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := strings.TrimSpace(os.Getenv("ROBO_RELEASES_INTERNAL_TOKEN"))
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"code":    "RELEASES_NOT_CONFIGURED",
				"message": "release management is not configured",
			})
			return
		}
		provided := strings.TrimSpace(c.GetHeader("X-Robo-Releases-Token"))
		if len(provided) != len(expected) || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"code":    "RELEASES_UNAUTHORIZED",
				"message": "invalid release service credential",
			})
			return
		}
		c.Next()
	}
}
