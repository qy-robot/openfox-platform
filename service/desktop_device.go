package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"gorm.io/gorm"
)

const DesktopLoginMethod = "desktop"
const DesktopDeviceTTL = 600

var ErrDesktopDeviceInvalid = errors.New("invalid_grant")
var ErrDesktopDevicePending = errors.New("authorization_pending")
var ErrDesktopDeviceSlowDown = errors.New("slow_down")
var ErrDesktopDeviceExpired = errors.New("expired_token")
var ErrDesktopDeviceDenied = errors.New("access_denied")

func DesktopPlatformURL() (string, error) {
	u, err := url.Parse(strings.TrimSpace(system_setting.ServerAddress))
	if err != nil || u.User != nil || u.Host == "" || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", errors.New("configure a valid platform origin")
	}
	local := u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"
	if u.Scheme != "https" && !(local && u.Scheme == "http") {
		return "", errors.New("platform origin requires HTTPS")
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func desktopCodeHash(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func normalizeDesktopUserCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
}

func desktopUserCodeHash(code string) string {
	return common.GenerateHMACWithKey([]byte("desktop-user-code-v1:"+common.SessionSecret), normalizeDesktopUserCode(code))
}

func CreateDesktopDeviceGrant(name, challenge string) (*model.DesktopDeviceGrant, string, string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil || len(decoded) != 32 || len(challenge) != 43 {
		return nil, "", "", ErrDesktopDeviceInvalid
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 80 {
		return nil, "", "", ErrDesktopDeviceInvalid
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, "", "", err
	}
	codeBytes := make([]byte, 5)
	if _, err := rand.Read(codeBytes); err != nil {
		return nil, "", "", err
	}
	deviceCode := base64.RawURLEncoding.EncodeToString(secret)
	userCode := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(codeBytes)
	grant := &model.DesktopDeviceGrant{DeviceHash: desktopCodeHash(deviceCode), UserCodeHash: desktopUserCodeHash(userCode), CodeChallenge: challenge, DeviceName: name, Status: "pending", ExpiresAt: time.Now().Unix() + DesktopDeviceTTL, PollInterval: 5}
	if err := model.DB.Create(grant).Error; err != nil {
		return nil, "", "", err
	}
	return grant, deviceCode, userCode[:4] + "-" + userCode[4:], nil
}

func GetDesktopDeviceAuthorization(code string) (*model.DesktopDeviceGrant, error) {
	code = normalizeDesktopUserCode(code)
	if len(code) != 8 {
		return nil, ErrDesktopDeviceInvalid
	}
	var grant model.DesktopDeviceGrant
	err := model.DB.Where("user_code_hash = ? AND status = ? AND expires_at > ?", desktopUserCodeHash(code), "pending", time.Now().Unix()).First(&grant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDesktopDeviceInvalid
	}
	return &grant, err
}

func ApproveDesktopDevice(code string, identity AuthIdentity, approve bool) error {
	session, _, err := ValidateLoginSession(identity)
	if err != nil {
		return err
	}
	// Device authorizations must originate from the browser, never another
	// delegated device credential.
	if session.LoginMethod == DesktopLoginMethod {
		return ErrDesktopDeviceDenied
	}
	grant, err := GetDesktopDeviceAuthorization(code)
	if err != nil {
		return err
	}
	status := "denied"
	if approve {
		status = "approved"
	}
	r := model.DB.Model(&model.DesktopDeviceGrant{}).Where("device_hash = ? AND status = ? AND expires_at > ?", grant.DeviceHash, "pending", time.Now().Unix()).Updates(map[string]any{"status": status, "user_id": identity.UserID, "auth_version": identity.UserAuthVersion, "approval_session_id": identity.SessionID})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected != 1 {
		return ErrDesktopDeviceInvalid
	}
	return nil
}

func ApproveCentralDesktopDevice(code string, userID int, principal *CentralPrincipal, approve bool) error {
	if principal == nil || userID <= 0 || principal.Subject == "" || principal.SessionID == "" || principal.AuthVersion <= 0 {
		return ErrDesktopDeviceDenied
	}
	if err := ValidateCentralSessionReference(principal.Subject, principal.SessionID, principal.AuthVersion); err != nil {
		return ErrDesktopDeviceDenied
	}
	grant, err := GetDesktopDeviceAuthorization(code)
	if err != nil {
		return err
	}
	status := "denied"
	if approve {
		status = "approved"
	}
	result := model.DB.Model(&model.DesktopDeviceGrant{}).Where("device_hash = ? AND status = ? AND expires_at > ?", grant.DeviceHash, "pending", time.Now().Unix()).Updates(map[string]any{"status": status, "user_id": userID, "authority_issuer": CentralAccountIssuer(), "authority_subject": principal.Subject, "authority_session_id": principal.SessionID, "authority_auth_version": principal.AuthVersion})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrDesktopDeviceInvalid
	}
	return nil
}

func ExchangeDesktopDevice(code, verifier, ip, userAgent string) (*AuthBundle, error) {
	if len(code) != 43 || len(verifier) < 43 || len(verifier) > 128 {
		return nil, ErrDesktopDeviceInvalid
	}
	for _, ch := range verifier {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || strings.ContainsRune("-._~", ch)) {
			return nil, ErrDesktopDeviceInvalid
		}
	}
	var grant model.DesktopDeviceGrant
	err := model.DB.First(&grant, "device_hash = ?", desktopCodeHash(code)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDesktopDeviceInvalid
	}
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	if subtle.ConstantTimeCompare([]byte(challenge), []byte(grant.CodeChallenge)) != 1 {
		return nil, ErrDesktopDeviceInvalid
	}
	now := time.Now().Unix()
	if grant.ExpiresAt <= now {
		return nil, ErrDesktopDeviceExpired
	}
	if grant.Status == "denied" {
		return nil, ErrDesktopDeviceDenied
	}
	if grant.Status == "consumed" {
		return nil, ErrDesktopDeviceInvalid
	}
	if grant.LastPollAt > 0 && now < grant.LastPollAt+grant.PollInterval {
		r := model.DB.Model(&model.DesktopDeviceGrant{}).Where("device_hash = ? AND last_poll_at = ? AND poll_interval = ?", grant.DeviceHash, grant.LastPollAt, grant.PollInterval).Updates(map[string]any{"last_poll_at": now, "poll_interval": grant.PollInterval + 5})
		if r.Error != nil {
			return nil, r.Error
		}
		return nil, ErrDesktopDeviceSlowDown
	}
	if grant.Status == "pending" {
		r := model.DB.Model(&model.DesktopDeviceGrant{}).Where("device_hash = ? AND last_poll_at = ?", grant.DeviceHash, grant.LastPollAt).Update("last_poll_at", now)
		if r.Error != nil {
			return nil, r.Error
		}
		return nil, ErrDesktopDevicePending
	}
	if grant.Status != "approved" {
		return nil, ErrDesktopDeviceInvalid
	}
	central := grant.AuthorityIssuer != "" || grant.AuthoritySubject != "" || grant.AuthoritySessionID != ""
	var identity AuthIdentity
	if central {
		if !CentralAccountEnabled() || grant.AuthorityIssuer != CentralAccountIssuer() || ValidateCentralSessionReference(grant.AuthoritySubject, grant.AuthoritySessionID, grant.AuthorityAuthVersion) != nil {
			return nil, ErrDesktopDeviceDenied
		}
		user, loadErr := model.GetUserCache(grant.UserID)
		if loadErr != nil {
			return nil, loadErr
		}
		identity = AuthIdentity{UserID: user.Id, UserAuthVersion: user.AuthVersion}
	} else {
		if CentralAccountEnabled() {
			return nil, ErrDesktopDeviceDenied
		}
		var validateErr error
		identity, validateErr = ValidateSessionReference(grant.UserID, grant.ApprovalSessionID)
		if validateErr != nil || identity.UserAuthVersion != grant.AuthVersion {
			return nil, ErrDesktopDeviceDenied
		}
	}
	r := model.DB.Model(&model.DesktopDeviceGrant{}).Where("device_hash = ? AND status = ? AND expires_at > ?", grant.DeviceHash, "approved", now).Update("status", "consumed")
	if r.Error != nil {
		return nil, r.Error
	}
	if r.RowsAffected != 1 {
		return nil, ErrDesktopDeviceInvalid
	}
	var bundle *AuthBundle
	if central {
		bundle, err = CreateCentralDesktopLoginSession(grant.UserID, identity.UserAuthVersion, CentralSessionAuthority{Issuer: grant.AuthorityIssuer, Subject: grant.AuthoritySubject, SessionID: grant.AuthoritySessionID, AuthVersion: grant.AuthorityAuthVersion}, ip, userAgent)
	} else {
		bundle, err = CreateLoginSessionAtAuthVersion(grant.UserID, grant.AuthVersion, DesktopLoginMethod, ip, userAgent)
	}
	if err != nil {
		// No credentials were returned. Release this exchange's single-use claim
		// so a temporary issuance failure does not destroy the browser approval.
		// Each retry still validates the browser session, auth version and expiry.
		restore := model.DB.Model(&model.DesktopDeviceGrant{}).
			Where("device_hash = ? AND status = ? AND expires_at > ?", grant.DeviceHash, "consumed", time.Now().Unix()).
			Update("status", "approved")
		if restore.Error != nil {
			common.SysError("failed to restore desktop approval after session issuance failure: " + restore.Error.Error())
		}
		return nil, err
	}
	return bundle, nil
}
