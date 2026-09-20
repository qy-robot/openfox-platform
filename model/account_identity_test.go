package model

import (
	"os"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestResolveAccountProductUserKeepsPlatformRoleIndependent(t *testing.T) {
	previousDB, previousRedis := DB, common.RedisEnabled
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &AccountProductIdentity{}))
	DB = db
	common.RedisEnabled = false
	t.Cleanup(func() { DB = previousDB; common.RedisEnabled = previousRedis })
	user, err := ResolveAccountProductUser("https://account.example.com", "acct_staff", "staff", "擎云·小高")
	require.NoError(t, err)
	assert.Equal(t, common.RoleCommonUser, user.Role, "central staff/admin roles must never elevate the platform ledger role")
	again, err := ResolveAccountProductUser("https://account.example.com", "acct_staff", "staff", "changed")
	require.NoError(t, err)
	assert.Equal(t, user.Id, again.Id)
	var mappings int64
	require.NoError(t, db.Model(&AccountProductIdentity{}).Count(&mappings).Error)
	assert.EqualValues(t, 1, mappings)
}

func TestResolveAccountProductUserUsesCentralUsername(t *testing.T) {
	previousDB, previousRedis := DB, common.RedisEnabled
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &AccountProductIdentity{}))
	DB = db
	common.RedisEnabled = false
	t.Cleanup(func() { DB = previousDB; common.RedisEnabled = previousRedis })
	require.NoError(t, db.Create(&User{Username: "admin", Role: common.RoleRootUser, Status: common.UserStatusEnabled, Group: "default", AuthVersion: 1}).Error)

	t.Run("provisions with the real account-center username", func(t *testing.T) {
		user, err := ResolveAccountProductUser("https://account.example.com", "acct_email", "7745491@qq.com", "7745491@qq.com")
		require.NoError(t, err)
		var stored User
		require.NoError(t, db.First(&stored, "id = ?", user.Id).Error)
		assert.Equal(t, "7745491@qq.com", stored.Username)
	})

	t.Run("falls back to the digest when the username is taken or too long", func(t *testing.T) {
		user, err := ResolveAccountProductUser("https://account.example.com", "acct_collide", "admin", "Admin Two")
		require.NoError(t, err)
		var stored User
		require.NoError(t, db.First(&stored, "id = ?", user.Id).Error)
		assert.Equal(t, productUsername("acct_collide"), stored.Username)

		user, err = ResolveAccountProductUser("https://account.example.com", "acct_long", "this-username-is-way-over-twenty-runes", "Long")
		require.NoError(t, err)
		stored = User{}
		require.NoError(t, db.First(&stored, "id = ?", user.Id).Error)
		assert.Equal(t, productUsername("acct_long"), stored.Username)
	})

	t.Run("adopts the real username for digest-provisioned users on later sign-in", func(t *testing.T) {
		user, err := ResolveAccountProductUser("https://account.example.com", "acct_late", "", "Late Adopter")
		require.NoError(t, err)
		var stored User
		require.NoError(t, db.First(&stored, "id = ?", user.Id).Error)
		require.Equal(t, productUsername("acct_late"), stored.Username)

		_, err = ResolveAccountProductUser("https://account.example.com", "acct_late", "late.adopter", "Late Adopter")
		require.NoError(t, err)
		stored = User{}
		require.NoError(t, db.First(&stored, "id = ?", user.Id).Error)
		assert.Equal(t, "late.adopter", stored.Username)
	})

	t.Run("never overwrites a platform-side rename", func(t *testing.T) {
		user, err := ResolveAccountProductUser("https://account.example.com", "acct_renamed", "", "Renamed")
		require.NoError(t, err)
		require.NoError(t, db.Model(&User{}).Where("id = ?", user.Id).Update("username", "custom-name").Error)

		_, err = ResolveAccountProductUser("https://account.example.com", "acct_renamed", "central.name", "Renamed")
		require.NoError(t, err)
		var stored User
		require.NoError(t, db.First(&stored, "id = ?", user.Id).Error)
		assert.Equal(t, "custom-name", stored.Username)
	})
}

func TestAccountProductIdentityDatabaseMatrix(t *testing.T) {
	for _, engine := range []struct{ name, env string }{{"mysql", "ACCOUNT_TEST_MYSQL_DSN"}, {"postgres", "ACCOUNT_TEST_POSTGRES_DSN"}} {
		t.Run(engine.name, func(t *testing.T) {
			dsn := strings.TrimSpace(os.Getenv(engine.env))
			if dsn == "" {
				t.Skip(engine.env + " is not configured")
			}
			var db *gorm.DB
			var err error
			if engine.name == "mysql" {
				db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
			} else {
				db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
			}
			require.NoError(t, err)
			require.NoError(t, db.AutoMigrate(&User{}, &AccountProductIdentity{}))
			previousDB, previousRedis := DB, common.RedisEnabled
			DB = db
			common.RedisEnabled = false
			t.Cleanup(func() { DB = previousDB; common.RedisEnabled = previousRedis })
			subject := "acct_platform_matrix_" + engine.name
			user, err := ResolveAccountProductUser("https://account.example.com", subject, "matrix_"+engine.name, "Matrix")
			require.NoError(t, err)
			assert.Equal(t, common.RoleCommonUser, user.Role)
			again, err := ResolveAccountProductUser("https://account.example.com", subject, "matrix_"+engine.name, "Matrix")
			require.NoError(t, err)
			assert.Equal(t, user.Id, again.Id)
			require.NoError(t, db.Where("issuer = ? AND subject = ?", "https://account.example.com", subject).Delete(&AccountProductIdentity{}).Error)
			require.NoError(t, db.Unscoped().Delete(&User{}, user.Id).Error)
		})
	}
}
