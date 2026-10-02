package controller

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDesktopSkillConsumerUsesVerifiedSessionAndFixedProxy(t *testing.T) {
	previousDB, previousRedis, previousSecret := model.DB, common.RedisEnabled, common.SessionSecret
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.UserSession{}))
	model.DB, common.RedisEnabled, common.SessionSecret = db, false, "desktop-skill-test-secret"
	t.Cleanup(func() {
		model.DB, common.RedisEnabled, common.SessionSecret = previousDB, previousRedis, previousSecret
	})
	user := &model.User{Username: "desktop-skills", Password: "unused", Role: common.RoleAdminUser, Status: common.UserStatusEnabled, Group: "default", AuthVersion: 1}
	require.NoError(t, db.Create(user).Error)
	active := true
	account := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/internal/session-status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if active {
			_, _ = w.Write([]byte(`{"success":true,"data":{"active":true}}`))
		} else {
			_, _ = w.Write([]byte(`{"success":true,"data":{"active":false}}`))
		}
	}))
	t.Cleanup(account.Close)
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_URL", account.URL)
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	t.Setenv("ROBO_ACCOUNT_INTERNAL_TOKEN", "account-status-secret")
	t.Setenv("ROBO_SKILL_CONSUMER_TOKEN", "consumer-only-secret")
	authority := service.CentralSessionAuthority{Issuer: service.CentralAccountIssuer(), Subject: "acct_verified", SessionID: "account-session-verified", AuthVersion: 1}
	bundle, err := service.CreateCentralDesktopLoginSession(user.Id, user.AuthVersion, authority, "127.0.0.1", "test")
	require.NoError(t, err)
	browser, err := service.CreateCentralBrowserLoginSession(user.Id, user.AuthVersion, authority, "127.0.0.1", "test")
	require.NoError(t, err)
	upstreamCalls, redirect := 0, false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		assert.Empty(t, r.Header.Get("Authorization"))
		assert.Empty(t, r.Header.Get("X-Workbench-Token"))
		assert.Equal(t, "consumer-only-secret", r.Header.Get("X-Skill-Consumer-Token"))
		delegation, err := base64.RawURLEncoding.DecodeString(r.Header.Get("X-Skill-Consumer-Identity"))
		require.NoError(t, err)
		var body map[string]any
		require.NoError(t, common.Unmarshal(delegation, &body))
		assert.Equal(t, "acct_verified", body["subject"])
		assert.NotContains(t, body, "roles")
		assert.Contains(t, []string{"/v1/skill-consumer/skills/robot-test/access", "/v1/skill-consumer/skills/robot-test/claim", "/v1/skill-consumer/skills/robot-test/download"}, r.URL.Path)
		if redirect {
			http.Redirect(w, r, "https://example.com", http.StatusFound)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/download") {
			w.Header().Set("Content-Type", "application/zip")
			_, _ = w.Write([]byte("original zip"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"skillId":"robot-test","owned":true,"reason":"owned"}`))
	}))
	t.Cleanup(upstream.Close)
	t.Setenv("ROBO_SKILL_SERVICE_URL", upstream.URL)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/desktop/skills/:id/access", middleware.UserAuth(), GetDesktopSkillAccess)
	router.POST("/api/desktop/skills/:id/claim", middleware.UserAuth(), ClaimDesktopSkill)
	router.POST("/api/desktop/skills/:id/download", middleware.UserAuth(), DownloadDesktopSkill)
	request := func(method, path, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(`{"subject":"acct_attacker","roles":["admin"]}`))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	for _, route := range []struct{ method, action string }{{"GET", "access"}, {"POST", "claim"}, {"POST", "download"}} {
		w := request(route.method, "/api/desktop/skills/robot-test/"+route.action, bundle.AccessToken)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		if route.action == "download" {
			assert.Equal(t, "original zip", w.Body.String())
		}
	}
	assert.Equal(t, 3, upstreamCalls)
	for _, token := range []string{"", "personal-api-key", browser.AccessToken} {
		w := request("GET", "/api/desktop/skills/robot-test/access", token)
		assert.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())
	}
	assert.Equal(t, 3, upstreamCalls)
	w := request("GET", "/api/desktop/skills/robot-test/access?subject=acct_attacker", bundle.AccessToken)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	redirect = true
	w = request("POST", "/api/desktop/skills/robot-test/claim", bundle.AccessToken)
	assert.Equal(t, http.StatusBadGateway, w.Code)
	active = false
	w = request("POST", "/api/desktop/skills/robot-test/download", bundle.AccessToken)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, 4, upstreamCalls)
}
