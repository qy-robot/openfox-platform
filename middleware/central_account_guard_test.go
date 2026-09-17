package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCentralAccountGuardBlocksLegacyIdentityAndPATButAllowsSSO(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_TEAM_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_URL", "https://account.example.com")
	router := gin.New()
	router.Use(CentralAccountLegacyGuard())
	router.Any("/*path", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	for _, scenario := range []struct {
		method, path, body string
		status             int
	}{
		{http.MethodPost, "/api/user/login", "", http.StatusConflict},
		{http.MethodPost, "/api/user/register", "", http.StatusConflict},
		{http.MethodPost, "/api/user/auth/refresh", "", http.StatusNoContent},
		{http.MethodGet, "/api/user/token", "", http.StatusConflict},
		{http.MethodGet, "/api/oauth/github", "", http.StatusConflict},
		{http.MethodPost, "/api/user/", "", http.StatusConflict},
		{http.MethodPut, "/api/user/", "", http.StatusNoContent},
		{http.MethodDelete, "/api/user/42", "", http.StatusConflict},
		{http.MethodDelete, "/api/user/42/bindings/github", "", http.StatusConflict},
		{http.MethodDelete, "/api/user/42/oauth/bindings/github", "", http.StatusConflict},
		{http.MethodPut, "/api/user/self", "", http.StatusNoContent},
		{http.MethodDelete, "/api/user/self", "", http.StatusConflict},
		{http.MethodPost, "/api/user/manage", `{"action":"disable","id":42}`, http.StatusNoContent},
		{http.MethodPost, "/api/user/manage", `{"action":"add_quota","id":42,"value":100}`, http.StatusNoContent},
		{http.MethodPost, "/api/user/manage", `{"action":"promote","id":42}`, http.StatusNoContent},
		{http.MethodPost, "/api/user/manage", `{"action":"demote","id":42}`, http.StatusNoContent},
		{http.MethodGet, "/api/teams", "", http.StatusNoContent},
		{http.MethodGet, "/api/teams/self", "", http.StatusNoContent},
		{http.MethodGet, "/api/teams/7", "", http.StatusNoContent},
		{http.MethodPost, "/api/teams/7/fund", "", http.StatusNoContent},
		{http.MethodPost, "/api/account/sso/exchange", "", http.StatusNoContent},
		{http.MethodGet, "/api/user/self", "", http.StatusNoContent},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(scenario.method, scenario.path, strings.NewReader(scenario.body))
		router.ServeHTTP(recorder, request)
		assert.Equal(t, scenario.status, recorder.Code, scenario.path)
	}
}
