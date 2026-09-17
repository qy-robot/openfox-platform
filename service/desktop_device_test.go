package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDesktopDeviceAuthorizationLifecycle(t *testing.T) {
	user := setupAuthSessionTestDB(t)
	useTestSessionSecret(t)
	require.NoError(t, model.DB.AutoMigrate(&model.DesktopDeviceGrant{}))
	browser, err := CreateLoginSession(user.Id, "password", "127.0.0.1", "browser")
	require.NoError(t, err)
	identity, err := ParseAccessToken(browser.AccessToken)
	require.NoError(t, err)
	verifier := strings.Repeat("a", 43)
	sum := sha256.Sum256([]byte(verifier))
	grant, code, userCode, err := CreateDesktopDeviceGrant("My laptop", base64.RawURLEncoding.EncodeToString(sum[:]))
	require.NoError(t, err)
	assert.NotContains(t, grant.DeviceHash, code)
	_, err = ExchangeDesktopDevice(code, strings.Repeat("b", 43), "127.0.0.1", "desktop")
	assert.ErrorIs(t, err, ErrDesktopDeviceInvalid)
	_, err = ExchangeDesktopDevice(code, verifier, "127.0.0.1", "desktop")
	assert.ErrorIs(t, err, ErrDesktopDevicePending)
	_, err = ExchangeDesktopDevice(code, verifier, "127.0.0.1", "desktop")
	assert.ErrorIs(t, err, ErrDesktopDeviceSlowDown)
	require.NoError(t, ApproveDesktopDevice(userCode, identity, true))
	assert.ErrorIs(t, ApproveDesktopDevice(userCode, identity, true), ErrDesktopDeviceInvalid)
	require.NoError(t, model.DB.Model(grant).Update("last_poll_at", time.Now().Unix()-60).Error)
	issuanceErr := errors.New("temporary session storage failure")
	const failSessionCreate = "test:desktop-session-create-failure"
	require.NoError(t, model.DB.Callback().Create().Before("gorm:create").Register(failSessionCreate, func(tx *gorm.DB) {
		if tx.Statement.Table == "user_sessions" {
			tx.AddError(issuanceErr)
		}
	}))
	_, err = ExchangeDesktopDevice(code, verifier, "127.0.0.1", "desktop")
	assert.ErrorIs(t, err, issuanceErr)
	require.NoError(t, model.DB.Callback().Create().Remove(failSessionCreate))
	bundle, err := ExchangeDesktopDevice(code, verifier, "127.0.0.1", "desktop")
	require.NoError(t, err)
	assert.Equal(t, DesktopLoginMethod, bundle.Session.LoginMethod)
	assert.NotEqual(t, browser.Session.SID, bundle.Session.SID)
	assert.NotEmpty(t, bundle.RefreshToken)
	_, err = ExchangeDesktopDevice(code, verifier, "127.0.0.1", "desktop")
	assert.ErrorIs(t, err, ErrDesktopDeviceInvalid)
	_, err = ValidateSessionReference(user.Id, bundle.Session.SID)
	require.NoError(t, err)
	_, err = model.RevokeUserSession(user.Id, bundle.Session.SID, "test_logout")
	require.NoError(t, err)
	_, err = ValidateSessionReference(user.Id, bundle.Session.SID)
	assert.Error(t, err)
}

func TestDesktopDeviceCannotRedeemUntrustedApproval(t *testing.T) {
	for _, scenario := range []string{"expired", "denied", "browser_revoked", "password_reset", "desktop_approval"} {
		t.Run(scenario, func(t *testing.T) {
			user := setupAuthSessionTestDB(t)
			useTestSessionSecret(t)
			require.NoError(t, model.DB.AutoMigrate(&model.DesktopDeviceGrant{}))
			method := "password"
			if scenario == "desktop_approval" {
				method = DesktopLoginMethod
			}
			browser, err := CreateLoginSession(user.Id, method, "127.0.0.1", "browser")
			require.NoError(t, err)
			identity, err := ParseAccessToken(browser.AccessToken)
			require.NoError(t, err)
			verifier := strings.Repeat("c", 43)
			sum := sha256.Sum256([]byte(verifier))
			grant, code, userCode, err := CreateDesktopDeviceGrant("Laptop", base64.RawURLEncoding.EncodeToString(sum[:]))
			require.NoError(t, err)
			if scenario == "desktop_approval" {
				assert.ErrorIs(t, ApproveDesktopDevice(userCode, identity, true), ErrDesktopDeviceDenied)
				return
			}
			require.NoError(t, ApproveDesktopDevice(userCode, identity, scenario != "denied"))
			switch scenario {
			case "expired":
				require.NoError(t, model.DB.Model(grant).Update("expires_at", time.Now().Unix()-1).Error)
			case "browser_revoked":
				_, err = model.RevokeUserSession(user.Id, browser.Session.SID, "test")
				require.NoError(t, err)
			case "password_reset":
				require.NoError(t, model.DB.Model(user).Update("auth_version", user.AuthVersion+1).Error)
			}
			bundle, err := ExchangeDesktopDevice(code, verifier, "127.0.0.1", "desktop")
			assert.Error(t, err)
			assert.Nil(t, bundle)
		})
	}
}

func TestDesktopDeviceConcurrentExchangeIsSingleUse(t *testing.T) {
	user := setupAuthSessionTestDB(t)
	useTestSessionSecret(t)
	require.NoError(t, model.DB.AutoMigrate(&model.DesktopDeviceGrant{}))
	browser, err := CreateLoginSession(user.Id, "password", "127.0.0.1", "browser")
	require.NoError(t, err)
	identity, err := ParseAccessToken(browser.AccessToken)
	require.NoError(t, err)
	verifier := strings.Repeat("d", 43)
	sum := sha256.Sum256([]byte(verifier))
	_, code, userCode, err := CreateDesktopDeviceGrant("Laptop", base64.RawURLEncoding.EncodeToString(sum[:]))
	require.NoError(t, err)
	require.NoError(t, ApproveDesktopDevice(userCode, identity, true))
	results := make(chan error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			<-start
			_, err := ExchangeDesktopDevice(code, verifier, "127.0.0.1", "desktop")
			results <- err
		})
	}
	close(start)
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else {
			assert.ErrorIs(t, err, ErrDesktopDeviceInvalid)
		}
	}
	assert.Equal(t, 1, success)
}

func TestCentralDesktopSessionTracksAndRevalidatesParentSession(t *testing.T) {
	user := setupAuthSessionTestDB(t)
	useTestSessionSecret(t)
	require.NoError(t, model.DB.AutoMigrate(&model.DesktopDeviceGrant{}))
	var parentActive atomic.Bool
	parentActive.Store(true)
	accountServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"data":{"active":%t}}`, parentActive.Load())))
	}))
	t.Cleanup(accountServer.Close)
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_URL", accountServer.URL)
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	t.Setenv("ROBO_ACCOUNT_INTERNAL_TOKEN", "internal-secret")

	verifier := strings.Repeat("e", 43)
	sum := sha256.Sum256([]byte(verifier))
	grant, code, userCode, err := CreateDesktopDeviceGrant("Central laptop", base64.RawURLEncoding.EncodeToString(sum[:]))
	require.NoError(t, err)
	principal := &CentralPrincipal{Subject: "acct_1", SessionID: "central-session", AuthVersion: 3}
	require.NoError(t, ApproveCentralDesktopDevice(userCode, user.Id, principal, true))
	bundle, err := ExchangeDesktopDevice(code, verifier, "127.0.0.1", "desktop")
	require.NoError(t, err)
	stored, err := model.GetUserSessionBySID(bundle.Session.SID)
	require.NoError(t, err)
	assert.Equal(t, "https://account.example.com", stored.AuthorityIssuer)
	assert.Equal(t, "acct_1", stored.AuthoritySubject)
	assert.Equal(t, "central-session", stored.AuthoritySessionID)
	assert.Equal(t, int64(3), stored.AuthorityAuthVersion)
	assert.Equal(t, "consumed", func() string {
		var current model.DesktopDeviceGrant
		require.NoError(t, model.DB.First(&current, "device_hash = ?", grant.DeviceHash).Error)
		return current.Status
	}())

	refreshed, _, err := RefreshLoginSession(bundle.RefreshToken, bundle.Session.SID, "127.0.0.1", "desktop")
	require.NoError(t, err)
	parentActive.Store(false)
	_, _, err = RefreshLoginSession(refreshed.RefreshToken, refreshed.Session.SID, "127.0.0.1", "desktop")
	assert.ErrorIs(t, err, ErrLoginSessionRevoked)
	_, err = ValidateSessionReference(user.Id, refreshed.Session.SID)
	assert.ErrorIs(t, err, ErrLoginSessionRevoked)
	parentActive.Store(true)
	accountServer.Close()
	_, err = ValidateSessionReference(user.Id, refreshed.Session.SID)
	assert.ErrorIs(t, err, ErrLoginSessionRevoked)
}

func TestCentralModeRejectsLegacyDesktopApproval(t *testing.T) {
	user := setupAuthSessionTestDB(t)
	useTestSessionSecret(t)
	require.NoError(t, model.DB.AutoMigrate(&model.DesktopDeviceGrant{}))
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	browser, err := CreateLoginSession(user.Id, "password", "127.0.0.1", "browser")
	require.NoError(t, err)
	identity, err := ParseAccessToken(browser.AccessToken)
	require.NoError(t, err)
	verifier := strings.Repeat("f", 43)
	sum := sha256.Sum256([]byte(verifier))
	_, code, userCode, err := CreateDesktopDeviceGrant("Legacy laptop", base64.RawURLEncoding.EncodeToString(sum[:]))
	require.NoError(t, err)
	require.NoError(t, ApproveDesktopDevice(userCode, identity, true))
	_, err = ExchangeDesktopDevice(code, verifier, "127.0.0.1", "desktop")
	assert.ErrorIs(t, err, ErrDesktopDeviceDenied)
}

func TestDesktopPlatformOriginRequiresTrustedTransport(t *testing.T) {
	old := system_setting.ServerAddress
	t.Cleanup(func() { system_setting.ServerAddress = old })
	for _, tc := range []struct {
		url     string
		allowed bool
	}{
		{"https://platform.example", true}, {"http://localhost:3000", true}, {"http://127.0.0.1:3000", true},
		{"http://platform.example", false}, {"https://user:password@platform.example", false}, {"https://platform.example/path", false}, {"https://platform.example?redirect=evil", false},
	} {
		t.Run(tc.url, func(t *testing.T) {
			system_setting.ServerAddress = tc.url
			_, err := DesktopPlatformURL()
			assert.Equal(t, tc.allowed, err == nil)
		})
	}
}
