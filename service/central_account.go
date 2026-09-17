package service

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	centralSessionIdentityPrefix = "central:"
	CentralReauthenticationAge   = 5 * time.Minute
)

var (
	ErrCentralAccountUnavailable       = errors.New("central account service is unavailable")
	ErrCentralAccountInactive          = errors.New("central account token is inactive")
	ErrCentralReauthenticationRequired = errors.New("central account reauthentication is required")
	ErrCentralOAuthCodeInvalid         = errors.New("central OAuth authorization code is invalid")
)

type CentralLabel struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CentralPrincipal struct {
	Subject     string             `json:"subject"`
	Username    string             `json:"username"`
	Email       string             `json:"email"`
	DisplayName string             `json:"displayName"`
	Status      string             `json:"status"`
	Roles       []string           `json:"roles"`
	Permissions []string           `json:"permissions"`
	Labels      []CentralLabel     `json:"labels"`
	SessionID   string             `json:"-"`
	AuthVersion int64              `json:"-"`
	Session     CentralSessionView `json:"-"`
}

type CentralSessionView struct {
	SID          string `json:"sid"`
	Current      bool   `json:"current"`
	LoginMethod  string `json:"login_method"`
	IP           string `json:"ip"`
	UserAgent    string `json:"user_agent"`
	CreatedAt    int64  `json:"created_at"`
	LastActiveAt int64  `json:"last_active_at"`
	ExpiresAt    int64  `json:"expires_at"`
}

type CentralToken struct {
	AccessToken string
	ExpiresAt   int64
	Issuer      string
	Subject     string
}

func CentralAccountEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ROBO_ACCOUNT_MODE")), "central")
}
func CentralAccountURL() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("ROBO_ACCOUNT_URL")), "/")
}
func CentralAccountIssuer() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("ROBO_ACCOUNT_ISSUER")), "/")
}
func CentralAccountClientID() string { return strings.TrimSpace(os.Getenv("ROBO_ACCOUNT_CLIENT_ID")) }

func centralAccountRequest(method, path string, body any, output any) error {
	_, err := centralAccountRequestWithStatus(method, path, body, output)
	return err
}

func centralAccountRequestWithStatus(method, path string, body any, output any) (int, error) {
	base, err := url.Parse(CentralAccountURL())
	if err != nil || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return 0, ErrCentralAccountUnavailable
	}
	if base.Scheme != "https" {
		ip := net.ParseIP(base.Hostname())
		if base.Scheme != "http" || ip == nil || !ip.IsLoopback() {
			return 0, ErrCentralAccountUnavailable
		}
	}
	encoded, err := common.Marshal(body)
	if err != nil {
		return 0, err
	}
	request, err := http.NewRequest(method, base.String()+path, bytes.NewReader(encoded))
	if err != nil {
		return 0, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(os.Getenv("ROBO_ACCOUNT_INTERNAL_TOKEN")))
	client := http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		return 0, ErrCentralAccountUnavailable
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil || response.StatusCode < 200 || response.StatusCode >= 300 {
		return response.StatusCode, ErrCentralAccountUnavailable
	}
	if err := common.Unmarshal(data, output); err != nil {
		return response.StatusCode, ErrCentralAccountUnavailable
	}
	return response.StatusCode, nil
}

func IntrospectCentralAccessToken(raw string) (*CentralPrincipal, error) {
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Active      bool               `json:"active"`
			Issuer      string             `json:"issuer"`
			Audience    string             `json:"audience"`
			SessionID   string             `json:"sessionId"`
			AuthVersion int64              `json:"authVersion"`
			Session     CentralSessionView `json:"session"`
			Principal   CentralPrincipal   `json:"principal"`
		} `json:"data"`
	}
	if err := centralAccountRequest(http.MethodPost, "/v1/internal/introspect", map[string]any{"token": raw, "clientID": CentralAccountClientID()}, &response); err != nil {
		return nil, err
	}
	if !response.Success || !response.Data.Active || response.Data.Issuer != CentralAccountIssuer() || response.Data.Audience != CentralAccountClientID() || strings.TrimSpace(response.Data.Principal.Subject) == "" {
		return nil, ErrCentralAccountInactive
	}
	response.Data.Principal.SessionID = response.Data.SessionID
	response.Data.Principal.AuthVersion = response.Data.AuthVersion
	response.Data.Principal.Session = response.Data.Session
	if response.Data.Principal.Roles == nil {
		response.Data.Principal.Roles = make([]string, 0)
	}
	if response.Data.Principal.Permissions == nil {
		response.Data.Principal.Permissions = make([]string, 0)
	}
	if response.Data.Principal.Labels == nil {
		response.Data.Principal.Labels = make([]CentralLabel, 0)
	}
	if response.Data.Principal.SessionID == "" || response.Data.Principal.AuthVersion <= 0 {
		return nil, ErrCentralAccountInactive
	}
	return &response.Data.Principal, nil
}

func ValidateCentralSessionReference(subject, sessionID string, authVersion int64) error {
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Active bool `json:"active"`
		} `json:"data"`
	}
	if err := centralAccountRequest(http.MethodPost, "/v1/internal/session-status", map[string]any{"subject": subject, "sessionId": sessionID, "authVersion": authVersion}, &response); err != nil {
		return err
	}
	if !response.Success || !response.Data.Active {
		return ErrCentralAccountInactive
	}
	return nil
}

func ValidateCentralAccountSubject(subject string) error {
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Active bool `json:"active"`
		} `json:"data"`
	}
	if err := centralAccountRequest(http.MethodPost, "/v1/internal/account-status", map[string]any{"subject": subject}, &response); err != nil {
		return err
	}
	if !response.Success || !response.Data.Active {
		return ErrCentralAccountInactive
	}
	return nil
}

func CentralAuthIdentity(userID int, userAuthVersion int64, principal *CentralPrincipal) AuthIdentity {
	if principal == nil {
		return AuthIdentity{}
	}
	return AuthIdentity{
		UserID:          userID,
		SessionID:       centralSessionIdentityPrefix + principal.SessionID,
		UserAuthVersion: userAuthVersion,
		SessionVersion:  principal.AuthVersion,
	}
}

func ValidateCentralVerificationSession(identity AuthIdentity, principal *CentralPrincipal, requireRecent bool) error {
	if principal == nil || identity.UserID <= 0 || identity.UserAuthVersion <= 0 || principal.Session.SID != principal.SessionID || principal.Subject == "" || principal.AuthVersion <= 0 {
		return ErrCentralAccountInactive
	}
	if strings.HasPrefix(identity.SessionID, centralSessionIdentityPrefix) {
		if identity.SessionID != centralSessionIdentityPrefix+principal.SessionID || identity.SessionVersion != principal.AuthVersion {
			return ErrCentralAccountInactive
		}
	} else {
		session, _, err := ValidateLoginSession(identity)
		if err != nil || session.AuthorityIssuer != CentralAccountIssuer() || session.AuthoritySubject != principal.Subject || session.AuthoritySessionID != principal.SessionID || session.AuthorityAuthVersion != principal.AuthVersion {
			return ErrCentralAccountInactive
		}
	}
	if requireRecent {
		createdAt := time.Unix(principal.Session.CreatedAt, 0)
		now := time.Now()
		if createdAt.After(now.Add(30*time.Second)) || createdAt.Before(now.Add(-CentralReauthenticationAge)) {
			return ErrCentralReauthenticationRequired
		}
	}
	if err := ValidateCentralSessionReference(principal.Subject, principal.SessionID, principal.AuthVersion); err != nil {
		return err
	}
	state, err := model.GetUserVerificationState(identity.UserID)
	if err != nil {
		return err
	}
	if state.Status != common.UserStatusEnabled || state.AuthVersion != identity.UserAuthVersion {
		return ErrAuthTokenInvalid
	}
	return nil
}

func ExchangeCentralAuthorizationCode(code, verifier, redirectURI string) (*CentralToken, error) {
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			AccessToken string `json:"access_token"`
			ExpiresAt   int64  `json:"expires_at"`
			Issuer      string `json:"issuer"`
			Subject     string `json:"subject"`
		} `json:"data"`
	}
	body := map[string]any{"grantType": "authorization_code", "code": code, "codeVerifier": verifier, "clientID": CentralAccountClientID(), "redirectURI": redirectURI}
	status, err := centralAccountRequestWithStatus(http.MethodPost, "/v1/oauth/token", body, &response)
	if err != nil {
		if status == http.StatusBadRequest {
			return nil, ErrCentralOAuthCodeInvalid
		}
		return nil, err
	}
	if !response.Success || response.Data.AccessToken == "" || response.Data.Issuer != CentralAccountIssuer() {
		return nil, ErrCentralAccountInactive
	}
	return &CentralToken{AccessToken: response.Data.AccessToken, ExpiresAt: response.Data.ExpiresAt, Issuer: response.Data.Issuer, Subject: response.Data.Subject}, nil
}
