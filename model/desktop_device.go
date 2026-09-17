package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// DesktopDeviceGrant contains only digests of the out-of-band codes. The
// approval is bound to a live browser session and a single PKCE verifier.
type DesktopDeviceGrant struct {
	DeviceHash           string `gorm:"type:char(64);primaryKey"`
	UserCodeHash         string `gorm:"type:char(64);uniqueIndex;not null"`
	CodeChallenge        string `gorm:"type:varchar(64);not null"`
	DeviceName           string `gorm:"type:varchar(80);not null"`
	Status               string `gorm:"type:varchar(16);not null"`
	UserID               int
	AuthVersion          int64
	ApprovalSessionID    string `gorm:"type:varchar(64)"`
	AuthorityIssuer      string `gorm:"type:varchar(255)"`
	AuthoritySubject     string `gorm:"type:varchar(64)"`
	AuthoritySessionID   string `gorm:"type:varchar(64)"`
	AuthorityAuthVersion int64
	ExpiresAt            int64 `gorm:"index;not null"`
	LastPollAt           int64
	PollInterval         int64
}

func DeleteExpiredDesktopAuth(now time.Time) error {
	cutoff := now.Add(-24 * time.Hour).Unix()
	if err := DB.Where("expires_at < ?", cutoff).Delete(&DesktopDeviceGrant{}).Error; err != nil {
		return err
	}
	return DB.Where("desktop_session_id <> ? AND expired_time > 0 AND expired_time < ?", "", cutoff).Delete(&Token{}).Error
}

func DisableDesktopSessionTokens(userID int, sessionID string) error {
	if userID <= 0 || sessionID == "" {
		return errors.New("invalid desktop session token scope")
	}
	var tokens []Token
	if err := DB.Where("user_id = ? AND desktop_session_id = ?", userID, sessionID).Find(&tokens).Error; err != nil {
		return err
	}
	if err := invalidateTokensCache(tokens); err != nil {
		common.SysLog("failed to invalidate desktop session token cache: " + err.Error())
	}
	return DB.Model(&Token{}).
		Where("user_id = ? AND desktop_session_id = ?", userID, sessionID).
		Update("status", common.TokenStatusDisabled).Error
}
