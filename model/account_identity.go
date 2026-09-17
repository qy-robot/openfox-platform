package model

import (
	"errors"
	"fmt"
	"strings"

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

func ResolveAccountProductUser(issuer, subject, displayName string) (*UserBase, error) {
	issuer = strings.TrimRight(strings.TrimSpace(issuer), "/")
	subject = strings.TrimSpace(subject)
	if issuer == "" || subject == "" {
		return nil, errors.New("central account identity is invalid")
	}
	var mapping AccountProductIdentity
	if err := DB.Where("issuer = ? AND subject = ?", issuer, subject).First(&mapping).Error; err == nil {
		return GetUserCache(mapping.UserID)
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
		username := productUsername(subject)
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
