package middleware

import (
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

// A desktop login grants account summaries and model access, not a second
// administrator session or access to password, payment and API-key management.
func desktopDashboardScopeAllowed(c *gin.Context, session *model.UserSession) bool {
	if session.LoginMethod != service.DesktopLoginMethod {
		return true
	}
	switch c.Request.Method + " " + c.Request.URL.Path {
	case "GET /api/user/self", "GET /api/user/models", "GET /api/teams/self", "POST /api/desktop/relay-token", "DELETE /api/desktop/session":
		return true
	default:
		return false
	}
}

func validateDesktopRelaySession(c *gin.Context, token *model.Token) bool {
	if token.DesktopSessionID == "" {
		return true
	}
	if _, err := service.ValidateSessionReference(token.UserId, token.DesktopSessionID); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "desktop session is no longer active", "type": "authentication_error"}})
		return false
	}
	return true
}
