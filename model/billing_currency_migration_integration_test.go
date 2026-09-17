package model

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBillingCurrencySQLiteFreshStartupTwice(t *testing.T) {
	previousPath := common.SQLitePath
	previousMaster := common.IsMasterNode
	previousDB, previousLogDB := DB, LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	previousMainDSN, mainWasSet := os.LookupEnv("SQL_DSN")
	previousLogDSN, logWasSet := os.LookupEnv("LOG_SQL_DSN")
	common.SQLitePath = filepath.Join(t.TempDir(), "billing-cny.db") + "?_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_txlock=immediate"
	common.IsMasterNode = true
	require.NoError(t, os.Unsetenv("SQL_DSN"))
	require.NoError(t, os.Unsetenv("LOG_SQL_DSN"))
	t.Cleanup(func() {
		DB, LOG_DB = previousDB, previousLogDB
		common.SetMainDatabaseType(previousMainType)
		common.SetLogDatabaseType(previousLogType)
		initCol()
		common.SQLitePath = previousPath
		common.IsMasterNode = previousMaster
		if mainWasSet {
			_ = os.Setenv("SQL_DSN", previousMainDSN)
		}
		if logWasSet {
			_ = os.Setenv("LOG_SQL_DSN", previousLogDSN)
		}
	})

	for range 2 {
		require.NoError(t, InitDB())
		require.NoError(t, InitLogDB())
		assertLedgerMarker(t, DB, billingLedgerMainScope)
		require.NoError(t, CloseDB())
	}
}

type legacyMigrationOption struct {
	Key   string `gorm:"primaryKey;type:varchar(255)"`
	Value string `gorm:"type:text"`
}

type legacyMigrationTopUp struct {
	Id              int `gorm:"primaryKey"`
	UserId          int `gorm:"index"`
	Amount          int64
	Money           float64
	TradeNo         string `gorm:"unique;type:varchar(255);index"`
	PaymentMethod   string `gorm:"type:varchar(50)"`
	PaymentProvider string `gorm:"type:varchar(50);default:''"`
	CreateTime      int64
	CompleteTime    int64
	Status          string
}

type legacyMigrationSubscriptionOrder struct {
	Id              int `gorm:"primaryKey"`
	UserId          int `gorm:"index"`
	PlanId          int `gorm:"index"`
	Money           float64
	TradeNo         string `gorm:"unique;type:varchar(255);index"`
	PaymentMethod   string `gorm:"type:varchar(50)"`
	PaymentProvider string `gorm:"type:varchar(50);default:''"`
	Status          string
	CreateTime      int64
	CompleteTime    int64
	ProviderPayload string `gorm:"type:text"`
}

func TestBillingCurrencyMigrationDatabaseMatrix(t *testing.T) {
	if os.Getenv("BILLING_CNY_INTEGRATION") != "1" {
		t.Skip("set BILLING_CNY_INTEGRATION=1 with dedicated migration database DSNs")
	}
	mainDSN := os.Getenv("BILLING_CNY_MAIN_DSN")
	logDSN := os.Getenv("BILLING_CNY_LOG_DSN")
	requireDedicatedMigrationDSN(t, mainDSN)
	requireDedicatedMigrationDSN(t, logDSN)

	previousMainDSN, mainWasSet := os.LookupEnv("SQL_DSN")
	previousLogDSN, logWasSet := os.LookupEnv("LOG_SQL_DSN")
	previousMaster := common.IsMasterNode
	require.NoError(t, os.Setenv("SQL_DSN", mainDSN))
	require.NoError(t, os.Setenv("LOG_SQL_DSN", logDSN))
	common.IsMasterNode = true
	t.Cleanup(func() {
		common.IsMasterNode = previousMaster
		if mainWasSet {
			_ = os.Setenv("SQL_DSN", previousMainDSN)
		} else {
			_ = os.Unsetenv("SQL_DSN")
		}
		if logWasSet {
			_ = os.Setenv("LOG_SQL_DSN", previousLogDSN)
		} else {
			_ = os.Unsetenv("LOG_SQL_DSN")
		}
	})

	t.Run("fresh startup twice", func(t *testing.T) {
		mainDB, mainType, err := chooseDB("SQL_DSN", false)
		require.NoError(t, err)
		logDB, logType, err := chooseDB("LOG_SQL_DSN", true)
		require.NoError(t, err)
		resetDedicatedMigrationDatabase(t, mainDB, mainType)
		resetDedicatedMigrationDatabase(t, logDB, logType)
		closeGormDB(t, mainDB)
		closeGormDB(t, logDB)

		require.NoError(t, InitDB())
		require.NoError(t, InitLogDB())
		assertLedgerMarker(t, DB, billingLedgerMainScope)
		assertLedgerMarker(t, LOG_DB, billingLedgerLogScope)
		require.NoError(t, CloseDB())

		require.NoError(t, InitDB())
		require.NoError(t, InitLogDB())
		assertLedgerMarker(t, DB, billingLedgerMainScope)
		assertLedgerMarker(t, LOG_DB, billingLedgerLogScope)
		require.NoError(t, CloseDB())
	})

	t.Run("legacy upgrade dry run apply and restart", func(t *testing.T) {
		require.NoError(t, InitBillingCurrencyMigrationDatabases())
		mainType := common.MainDatabaseType()
		logType := common.LogDatabaseType()
		resetDedicatedMigrationDatabase(t, DB, mainType)
		resetDedicatedMigrationDatabase(t, LOG_DB, logType)
		createLegacyMigrationFixture(t, DB, LOG_DB)

		_, err := MigrateLegacyUSDLedgerToCNY("", false)
		require.NoError(t, err)
		assert.False(t, DB.Migrator().HasColumn(&TopUp{}, "Currency"))
		assert.False(t, DB.Migrator().HasColumn(&TopUp{}, "CreditedQuota"))
		assert.False(t, DB.Migrator().HasColumn(&SubscriptionOrder{}, "Currency"))

		report, err := MigrateLegacyUSDLedgerToCNY("7.3", true)
		require.NoError(t, err)
		assert.True(t, report.Applied)
		assertLedgerMarker(t, DB, billingLedgerMainScope)
		assertLedgerMarker(t, LOG_DB, billingLedgerLogScope)
		var topup TopUp
		require.NoError(t, DB.First(&topup, 1).Error)
		assert.Equal(t, "CNY", topup.Currency)
		assert.Equal(t, int64(7_300_000), topup.CreditedQuota)
		var log Log
		require.NoError(t, LOG_DB.First(&log, 1).Error)
		assert.Equal(t, 73, log.Quota)

		again, err := MigrateLegacyUSDLedgerToCNY("7.3", true)
		require.NoError(t, err)
		assert.True(t, again.MainAlreadyMigrated)
		assert.True(t, again.LogAlreadyMigrated)
		require.NoError(t, CloseDB())

		require.NoError(t, InitDB())
		require.NoError(t, InitLogDB())
		require.NoError(t, CloseDB())
		require.NoError(t, InitDB())
		require.NoError(t, InitLogDB())
		require.NoError(t, CloseDB())
	})
}

func requireDedicatedMigrationDSN(t *testing.T, dsn string) {
	t.Helper()
	require.NoError(t, validateDedicatedMigrationDSN(dsn))
}

func validateDedicatedMigrationDSN(dsn string) error {
	dsn = strings.TrimSpace(dsn)
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			return fmt.Errorf("parse PostgreSQL integration DSN: %w", err)
		}
		database, err := url.PathUnescape(strings.TrimPrefix(parsed.EscapedPath(), "/"))
		if err != nil {
			return fmt.Errorf("parse PostgreSQL integration database: %w", err)
		}
		if parsed.Hostname() != "127.0.0.1" || !strings.HasPrefix(database, "robocoding_cny_migration_") {
			return fmt.Errorf("integration DSN must target a 127.0.0.1 robocoding_cny_migration_* database")
		}
		return nil
	}

	parsed, err := mysql.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("parse MySQL integration DSN: %w", err)
	}
	host, _, err := net.SplitHostPort(parsed.Addr)
	if err != nil {
		return fmt.Errorf("parse MySQL integration address: %w", err)
	}
	if parsed.Net != "tcp" || host != "127.0.0.1" || !strings.HasPrefix(parsed.DBName, "robocoding_cny_migration_") {
		return fmt.Errorf("integration DSN must target a 127.0.0.1 robocoding_cny_migration_* database")
	}
	return nil
}

func TestValidateDedicatedMigrationDSN(t *testing.T) {
	for _, tc := range []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{name: "mysql dedicated", dsn: "root:pass@tcp(127.0.0.1:33307)/robocoding_cny_migration_main"},
		{name: "postgres dedicated", dsn: "postgres://root:pass@127.0.0.1:35433/robocoding_cny_migration_main?sslmode=disable"},
		{name: "mysql marker only in query", dsn: "root:pass@tcp(10.0.0.2:3306)/production?note=127.0.0.1_robocoding_cny_migration_main", wantErr: true},
		{name: "postgres marker only in query", dsn: "postgres://root:pass@10.0.0.2:5432/production?note=127.0.0.1/robocoding_cny_migration_main", wantErr: true},
		{name: "mysql production database", dsn: "root:pass@tcp(127.0.0.1:33307)/production", wantErr: true},
		{name: "postgres production database", dsn: "postgres://root:pass@127.0.0.1:35433/production", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDedicatedMigrationDSN(tc.dsn)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func resetDedicatedMigrationDatabase(t *testing.T, db *gorm.DB, dbType common.DatabaseType) {
	t.Helper()
	switch dbType {
	case common.DatabaseTypePostgreSQL:
		require.NoError(t, db.Exec(`DROP SCHEMA public CASCADE`).Error)
		require.NoError(t, db.Exec(`CREATE SCHEMA public`).Error)
	case common.DatabaseTypeMySQL:
		require.NoError(t, db.Exec(`SET FOREIGN_KEY_CHECKS = 0`).Error)
		tables, err := db.Migrator().GetTables()
		require.NoError(t, err)
		for _, table := range tables {
			require.NoError(t, db.Migrator().DropTable(table))
		}
		require.NoError(t, db.Exec(`SET FOREIGN_KEY_CHECKS = 1`).Error)
	default:
		t.Fatalf("unsupported integration database type %s", dbType)
	}
}

func createLegacyMigrationFixture(t *testing.T, mainDB, logDB *gorm.DB) {
	t.Helper()
	require.NoError(t, mainDB.Table("options").AutoMigrate(&legacyMigrationOption{}))
	require.NoError(t, mainDB.Table("top_ups").AutoMigrate(&legacyMigrationTopUp{}))
	require.NoError(t, mainDB.Table("subscription_orders").AutoMigrate(&legacyMigrationSubscriptionOrder{}))
	require.NoError(t, mainDB.Table("options").Create(&legacyMigrationOption{Key: "USDExchangeRate", Value: "7.3"}).Error)
	require.NoError(t, mainDB.Table("options").Create(&legacyMigrationOption{Key: "QuotaPerUnit", Value: "1000000"}).Error)
	require.NoError(t, mainDB.Table("top_ups").Create(&legacyMigrationTopUp{
		Id: 1, UserId: 1, Amount: 2, Money: 14.6, TradeNo: "legacy-epay",
		PaymentProvider: PaymentProviderEpay, Status: common.TopUpStatusSuccess,
	}).Error)
	require.NoError(t, mainDB.Table("subscription_orders").Create(&legacyMigrationSubscriptionOrder{
		Id: 1, UserId: 1, PlanId: 1, Money: 10, TradeNo: "legacy-sub-epay",
		PaymentProvider: PaymentProviderEpay, Status: common.TopUpStatusSuccess,
	}).Error)
	require.NoError(t, logDB.AutoMigrate(&Log{}))
	require.NoError(t, logDB.Create(&Log{Id: 1, Quota: 20, Other: `{"actual_quota":20}`}).Error)
}

func assertLedgerMarker(t *testing.T, db *gorm.DB, scope string) {
	t.Helper()
	var marker BillingLedgerVersion
	require.NoError(t, db.Where("scope = ?", scope).First(&marker).Error)
	assert.Equal(t, billingLedgerCurrency, marker.Currency)
	assert.Equal(t, billingLedgerVersion, marker.Version)
}

func closeGormDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}
