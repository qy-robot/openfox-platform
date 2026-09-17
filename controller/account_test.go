package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCentralAccountSSOStartUsesPublicIssuerNotInternalServiceURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_URL", "http://127.0.0.1:8780")
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	t.Setenv("ROBO_ACCOUNT_CLIENT_ID", "platform")
	t.Setenv("ROBO_ACCOUNT_REDIRECT_URIS", "https://ai.example.com/account/callback")
	router := gin.New()
	router.GET("/api/account/sso/start", CentralAccountSSOStart)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/account/sso/start?codeChallenge="+strings.Repeat("a", 43)+"&state="+strings.Repeat("s", 16)+"&redirectUri=https%3A%2F%2Fai.example.com%2Faccount%2Fcallback", nil)
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data struct {
			AuthorizationURL string `json:"authorizationUrl"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	assert.True(t, strings.HasPrefix(response.Data.AuthorizationURL, "https://account.example.com/v1/oauth/authorize?"))
	assert.NotContains(t, response.Data.AuthorizationURL, "127.0.0.1")
}

func TestCentralAccountSSOStartRequestsFreshLoginForReauthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_URL", "http://127.0.0.1:8780")
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	t.Setenv("ROBO_ACCOUNT_CLIENT_ID", "platform")
	t.Setenv("ROBO_ACCOUNT_REDIRECT_URIS", "https://ai.example.com/account/callback")
	router := gin.New()
	router.GET("/api/account/sso/start", CentralAccountSSOStart)
	request := httptest.NewRequest(http.MethodGet, "/api/account/sso/start?reauth=true&codeChallenge="+strings.Repeat("a", 43)+"&state="+strings.Repeat("s", 16)+"&redirectUri=https%3A%2F%2Fai.example.com%2Faccount%2Fcallback", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var body struct {
		Data struct {
			AuthorizationURL string `json:"authorizationUrl"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
	authorizationURL, err := url.Parse(body.Data.AuthorizationURL)
	require.NoError(t, err)
	assert.NotEmpty(t, authorizationURL.Query().Get("reauth_after"))
}

func TestCentralAccountSSOStartBuildsWebMessageAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	t.Setenv("ROBO_ACCOUNT_CLIENT_ID", "platform")
	t.Setenv("ROBO_ACCOUNT_REDIRECT_URIS", "https://ai.example.com/account/callback")
	router := gin.New()
	router.GET("/api/account/sso/start", CentralAccountSSOStart)
	request := httptest.NewRequest(http.MethodGet, "/api/account/sso/start?responseMode=web_message&prompt=none&codeChallenge="+strings.Repeat("a", 43)+"&state="+strings.Repeat("s", 16)+"&redirectUri=https%3A%2F%2Fai.example.com%2Faccount%2Fcallback", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var body struct {
		Data struct {
			AuthorizationURL string `json:"authorizationUrl"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
	authorizationURL, err := url.Parse(body.Data.AuthorizationURL)
	require.NoError(t, err)
	assert.Equal(t, "web_message", authorizationURL.Query().Get("response_mode"))
	assert.Equal(t, "none", authorizationURL.Query().Get("prompt"))
}

func TestCentralAccountSSOStartBuildsJSONAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	t.Setenv("ROBO_ACCOUNT_CLIENT_ID", "platform")
	t.Setenv("ROBO_ACCOUNT_REDIRECT_URIS", "https://ai.example.com/account/callback")
	router := gin.New()
	router.GET("/api/account/sso/start", CentralAccountSSOStart)
	request := httptest.NewRequest(http.MethodGet, "/api/account/sso/start?responseMode=json&prompt=none&codeChallenge="+strings.Repeat("a", 43)+"&state="+strings.Repeat("s", 16)+"&redirectUri=https%3A%2F%2Fai.example.com%2Faccount%2Fcallback", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var body struct {
		Data struct {
			AuthorizationURL string `json:"authorizationUrl"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
	authorizationURL, err := url.Parse(body.Data.AuthorizationURL)
	require.NoError(t, err)
	assert.Equal(t, "json", authorizationURL.Query().Get("response_mode"))
	assert.Equal(t, "none", authorizationURL.Query().Get("prompt"))
}

func TestCentralAccountSSOStartRejectsUnsupportedWebMessageOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	t.Setenv("ROBO_ACCOUNT_CLIENT_ID", "platform")
	t.Setenv("ROBO_ACCOUNT_REDIRECT_URIS", "https://ai.example.com/account/callback")
	router := gin.New()
	router.GET("/api/account/sso/start", CentralAccountSSOStart)
	for _, suffix := range []string{"responseMode=fragment", "responseMode=web_message&prompt=login"} {
		request := httptest.NewRequest(http.MethodGet, "/api/account/sso/start?"+suffix+"&codeChallenge="+strings.Repeat("a", 43)+"&state="+strings.Repeat("s", 16)+"&redirectUri=https%3A%2F%2Fai.example.com%2Faccount%2Fcallback", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
}

func TestCentralAccountSSOExchangeDistinguishesInvalidCodeFromOutage(t *testing.T) {
	for _, scenario := range []struct {
		name          string
		accountStatus int
		wantStatus    int
		wantCode      string
	}{
		{name: "invalid code", accountStatus: http.StatusBadRequest, wantStatus: http.StatusBadRequest, wantCode: "OAUTH_CODE_INVALID"},
		{name: "account outage", accountStatus: http.StatusInternalServerError, wantStatus: http.StatusServiceUnavailable, wantCode: "AUTH_SERVICE_UNAVAILABLE"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			accountServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(scenario.accountStatus)
			}))
			t.Cleanup(accountServer.Close)
			t.Setenv("ROBO_ACCOUNT_MODE", "central")
			t.Setenv("ROBO_ACCOUNT_URL", accountServer.URL)
			t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
			t.Setenv("ROBO_ACCOUNT_CLIENT_ID", "platform")
			t.Setenv("ROBO_ACCOUNT_INTERNAL_TOKEN", "internal-secret")
			t.Setenv("ROBO_ACCOUNT_REDIRECT_URIS", "https://ai.example.com/account/callback")
			router := gin.New()
			router.POST("/api/account/sso/exchange", CentralAccountSSOExchange)
			request := httptest.NewRequest(http.MethodPost, "/api/account/sso/exchange", strings.NewReader(`{"code":"invalid","codeVerifier":"`+strings.Repeat("v", 43)+`","redirectUri":"https://ai.example.com/account/callback"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assert.Equal(t, scenario.wantStatus, response.Code, response.Body.String())
			assert.Contains(t, response.Body.String(), scenario.wantCode)
		})
	}
}

func TestCentralAccountSSOExchangeCreatesBoundLocalSession(t *testing.T) {
	previousDB, previousRedis, previousSecret := model.DB, common.RedisEnabled, common.SessionSecret
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.AccountProductIdentity{}, &model.UserSession{}))
	model.DB, common.RedisEnabled, common.SessionSecret = db, false, "account-exchange-test-secret"
	t.Cleanup(func() {
		model.DB, common.RedisEnabled, common.SessionSecret = previousDB, previousRedis, previousSecret
	})
	accountServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/oauth/token":
			_, _ = w.Write([]byte(`{"success":true,"data":{"access_token":"account-opaque-token","expires_at":9999999999,"issuer":"https://account.example.com","subject":"acct-1"}}`))
		case "/v1/internal/introspect":
			_, _ = w.Write([]byte(`{"success":true,"data":{"active":true,"issuer":"https://account.example.com","audience":"platform","sessionId":"account-session","authVersion":7,"session":{"sid":"account-session","created_at":1},"principal":{"subject":"acct-1","displayName":"Account User","permissions":["ai.root"]}}}`))
		case "/v1/internal/session-status":
			_, _ = w.Write([]byte(`{"success":true,"data":{"active":true}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(accountServer.Close)
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_URL", accountServer.URL)
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	t.Setenv("ROBO_ACCOUNT_CLIENT_ID", "platform")
	t.Setenv("ROBO_ACCOUNT_INTERNAL_TOKEN", "internal-secret")
	t.Setenv("ROBO_ACCOUNT_REDIRECT_URIS", "https://ai.example.com/account/callback")
	router := gin.New()
	router.POST("/api/account/sso/exchange", CentralAccountSSOExchange)
	request := httptest.NewRequest(http.MethodPost, "/api/account/sso/exchange", strings.NewReader(`{"code":"one-time-code","codeVerifier":"`+strings.Repeat("v", 43)+`","redirectUri":"https://ai.example.com/account/callback"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body struct {
		Data struct {
			AccessToken string                     `json:"access_token"`
			User        struct{ Role, Status int } `json:"user"`
			Session     service.LoginSessionView   `json:"session"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
	assert.NotEqual(t, "account-opaque-token", body.Data.AccessToken)
	identity, internal, err := service.ParseDashboardAccessToken(body.Data.AccessToken)
	require.True(t, internal)
	require.NoError(t, err)
	assert.Equal(t, common.RoleCommonUser, body.Data.User.Role)
	assert.Equal(t, common.UserStatusEnabled, body.Data.User.Status)
	assert.Equal(t, body.Data.Session.SID, identity.SessionID)
	assert.Contains(t, strings.Join(response.Header().Values("Set-Cookie"), "\n"), service.RefreshCookieName+"=")
	var session model.UserSession
	require.NoError(t, db.Where("sid = ?", body.Data.Session.SID).First(&session).Error)
	assert.Equal(t, service.CentralBrowserLoginMethod, session.LoginMethod)
	assert.Equal(t, "acct-1", session.AuthoritySubject)
	assert.Equal(t, "account-session", session.AuthoritySessionID)
	assert.Equal(t, int64(7), session.AuthorityAuthVersion)
}

func TestCentralVerificationHandlersReturnServiceUnavailableOnAccountOutage(t *testing.T) {
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_URL", "http://example.com")
	principal := &service.CentralPrincipal{
		Subject: "acct_1", SessionID: "session-1", AuthVersion: 1,
		Session: service.CentralSessionView{SID: "session-1", CreatedAt: time.Now().Unix()},
	}
	identity := service.CentralAuthIdentity(1, 1, principal)
	for _, scenario := range []struct {
		name   string
		method string
		path   string
		body   string
		invoke func(*gin.Context)
	}{
		{name: "methods", method: http.MethodGet, path: "/api/verification/methods?scope=channel.key.read", invoke: GetVerificationMethods},
		{name: "verify", method: http.MethodPost, path: "/api/verify", body: `{"method":"session","scope":"channel.key.read","context":{"channel_id":1}}`, invoke: UniversalVerify},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest(scenario.method, scenario.path, strings.NewReader(scenario.body))
			context.Set("auth_identity", identity)
			context.Set("central_principal", principal)
			scenario.invoke(context)
			assert.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
			assert.Contains(t, response.Body.String(), "AUTH_SERVICE_UNAVAILABLE")
		})
	}
}
