package model

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	billingLedgerCurrency        = "CNY"
	billingLedgerVersion         = 1
	billingLedgerMainScope       = "main"
	billingLedgerLogScope        = "log"
	legacyDefaultUSDExchangeRate = 7.3
)

var ErrBillingLedgerMigrationRequired = errors.New("legacy USD billing ledger requires explicit offline CNY migration")

type BillingLedgerVersion struct {
	Scope              string `gorm:"primaryKey;type:varchar(16)"`
	Currency           string `gorm:"type:varchar(8);not null"`
	Version            int    `gorm:"not null"`
	LegacyUSDCNYRate   string `gorm:"type:varchar(64);not null"`
	LegacyQuotaPerUnit string `gorm:"type:varchar(64);not null"`
	MigrationUnixTime  int64  `gorm:"type:bigint;not null"`
}

type BillingCurrencyMigrationReport struct {
	Applied             bool             `json:"applied"`
	LegacyUSDCNYRate    string           `json:"legacy_usd_cny_rate"`
	LegacyQuotaPerUnit  string           `json:"legacy_quota_per_unit"`
	QuotaScaleFactor    string           `json:"quota_scale_factor"`
	RateSource          string           `json:"rate_source"`
	MainAlreadyMigrated bool             `json:"main_already_migrated"`
	LogAlreadyMigrated  bool             `json:"log_already_migrated"`
	UpdatedRows         map[string]int64 `json:"updated_rows"`
}

func ensureBillingLedgerReady(db *gorm.DB, scope string, legacyTables ...string) (bool, error) {
	if db == nil {
		return false, errors.New("billing ledger database is nil")
	}
	if db.Migrator().HasTable(&BillingLedgerVersion{}) {
		var marker BillingLedgerVersion
		err := db.Where("scope = ?", scope).First(&marker).Error
		if err == nil {
			if marker.Currency != billingLedgerCurrency || marker.Version != billingLedgerVersion {
				return false, fmt.Errorf("unsupported billing ledger marker %s/%d for %s", marker.Currency, marker.Version, scope)
			}
			return false, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, err
		}
	}
	for _, table := range legacyTables {
		if db.Migrator().HasTable(table) {
			return false, fmt.Errorf("%w: %s database contains legacy table %s", ErrBillingLedgerMigrationRequired, scope, table)
		}
	}
	return true, nil
}

func writeBillingLedgerMarker(db *gorm.DB, scope, rate, oldQuotaPerUnit string) error {
	marker := BillingLedgerVersion{
		Scope: scope, Currency: billingLedgerCurrency, Version: billingLedgerVersion,
		LegacyUSDCNYRate: rate, LegacyQuotaPerUnit: oldQuotaPerUnit, MigrationUnixTime: time.Now().Unix(),
	}
	return db.Create(&marker).Error
}

func InitBillingCurrencyMigrationDatabases() error {
	db, dbType, err := chooseDB("SQL_DSN", false)
	if err != nil {
		return err
	}
	DB = db
	common.SetMainDatabaseType(dbType)
	if osLogDSN := strings.TrimSpace(getenv("LOG_SQL_DSN")); osLogDSN == "" {
		LOG_DB = DB
		common.SetLogDatabaseType(dbType)
	} else {
		LOG_DB, dbType, err = chooseDB("LOG_SQL_DSN", true)
		if err != nil {
			return err
		}
		if dbType == common.DatabaseTypeClickHouse {
			return errors.New("CNY ledger migration does not support a ClickHouse log database")
		}
		common.SetLogDatabaseType(dbType)
	}
	initCol()
	return nil
}

// getenv is replaceable in tests without exposing general process setup here.
var getenv = os.Getenv

func MigrateLegacyUSDLedgerToCNY(expectedRate string, apply bool) (*BillingCurrencyMigrationReport, error) {
	if DB == nil || LOG_DB == nil {
		return nil, errors.New("billing migration databases are not initialized")
	}
	rate, oldQuotaPerUnit, quotaFactor, source, err := legacyBillingFactors(DB)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(expectedRate) != "" {
		expected, parseErr := decimal.NewFromString(strings.TrimSpace(expectedRate))
		if parseErr != nil || !expected.Equal(rate) {
			return nil, fmt.Errorf("expected legacy USD/CNY rate %q does not match captured rate %s", expectedRate, rate.String())
		}
	}
	if apply && strings.TrimSpace(expectedRate) == "" {
		return nil, errors.New("--rate is required with --apply; use the exact rate shown by the dry-run")
	}
	report := &BillingCurrencyMigrationReport{
		Applied: apply, LegacyUSDCNYRate: rate.String(), LegacyQuotaPerUnit: oldQuotaPerUnit.String(), QuotaScaleFactor: quotaFactor.String(), RateSource: source,
		UpdatedRows: make(map[string]int64),
	}
	mainDone, err := billingLedgerMarkerExists(DB, billingLedgerMainScope)
	if err != nil {
		return nil, err
	}
	report.MainAlreadyMigrated = mainDone
	logDone := mainDone
	if LOG_DB != DB {
		logDone, err = billingLedgerMarkerExists(LOG_DB, billingLedgerLogScope)
		if err != nil {
			return nil, err
		}
		if logDone && !mainDone {
			var marker BillingLedgerVersion
			if err := LOG_DB.Where("scope = ?", billingLedgerLogScope).First(&marker).Error; err != nil {
				return nil, err
			}
			if marker.LegacyUSDCNYRate != rate.String() || marker.LegacyQuotaPerUnit != oldQuotaPerUnit.String() {
				return nil, fmt.Errorf(
					"separate log ledger was migrated with legacy rate %s and quota-per-unit %s, but main ledger now reports %s and %s",
					marker.LegacyUSDCNYRate, marker.LegacyQuotaPerUnit, rate.String(), oldQuotaPerUnit.String(),
				)
			}
		}
	}
	report.LogAlreadyMigrated = logDone
	if mainDone && logDone {
		return report, nil
	}
	if mainDone && !logDone {
		return nil, errors.New("main ledger is already CNY but the separate log ledger is not; rerun with the original rate recorded in the main marker")
	}
	if err := rejectUnsettledBillingState(DB); err != nil {
		return nil, err
	}
	if !apply {
		if LOG_DB != DB && !logDone {
			if err := validateLogLedgerMigration(LOG_DB, quotaFactor, rate, report); err != nil {
				return nil, err
			}
		}
		if err := validateMainLedgerMigration(DB, quotaFactor, rate, oldQuotaPerUnit, LOG_DB == DB, report); err != nil {
			return nil, err
		}
		return report, nil
	}
	// Validate both ledgers with the exact migration logic before committing
	// either database. This catches deterministic data errors before a separate
	// log database can be marked migrated while the main ledger still fails.
	validationReport := &BillingCurrencyMigrationReport{UpdatedRows: make(map[string]int64)}
	if LOG_DB != DB && !logDone {
		if err := validateLogLedgerMigration(LOG_DB, quotaFactor, rate, validationReport); err != nil {
			return nil, err
		}
	}
	if err := validateMainLedgerMigration(DB, quotaFactor, rate, oldQuotaPerUnit, LOG_DB == DB, validationReport); err != nil {
		return nil, err
	}
	if err := prepareMainLedgerMigrationSchema(DB); err != nil {
		return nil, err
	}
	if LOG_DB != DB && !logDone {
		if err := migrateLogLedger(LOG_DB, quotaFactor, rate, oldQuotaPerUnit, billingLedgerLogScope, report); err != nil {
			return nil, err
		}
	}
	if err := migrateMainLedger(DB, quotaFactor, rate, oldQuotaPerUnit, LOG_DB == DB, report); err != nil {
		return nil, err
	}
	return report, nil
}

func prepareMainLedgerMigrationSchema(db *gorm.DB) error {
	columns := []struct {
		model any
		name  string
	}{
		{model: &TopUp{}, name: "CreditedQuota"},
		{model: &TopUp{}, name: "Currency"},
		{model: &SubscriptionOrder{}, name: "Currency"},
	}
	for _, column := range columns {
		if !db.Migrator().HasTable(column.model) || db.Migrator().HasColumn(column.model, column.name) {
			continue
		}
		if err := db.Migrator().AddColumn(column.model, column.name); err != nil {
			return fmt.Errorf("add billing migration column %s: %w", column.name, err)
		}
	}
	return nil
}

func billingLedgerMarkerExists(db *gorm.DB, scope string) (bool, error) {
	if !db.Migrator().HasTable(&BillingLedgerVersion{}) {
		return false, nil
	}
	var count int64
	if err := db.Model(&BillingLedgerVersion{}).Where("scope = ? AND currency = ? AND version = ?", scope, billingLedgerCurrency, billingLedgerVersion).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 1, nil
}

func legacyBillingFactors(db *gorm.DB) (decimal.Decimal, decimal.Decimal, decimal.Decimal, string, error) {
	if db.Migrator().HasTable(&BillingLedgerVersion{}) {
		var marker BillingLedgerVersion
		if err := db.Where("scope = ?", billingLedgerMainScope).First(&marker).Error; err == nil {
			rate, parseErr := decimal.NewFromString(marker.LegacyUSDCNYRate)
			if parseErr != nil {
				return decimal.Zero, decimal.Zero, decimal.Zero, "", parseErr
			}
			oldQuotaPerUnit, parseErr := decimal.NewFromString(marker.LegacyQuotaPerUnit)
			if parseErr != nil {
				return decimal.Zero, decimal.Zero, decimal.Zero, "", parseErr
			}
			return rate, oldQuotaPerUnit, rate.Mul(decimal.NewFromInt(500000)).Div(oldQuotaPerUnit), "migration_marker", nil
		}
	}
	rate := decimal.NewFromFloat(legacyDefaultUSDExchangeRate)
	oldQuotaPerUnit := decimal.NewFromInt(500000)
	source := "historical_default"
	if db.Migrator().HasTable(&Option{}) {
		var option Option
		err := db.Where(commonKeyCol+" = ?", "USDExchangeRate").First(&option).Error
		if err == nil {
			parsed, parseErr := decimal.NewFromString(strings.TrimSpace(option.Value))
			if parseErr != nil {
				return decimal.Zero, decimal.Zero, decimal.Zero, "", fmt.Errorf("invalid stored USDExchangeRate: %w", parseErr)
			}
			rate, source = parsed, "option"
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return decimal.Zero, decimal.Zero, decimal.Zero, "", err
		}
		var quotaOption Option
		err = db.Where(commonKeyCol+" = ?", "QuotaPerUnit").First(&quotaOption).Error
		if err == nil {
			parsed, parseErr := decimal.NewFromString(strings.TrimSpace(quotaOption.Value))
			if parseErr != nil {
				return decimal.Zero, decimal.Zero, decimal.Zero, "", fmt.Errorf("invalid stored QuotaPerUnit: %w", parseErr)
			}
			oldQuotaPerUnit = parsed
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return decimal.Zero, decimal.Zero, decimal.Zero, "", err
		}
	}
	value, _ := rate.Float64()
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", errors.New("legacy USDExchangeRate must be finite and positive")
	}
	quotaValue, _ := oldQuotaPerUnit.Float64()
	if quotaValue <= 0 || math.IsNaN(quotaValue) || math.IsInf(quotaValue, 0) {
		return decimal.Zero, decimal.Zero, decimal.Zero, "", errors.New("legacy QuotaPerUnit must be finite and positive")
	}
	return rate, oldQuotaPerUnit, rate.Mul(decimal.NewFromInt(500000)).Div(oldQuotaPerUnit), source, nil
}

func rejectUnsettledBillingState(db *gorm.DB) error {
	checks := []struct {
		model any
		where string
		args  []any
		name  string
	}{
		{&TopUp{}, "status = ?", []any{common.TopUpStatusPending}, "pending top-up"},
		{&SubscriptionOrder{}, "status = ?", []any{common.TopUpStatusPending}, "pending subscription order"},
		{&TeamQuotaReservation{}, "status = ?", []any{TeamReservationReserved}, "reserved team quota"},
		{&Task{}, "status NOT IN ?", []any{[]TaskStatus{TaskStatusSuccess, TaskStatusFailure}}, "non-terminal task"},
	}
	for _, check := range checks {
		if !db.Migrator().HasTable(check.model) {
			continue
		}
		var count int64
		if err := db.Model(check.model).Where(check.where, check.args...).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("billing migration blocked by %d %s record(s)", count, check.name)
		}
	}
	return nil
}

func scaleQuota(value int, rate decimal.Decimal) (int, error) {
	scaled := decimal.NewFromInt(int64(value)).Mul(rate).Round(0)
	if !scaled.IsInteger() || scaled.GreaterThan(decimal.NewFromInt(math.MaxInt64)) || scaled.LessThan(decimal.NewFromInt(math.MinInt64)) {
		return 0, errors.New("scaled quota overflows int64")
	}
	result := scaled.IntPart()
	if strconv.IntSize == 32 && (result > math.MaxInt32 || result < math.MinInt32) {
		return 0, errors.New("scaled quota overflows int")
	}
	return int(result), nil
}

func scaleWalletQuota(value int, rate decimal.Decimal) (int, error) {
	scaled, err := scaleQuota(value, rate)
	if err != nil {
		return 0, err
	}
	if scaled > common.MaxWalletQuota || scaled < -common.MaxWalletQuota {
		return 0, fmt.Errorf("scaled wallet quota exceeds %d", common.MaxWalletQuota)
	}
	return scaled, nil
}

func scaleQuota64(value int64, rate decimal.Decimal) (int64, error) {
	scaled := decimal.NewFromInt(value).Mul(rate).Round(0)
	if !scaled.IsInteger() || scaled.GreaterThan(decimal.NewFromInt(math.MaxInt64)) || scaled.LessThan(decimal.NewFromInt(math.MinInt64)) {
		return 0, errors.New("scaled quota overflows int64")
	}
	return scaled.IntPart(), nil
}

func scaleMoney(value float64, rate decimal.Decimal) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, errors.New("stored monetary value must be finite")
	}
	scaled := decimal.NewFromFloat(value).Mul(rate)
	result, _ := scaled.Float64()
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, errors.New("scaled monetary value cannot be represented")
	}
	return result, nil
}

func scaleExpression(expression string, rate decimal.Decimal) (string, error) {
	factor, _ := rate.Float64()
	return billingexpr.ScaleCurrency(expression, factor)
}
