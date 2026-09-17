package controller

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func desktopError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	code := err.Error()
	switch {
	case errors.Is(err, service.ErrDesktopDeviceInvalid), errors.Is(err, service.ErrDesktopDevicePending), errors.Is(err, service.ErrDesktopDeviceSlowDown), errors.Is(err, service.ErrDesktopDeviceExpired), errors.Is(err, service.ErrDesktopDeviceDenied):
	default:
		status, code = http.StatusInternalServerError, "server_error"
	}
	c.JSON(status, gin.H{"success": false, "error": code, "message": code})
}

func CreateDesktopDeviceCode(c *gin.Context) {
	var input struct {
		DeviceName    string `json:"device_name"`
		CodeChallenge string `json:"code_challenge"`
	}
	if c.ShouldBindJSON(&input) != nil {
		desktopError(c, service.ErrDesktopDeviceInvalid)
		return
	}
	origin, err := service.DesktopPlatformURL()
	if err != nil {
		desktopError(c, err)
		return
	}
	grant, deviceCode, userCode, err := service.CreateDesktopDeviceGrant(input.DeviceName, input.CodeChallenge)
	if err != nil {
		desktopError(c, err)
		return
	}
	verificationURI := origin + "/desktop/authorize"
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"device_code": deviceCode, "user_code": userCode, "verification_uri": verificationURI, "verification_uri_complete": verificationURI + "?user_code=" + url.QueryEscape(userCode), "expires_in": service.DesktopDeviceTTL, "interval": grant.PollInterval}})
}

func GetDesktopDeviceAuthorization(c *gin.Context) {
	if service.CentralAccountEnabled() {
		if _, ok := centralBrowserPrincipal(c); !ok {
			desktopError(c, service.ErrDesktopDeviceDenied)
			return
		}
		code := c.Query("user_code")
		grant, err := service.GetDesktopDeviceAuthorization(code)
		if err != nil {
			desktopError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"user_code": code, "device_name": grant.DeviceName, "expires_at": grant.ExpiresAt}})
		return
	}
	identity, ok := requireBrowserSession(c)
	if !ok {
		return
	}
	session, _, err := service.ValidateLoginSession(identity)
	if err != nil || session.LoginMethod == service.DesktopLoginMethod {
		desktopError(c, service.ErrDesktopDeviceDenied)
		return
	}
	code := c.Query("user_code")
	grant, err := service.GetDesktopDeviceAuthorization(code)
	if err != nil {
		desktopError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"user_code": code, "device_name": grant.DeviceName, "expires_at": grant.ExpiresAt}})
}

func ApproveDesktopDevice(c *gin.Context) {
	var input struct {
		UserCode string `json:"user_code"`
		Approve  *bool  `json:"approve"`
	}
	if c.ShouldBindJSON(&input) != nil || input.Approve == nil {
		desktopError(c, service.ErrDesktopDeviceInvalid)
		return
	}
	if service.CentralAccountEnabled() {
		principal, ok := centralBrowserPrincipal(c)
		if !ok {
			desktopError(c, service.ErrDesktopDeviceDenied)
			return
		}
		userID := c.GetInt("id")
		if err := service.ApproveCentralDesktopDevice(input.UserCode, userID, principal, *input.Approve); err != nil {
			recordUserSecurityAudit(c, userID, "user.desktop_authorize", map[string]any{"success": false})
			desktopError(c, err)
			return
		}
		recordUserSecurityAudit(c, userID, "user.desktop_authorize", map[string]any{"success": true, "approved": *input.Approve})
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"approved": *input.Approve}})
		return
	}
	identity, ok := requireBrowserSession(c)
	if !ok {
		return
	}
	if err := service.ApproveDesktopDevice(input.UserCode, identity, *input.Approve); err != nil {
		recordUserSecurityAudit(c, identity.UserID, "user.desktop_authorize", map[string]any{"success": false})
		desktopError(c, err)
		return
	}
	recordUserSecurityAudit(c, identity.UserID, "user.desktop_authorize", map[string]any{"success": true, "approved": *input.Approve})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"approved": *input.Approve}})
}

func centralBrowserPrincipal(c *gin.Context) (*service.CentralPrincipal, bool) {
	identity, ok := middleware.GetSessionAuthIdentity(c)
	if !ok {
		return nil, false
	}
	session, _, err := service.ValidateLoginSession(identity)
	if err != nil || session.LoginMethod == service.DesktopLoginMethod || session.AuthorityIssuer != service.CentralAccountIssuer() || session.AuthoritySubject == "" || session.AuthoritySessionID == "" || session.AuthorityAuthVersion <= 0 {
		return nil, false
	}
	return &service.CentralPrincipal{
		Subject: session.AuthoritySubject, SessionID: session.AuthoritySessionID, AuthVersion: session.AuthorityAuthVersion,
	}, true
}

func desktopAuthResponse(c *gin.Context, bundle *service.AuthBundle) {
	data := authRotationData(bundle)
	data["refresh_token"] = bundle.RefreshToken
	identity, _, err := service.ParseDashboardAccessToken(bundle.AccessToken)
	if err != nil {
		writeAuthSessionError(c, err)
		return
	}
	user, err := model.GetUserById(identity.UserID, false)
	if err != nil {
		writeAuthSessionError(c, err)
		return
	}
	data["user"] = gin.H{"id": user.Id, "username": user.Username, "display_name": user.DisplayName}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func ExchangeDesktopDeviceCode(c *gin.Context) {
	var input struct {
		DeviceCode   string `json:"device_code"`
		CodeVerifier string `json:"code_verifier"`
	}
	if c.ShouldBindJSON(&input) != nil {
		desktopError(c, service.ErrDesktopDeviceInvalid)
		return
	}
	bundle, err := service.ExchangeDesktopDevice(input.DeviceCode, input.CodeVerifier, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		desktopError(c, err)
		return
	}
	desktopAuthResponse(c, bundle)
}

func RefreshDesktopSession(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
		SessionID    string `json:"session_id"`
	}
	if c.ShouldBindJSON(&input) != nil || input.SessionID == "" || len(input.RefreshToken) > 256 {
		writeAuthSessionError(c, service.ErrRefreshTokenInvalid)
		return
	}
	sid, ok := service.RefreshTokenSID(input.RefreshToken)
	if !ok || sid != input.SessionID {
		writeAuthSessionError(c, service.ErrRefreshTokenInvalid)
		return
	}
	session, err := model.GetUserSessionBySID(sid)
	if err != nil || session.LoginMethod != service.DesktopLoginMethod {
		writeAuthSessionError(c, service.ErrRefreshTokenInvalid)
		return
	}
	bundle, _, err := service.RefreshLoginSession(input.RefreshToken, input.SessionID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		writeAuthSessionError(c, err)
		return
	}
	desktopAuthResponse(c, bundle)
}

func requireDesktopSession(c *gin.Context) (service.AuthIdentity, bool) {
	identity, ok := middleware.GetSessionAuthIdentity(c)
	if !ok {
		writeAuthSessionError(c, service.ErrLoginSessionInvalid)
		return identity, false
	}
	session, _, err := service.ValidateLoginSession(identity)
	if err != nil || session.LoginMethod != service.DesktopLoginMethod {
		writeAuthSessionError(c, service.ErrLoginSessionInvalid)
		return identity, false
	}
	return identity, true
}

func CreateDesktopRelayToken(c *gin.Context) {
	identity, ok := requireDesktopSession(c)
	if !ok {
		return
	}
	var input struct {
		FundingMode string `json:"funding_mode"`
		TeamID      int    `json:"team_id"`
		ConfirmTeam bool   `json:"confirm_team"`
	}
	if c.ShouldBindJSON(&input) != nil || (input.FundingMode != "personal_only" && input.FundingMode != "team_only") || (input.FundingMode == "team_only" && !input.ConfirmTeam) {
		desktopError(c, service.ErrDesktopDeviceInvalid)
		return
	}
	if err := model.ValidateTeamFunding(identity.UserID, input.TeamID, input.FundingMode); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "payment source is unavailable"})
		return
	}
	origin, err := service.DesktopPlatformURL()
	if err != nil {
		desktopError(c, err)
		return
	}
	now := time.Now().Unix()
	// Reuse the credential for this session/source while valid. Switching payer
	// never mutates an existing key, so already-started requests retain identity.
	var token model.Token
	err = model.DB.Where("user_id = ? AND desktop_session_id = ? AND funding_mode = ? AND team_id = ? AND status = ? AND expired_time > ?", identity.UserID, identity.SessionID, input.FundingMode, input.TeamID, common.TokenStatusEnabled, now+300).Order("id desc").First(&token).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		desktopError(c, err)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var count int64
		if err := model.DB.Model(&model.Token{}).Where("desktop_session_id = ? AND expired_time > ?", identity.SessionID, now).Count(&count).Error; err != nil {
			desktopError(c, err)
			return
		}
		if count >= 12 {
			c.JSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "too many active desktop payment sources"})
			return
		}
		key, err := common.GenerateKey()
		if err != nil {
			desktopError(c, err)
			return
		}
		token = model.Token{UserId: identity.UserID, Key: key, Name: "robocoding desktop", Status: common.TokenStatusEnabled, CreatedTime: now, AccessedTime: now, ExpiredTime: now + 3600, UnlimitedQuota: true, FundingMode: input.FundingMode, TeamId: input.TeamID, DesktopSessionID: identity.SessionID}
		if err := token.Insert(); err != nil {
			desktopError(c, err)
			return
		}
		recordUserSecurityAudit(c, identity.UserID, "user.desktop_credential", map[string]any{"success": true, "token_id": token.Id, "funding_mode": token.FundingMode, "team_id": token.TeamId})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"key": "sk-" + strings.TrimPrefix(token.Key, "sk-"), "expires_at": token.ExpiredTime, "base_url": origin + "/v1", "funding_mode": token.FundingMode, "team_id": token.TeamId}})
}

func DeleteDesktopSession(c *gin.Context) {
	identity, ok := requireDesktopSession(c)
	if !ok {
		return
	}
	// Session validation on every desktop relay request makes revocation
	// immediate even if token cache cleanup is delayed.
	if _, err := model.RevokeUserSession(identity.UserID, identity.SessionID, "desktop_logout"); err != nil {
		writeAuthSessionError(c, err)
		return
	}
	recordUserSecurityAudit(c, identity.UserID, "user.desktop_logout", map[string]any{"success": true})
	c.JSON(http.StatusOK, gin.H{"success": true})
}
