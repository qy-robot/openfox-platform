package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func CentralAccountLegacyGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !service.CentralAccountEnabled() {
			c.Next()
			return
		}
		blocked := legacyIdentityRoute(c.Request.Method, c.Request.URL.Path)
		if !blocked {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"success":    false,
			"code":       "CENTRAL_ACCOUNT_REQUIRED",
			"message":    "account authentication is managed by the account center",
			"accountUrl": service.CentralAccountIssuer(),
		})
	}
}

func legacyIdentityRoute(method, path string) bool {
	if strings.HasPrefix(path, "/api/account/") {
		return false
	}
	if strings.HasPrefix(path, "/api/oauth/") || strings.HasPrefix(path, "/api/user/login") || strings.HasPrefix(path, "/api/user/passkey") || strings.HasPrefix(path, "/api/user/2fa") || strings.HasPrefix(path, "/api/user/oauth/bindings") {
		return true
	}
	switch path {
	case "/api/verification", "/api/reset_password", "/api/user/reset", "/api/user/register":
		return true
	case "/api/user", "/api/user/":
		return method == http.MethodPost || method == http.MethodDelete
	case "/api/user/self":
		return method == http.MethodDelete
	case "/api/user/token", "/api/user/token/status":
		return method != http.MethodOptions
	default:
		if method != http.MethodDelete || !strings.HasPrefix(path, "/api/user/") {
			return false
		}
		parts := strings.Split(strings.TrimPrefix(path, "/api/user/"), "/")
		if len(parts) == 1 {
			_, err := strconv.Atoi(parts[0])
			return err == nil
		}
		if len(parts) == 2 {
			return parts[1] == "reset_passkey" || parts[1] == "2fa"
		}
		return len(parts) >= 3 && (parts[1] == "bindings" || parts[1] == "oauth")
	}
}
