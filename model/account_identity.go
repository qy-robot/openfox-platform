package model

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountProductIdentity struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Issuer    string `json:"issuer" gorm:"type:varchar(255);not null;uniqueIndex:idx_account_product_subject"`
	Subject   string `json:"subject" gorm:"type:varchar(64);not null;uniqueIndex:idx_account_product_subject"`
	UserID    int    `json:"user_id" gorm:"not null;uniqueIndex"`
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`
}

func GetAccountProductIdentityByUserID(userID int) (*AccountProductIdentity, error) {
	if userID <= 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var identity AccountProductIdentity
	if err := DB.Where("user_id = ?", userID).First(&identity).Error; err != nil {
		return nil, err
	}
	return &identity, nil
}

func ResolveAccountProductUser(issuer, subject, centralUsername, displayName string) (*UserBase, error) {
	issuer = strings.TrimRight(strings.TrimSpace(issuer), "/")
	subject = strings.TrimSpace(subject)
	if issuer == "" || subject == "" {
		return nil, errors.New("central account identity is invalid")
	}
	var mapping AccountProductIdentity
	if err := DB.Where("issuer = ? AND subject = ?", issuer, subject).First(&mapping).Error; err == nil {
		user, err := GetUserCache(mapping.UserID)
		if err != nil {
			return nil, err
		}
		if err := adoptCentralUsername(mapping.UserID, user.Username, productUsername(subject), centralUsername); err != nil {
			return nil, err
		}
		return user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	var created User
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("issuer = ? AND subject = ?", issuer, subject).First(&mapping).Error; err == nil {
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		username := provisionedUsername(tx, subject, centralUsername)
		created = User{Username: username, Password: common.GetRandomString(64), DisplayName: truncateProductIdentity(displayName, 20), Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "default", AuthVersion: 1}
		if created.DisplayName == "" {
			created.DisplayName = username
		}
		if err := created.InsertWithTx(tx, 0); err != nil {
			return err
		}
		mapping = AccountProductIdentity{Issuer: issuer, Subject: subject, UserID: created.Id}
		return tx.Create(&mapping).Error
	})
	if err != nil {
		return nil, err
	}
	if created.Id > 0 {
		created.FinishInsert(0)
	}
	return GetUserCache(mapping.UserID)
}

func productUsername(subject string) string {
	digest := common.GenerateHMACWithKey([]byte("account-product-username-v1"), subject)
	if len(digest) > 15 {
		digest = digest[:15]
	}
	return "acct_" + digest
}

// centralUsernameUsable reports whether the account-center username fits the
// platform user table: usernames are unique and capped at 20 runes
// (User.Username validate tag).
func centralUsernameUsable(username string) bool {
	username = strings.TrimSpace(username)
	return username != "" && utf8.RuneCountInString(username) <= 20
}

// provisionedUsername picks the username for a first sign-in: the real
// account-center username when it fits and is not taken, otherwise the
// generated subject digest so a collision can never block provisioning.
func provisionedUsername(tx *gorm.DB, subject, centralUsername string) string {
	username := strings.TrimSpace(centralUsername)
	if centralUsernameUsable(username) {
		var taken int64
		if err := tx.Model(&User{}).Where("username = ?", username).Count(&taken).Error; err != nil || taken == 0 {
			return username
		}
	}
	return productUsername(subject)
}

// adoptCentralUsername replaces the generated acct_ digest with the real
// account-center username once it becomes usable. A user renamed on the
// platform side no longer matches the digest and is never overwritten.
func adoptCentralUsername(userId int, current, generated, centralUsername string) error {
	username := strings.TrimSpace(centralUsername)
	if !centralUsernameUsable(username) || current != generated || current == username {
		return nil
	}
	var taken int64
	if err := DB.Model(&User{}).Where("username = ? AND id <> ?", username, userId).Count(&taken).Error; err != nil {
		return err
	}
	if taken > 0 {
		return nil
	}
	if err := DB.Model(&User{}).Where("id = ? AND username = ?", userId, current).Update("username", username).Error; err != nil {
		return err
	}
	return updateUserCacheField(userId, "Username", username)
}

func truncateProductIdentity(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return value
}

func BindMigratedAccountProductIdentity(issuer, subject string, userID int) error {
	issuer = strings.TrimRight(strings.TrimSpace(issuer), "/")
	subject = strings.TrimSpace(subject)
	if issuer == "" || subject == "" || userID <= 0 {
		return fmt.Errorf("invalid migrated account mapping")
	}
	mapping := AccountProductIdentity{Issuer: issuer, Subject: subject, UserID: userID}
	return DB.Where("issuer = ? AND subject = ?", issuer, subject).FirstOrCreate(&mapping).Error
}
