package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCentralIntrospectionValidatesServiceAuthIssuerAndAudience(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer internal-secret", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"active":true,"issuer":"https://account.example.com","audience":"platform","sessionId":"central-session","authVersion":1,"session":{"sid":"central-session","current":true,"login_method":"central_sso","created_at":1,"last_active_at":1,"expires_at":9999999999},"principal":{"subject":"acct_1","displayName":"User","roles":["staff"],"permissions":["workbench.access"],"labels":[]}}}`))
	}))
	defer server.Close()
	t.Setenv("ROBO_ACCOUNT_MODE", "central")
	t.Setenv("ROBO_ACCOUNT_URL", server.URL)
	t.Setenv("ROBO_ACCOUNT_ISSUER", "https://account.example.com")
	t.Setenv("ROBO_ACCOUNT_CLIENT_ID", "platform")
	t.Setenv("ROBO_ACCOUNT_INTERNAL_TOKEN", "internal-secret")
	principal, err := IntrospectCentralAccessToken("opaque")
	require.NoError(t, err)
	assert.Equal(t, "acct_1", principal.Subject)
	assert.Equal(t, "central-session", principal.Session.SID)
	assert.Equal(t, "central_sso", principal.Session.LoginMethod)
	assert.NotNil(t, principal.Roles)
	assert.NotNil(t, principal.Permissions)
	assert.NotNil(t, principal.Labels)
	t.Setenv("ROBO_ACCOUNT_CLIENT_ID", "other")
	_, err = IntrospectCentralAccessToken("opaque")
	assert.ErrorIs(t, err, ErrCentralAccountInactive)
}

func TestCentralSecurityProofUsesRecentAccountSession(t *testing.T) {
	useTestSessionSecret(t)
	user := setupAuthSessionTestDB(t)
	require.NoError(t, model.DB.AutoMigrate(&model.TwoFA{}, &model.PasskeyCredential{}))
	require.NoError(t, model.DB.Model(user).Update("role", common.RoleRootUser).Error)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/internal/session-status", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"active":true}}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv("ROBO_ACCOUNT_URL", server.URL)
	t.Setenv("ROBO_ACCOUNT_INTERNAL_TOKEN", "internal-secret")
	principal := &CentralPrincipal{
		Subject: "acct_recent", SessionID: "central-session", AuthVersion: 3,
		Session: CentralSessionView{SID: "central-session", CreatedAt: time.Now().Unix()},
	}
	identity := CentralAuthIdentity(user.Id, user.AuthVersion, principal)
	operation := VerificationOperation{Scope: VerificationScopeChannelKeyRead, Context: []byte(`{"channel_id":7}`)}
	requirements, err := GetCentralVerificationRequirements(identity, principal, operation.Scope)
	require.NoError(t, err)
	require.Equal(t, []VerificationMethodOption{{Method: VerificationMethodSession, Available: true}}, requirements.Methods)
	proof, err := VerifyCentralSecurityInput(identity, principal, VerificationInput{Method: VerificationMethodSession, Scope: operation.Scope, Context: operation.Context})
	require.NoError(t, err)
	_, err = ConsumeCentralOperationProof(proof.ProofToken, identity, principal, operation)
	require.NoError(t, err)
	_, err = ConsumeCentralOperationProof(proof.ProofToken, identity, principal, operation)
	assert.ErrorIs(t, err, ErrProofConsumed)

	principal.Session.CreatedAt = time.Now().Add(-CentralReauthenticationAge - time.Second).Unix()
	_, err = GetCentralVerificationRequirements(identity, principal, operation.Scope)
	assert.ErrorIs(t, err, ErrCentralReauthenticationRequired)
}

func TestCentralAccountRejectsNonHTTPSRemoteEndpoint(t *testing.T) {
	t.Setenv("ROBO_ACCOUNT_URL", "http://example.com")
	t.Setenv("ROBO_ACCOUNT_INTERNAL_TOKEN", "secret")
	_, err := IntrospectCentralAccessToken("opaque")
	assert.ErrorIs(t, err, ErrCentralAccountUnavailable)
}
