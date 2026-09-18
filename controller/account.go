package controller

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func CentralAccountSSOStart(c *gin.Context) {
	if !service.CentralAccountEnabled() {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "account center is not enabled"})
		return
	}
	challenge, state := strings.TrimSpace(c.Query("codeChallenge")), strings.TrimSpace(c.Query("state"))
	redirectURI := strings.TrimSpace(c.Query("redirectUri"))
	responseMode, prompt := strings.TrimSpace(c.Query("responseMode")), strings.TrimSpace(c.Query("prompt"))
	if (responseMode != "" && responseMode != "web_message" && responseMode != "json") || (prompt != "" && prompt != "none") || len(challenge) != 43 || len(state) < 16 || !centralRedirectAllowed(redirectURI) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": "OAUTH_REQUEST_INVALID", "message": "authorization request is invalid"})
		return
	}
	base, err := url.Parse(service.CentralAccountIssuer())
	if err != nil || base.Host == "" || base.User != nil || (base.Scheme != "https" && !loopbackHTTP(base)) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "account service is unavailable"})
		return
	}
	base.Path = "/v1/oauth/authorize"
	query := base.Query()
	query.Set("client_id", service.CentralAccountClientID())
	query.Set("redirect_uri", redirectURI)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	query.Set("state", state)
	if responseMode != "" {
		query.Set("response_mode", responseMode)
	}
	if prompt != "" {
		query.Set("prompt", prompt)
	}
	if c.Query("reauth") == "true" {
		query.Set("reauth_after", strconv.FormatInt(time.Now().UnixMilli(), 10))
	}
	base.RawQuery = query.Encode()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"authorizationUrl": base.String(), "state": state, "redirectUri": redirectURI}})
}

func loopbackHTTP(value *url.URL) bool {
	ip := net.ParseIP(value.Hostname())
	return value.Scheme == "http" && ip != nil && ip.IsLoopback()
}

func centralRedirectAllowed(candidate string) bool {
	for allowed := range strings.SplitSeq(os.Getenv("ROBO_ACCOUNT_REDIRECT_URIS"), ",") {
		if candidate != "" && candidate == strings.TrimSpace(allowed) {
			return true
		}
	}
	return candidate != "" && candidate == strings.TrimSpace(os.Getenv("ROBO_ACCOUNT_REDIRECT_URI"))
}

func CentralAccountSSOExchange(c *gin.Context) {
	if !service.CentralAccountEnabled() {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "account center is not enabled"})
		return
	}
	var request struct {
		Code         string `json:"code"`
		CodeVerifier string `json:"codeVerifier"`
		RedirectURI  string `json:"redirectUri"`
	}
	if err := common.DecodeJson(c.Request.Body, &request); err != nil || !centralRedirectAllowed(request.RedirectURI) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": "OAUTH_REQUEST_INVALID", "message": "authorization request is invalid"})
		return
	}
	token, err := service.ExchangeCentralAuthorizationCode(request.Code, request.CodeVerifier, request.RedirectURI)
	if err != nil {
		if errors.Is(err, service.ErrCentralOAuthCodeInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": "OAUTH_CODE_INVALID", "message": "authorization code is invalid"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "code": "AUTH_SERVICE_UNAVAILABLE", "message": "account service is unavailable"})
		return
	}
	principal, err := service.IntrospectCentralAccessToken(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "code": "AUTH_SERVICE_UNAVAILABLE", "message": "account service is unavailable"})
		return
	}
	if principal.Subject != token.Subject {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "code": "AUTH_SUBJECT_MISMATCH", "message": "account identity did not match the authorization grant"})
		return
	}
	userBase, err := model.ResolveAccountProductUser(token.Issuer, principal.Subject, principal.DisplayName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "could not provision product account"})
		return
	}
	user, err := model.GetSelfUserById(userBase.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "could not load product account"})
		return
	}
	bundle, err := service.CreateCentralBrowserLoginSession(userBase.Id, userBase.AuthVersion, service.CentralSessionAuthority{
		Issuer: token.Issuer, Subject: principal.Subject, SessionID: principal.SessionID, AuthVersion: principal.AuthVersion,
	}, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoginSessionRevoked):
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "code": "AUTH_SESSION_REVOKED", "message": "account session is no longer active"})
		case errors.Is(err, service.ErrLoginSessionInvalid):
			c.JSON(http.StatusForbidden, gin.H{"success": false, "code": "AUTH_PRODUCT_ACCOUNT_DISABLED", "message": "AI account is disabled"})
		default:
			status, code := service.AuthSessionErrorCode(err)
			if status == http.StatusInternalServerError {
				status = http.StatusServiceUnavailable
				code = "AUTH_SERVICE_UNAVAILABLE"
			}
			c.JSON(status, gin.H{"success": false, "code": code, "message": "could not create AI login session"})
		}
		return
	}
	service.WriteRefreshCookie(c, bundle.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": gin.H{
		"access_token": bundle.AccessToken, "token_type": bundle.TokenType, "access_expires_at": bundle.AccessExpiresAt,
		"session": bundle.Session, "user": buildSelfUserData(user),
	}})
}
