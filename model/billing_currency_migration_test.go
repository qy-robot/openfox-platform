package model

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func useBillingMigrationSQLite(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	previousDB, previousLogDB := DB, LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	DB, LOG_DB = db, db
	common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	common.SetLogDatabaseType(common.DatabaseTypeSQLite)
	initCol()
	t.Cleanup(func() {
		DB, LOG_DB = previousDB, previousLogDB
		common.SetMainDatabaseType(previousMainType)
		common.SetLogDatabaseType(previousLogType)
		initCol()
		sqlDB, sqlErr := db.DB()
		if sqlErr == nil {
			require.NoError(t, sqlDB.Close())
		}
	})
	return db
}

func migrateFixtureSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.AutoMigrate(
		&User{}, &Token{}, &Channel{}, &Team{}, &TeamMember{}, &TeamMonthlyUsage{},
		&TeamQuotaReservation{}, &TeamQuotaTransfer{}, &Redemption{}, &Checkin{},
		&Midjourney{}, &QuotaData{}, &Task{}, &SubscriptionPlan{}, &SubscriptionOrder{},
		&UserSubscription{}, &SubscriptionPreConsumeRecord{}, &TopUp{}, &Log{}, &Option{},
	))
}

func TestLegacyUSDLedgerMigrationPreservesValueAndIsIdempotent(t *testing.T) {
	db := useBillingMigrationSQLite(t)
	migrateFixtureSchema(t, db)
	require.NoError(t, db.Create(&Option{Key: "USDExchangeRate", Value: "7.3"}).Error)
	require.NoError(t, db.Create(&Option{Key: "QuotaPerUnit", Value: "1000000"}).Error)
	require.NoError(t, db.Create(&Option{Key: "ModelPrice", Value: `{"fixed":0.00123456789}`}).Error)
	require.NoError(t, db.Create(&Option{Key: "ModelRatio", Value: `{"token":2}`}).Error)
	require.NoError(t, db.Create(&Option{Key: "billing_setting.billing_expr", Value: `{"tiered":"tier(\"base\", p * 2 + c * 8)"}`}).Error)
	require.NoError(t, db.Create(&Option{Key: "CreemProducts", Value: `[{"productId":"usd-product","price":5,"currency":"USD","quota":100}]`}).Error)
	require.NoError(t, db.Create(&User{Id: 1, Username: "legacy", Password: "x", Quota: 100, UsedQuota: 20, AffQuota: 10, AffHistoryQuota: 5}).Error)
	require.NoError(t, db.Create(&Token{Id: 1, UserId: 1, Key: "legacy-key", RemainQuota: 40, UsedQuota: 60}).Error)
	require.NoError(t, db.Create(&SubscriptionPlan{Id: 1, Title: "legacy", PriceAmount: 10, Currency: "USD", TotalAmount: 100}).Error)
	require.NoError(t, db.Create(&SubscriptionOrder{Id: 1, UserId: 1, PlanId: 1, Money: 10, TradeNo: "epay-sub", PaymentProvider: PaymentProviderEpay, Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, db.Create(&SubscriptionOrder{Id: 2, UserId: 1, PlanId: 1, Money: 2, TradeNo: "stripe-sub", PaymentProvider: PaymentProviderStripe, Status: common.TopUpStatusFailed}).Error)
	require.NoError(t, db.Create(&SubscriptionOrder{Id: 3, UserId: 1, PlanId: 1, Money: 2, TradeNo: "balance-sub", PaymentProvider: PaymentProviderBalance, Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, db.Create(&TopUp{Id: 1, UserId: 1, Amount: 2, Money: 14.6, TradeNo: "epay-topup", PaymentProvider: PaymentProviderEpay, Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, db.Create(&TopUp{Id: 2, UserId: 1, Amount: 2, Money: 3, TradeNo: "stripe-topup", PaymentProvider: PaymentProviderStripe, Status: common.TopUpStatusSuccess}).Error)
	require.NoError(t, db.Create(&TopUp{Id: 3, UserId: 1, Money: 10, TradeNo: "epay-sub", Status: common.TopUpStatusSuccess}).Error)
	snapshot := &billingexpr.BillingSnapshot{ExprString: `tier("base", p * 2)`, ExprHash: billingexpr.ExprHashString(`tier("base", p * 2)`), QuotaPerUnit: 1_000_000, EstimatedQuotaBeforeGroup: 20, EstimatedQuotaAfterGroup: 20}
	task := Task{ID: 1, TaskID: "done", Status: TaskStatusSuccess, Quota: 20, PrivateData: TaskPrivateData{BillingContext: &TaskBillingContext{ModelPrice: 0.25, ModelRatio: 2, TieredSnapshot: snapshot}}}
	require.NoError(t, db.Create(&task).Error)
	require.NoError(t, db.Create(&Log{Id: 1, Quota: 20, Other: `{"model_price":0.25,"model_ratio":2,"actual_quota":20}`}).Error)

	dryRun, err := MigrateLegacyUSDLedgerToCNY("", false)
	require.NoError(t, err)
	assert.Equal(t, "7.3", dryRun.LegacyUSDCNYRate)
	assert.Equal(t, "1000000", dryRun.LegacyQuotaPerUnit)
	assert.Equal(t, "3.65", dryRun.QuotaScaleFactor)

	report, err := MigrateLegacyUSDLedgerToCNY("7.3", true)
	require.NoError(t, err)
	assert.True(t, report.Applied)
	var user User
	require.NoError(t, db.First(&user, 1).Error)
	assert.Equal(t, 365, user.Quota)
	assert.Equal(t, 73, user.UsedQuota)
	var plan SubscriptionPlan
	require.NoError(t, db.First(&plan, 1).Error)
	assert.Equal(t, "CNY", plan.Currency)
	assert.InDelta(t, 73, plan.PriceAmount, 1e-9)
	assert.Equal(t, int64(365), plan.TotalAmount)
	var orders []SubscriptionOrder
	require.NoError(t, db.Order("id").Find(&orders).Error)
	assert.Equal(t, "CNY", orders[0].Currency)
	assert.InDelta(t, 10, orders[0].Money, 1e-12)
	assert.Equal(t, "USD", orders[1].Currency)
	assert.InDelta(t, 2, orders[1].Money, 1e-12)
	assert.Equal(t, "CNY", orders[2].Currency)
	assert.InDelta(t, 14.6, orders[2].Money, 1e-12)
	var topups []TopUp
	require.NoError(t, db.Order("id").Find(&topups).Error)
	assert.Equal(t, "CNY", topups[0].Currency)
	assert.Equal(t, int64(7_300_000), topups[0].CreditedQuota)
	assert.InDelta(t, 14.6, topups[0].Money, 1e-12)
	assert.Equal(t, "USD", topups[1].Currency)
	assert.Equal(t, int64(10_950_000), topups[1].CreditedQuota)
	assert.InDelta(t, 3, topups[1].Money, 1e-12)
	assert.Equal(t, "CNY", topups[2].Currency)
	assert.Zero(t, topups[2].CreditedQuota)
	var migratedTask Task
	require.NoError(t, db.First(&migratedTask, 1).Error)
	assert.Equal(t, 73, migratedTask.Quota)
	assert.InDelta(t, 1.825, migratedTask.PrivateData.BillingContext.ModelPrice, 1e-12)
	assert.InDelta(t, 7.3, migratedTask.PrivateData.BillingContext.ModelRatio, 1e-12)
	assert.Equal(t, float64(500000), migratedTask.PrivateData.BillingContext.TieredSnapshot.QuotaPerUnit)
	var price Option
	require.NoError(t, db.Where("key = ?", "ModelPrice").First(&price).Error)
	assert.JSONEq(t, `{"fixed":0.009012345597}`, price.Value)
	var ratio Option
	require.NoError(t, db.Where("key = ?", "ModelRatio").First(&ratio).Error)
	var ratioValues map[string]float64
	require.NoError(t, common.UnmarshalJsonStr(ratio.Value, &ratioValues))
	assert.Equal(t, 7.3, ratioValues["token"])
	assert.Equal(t, 3.65, ratioValues["gpt-4.1"])
	var creemProducts Option
	require.NoError(t, db.Where("key = ?", "CreemProducts").First(&creemProducts).Error)
	assert.JSONEq(t, `[{"productId":"usd-product","price":5,"currency":"USD","quota":365}]`, creemProducts.Value)

	again, err := MigrateLegacyUSDLedgerToCNY("7.3", true)
	require.NoError(t, err)
	assert.True(t, again.MainAlreadyMigrated)
	require.NoError(t, db.First(&user, 1).Error)
	assert.Equal(t, 365, user.Quota)
}

func TestLegacyUSDLedgerMigrationBlocksPendingState(t *testing.T) {
	for _, tc := range []struct {
		name    string
		wantErr string
		insert  func(*gorm.DB) error
	}{
		{name: "task", wantErr: "non-terminal task", insert: func(db *gorm.DB) error {
			return db.Create(&Task{TaskID: "pending", Status: TaskStatusInProgress}).Error
		}},
		{name: "top-up", wantErr: "pending top-up", insert: func(db *gorm.DB) error {
			return db.Create(&TopUp{TradeNo: "pending", PaymentProvider: PaymentProviderEpay, Status: common.TopUpStatusPending}).Error
		}},
		{name: "subscription order", wantErr: "pending subscription order", insert: func(db *gorm.DB) error {
			return db.Create(&SubscriptionOrder{TradeNo: "pending", PaymentProvider: PaymentProviderEpay, Status: common.TopUpStatusPending}).Error
		}},
		{name: "team reservation", wantErr: "reserved team quota", insert: func(db *gorm.DB) error {
			return db.Create(&TeamQuotaReservation{Status: TeamReservationReserved, Quota: 20}).Error
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := useBillingMigrationSQLite(t)
			migrateFixtureSchema(t, db)
			require.NoError(t, db.Create(&Option{Key: "USDExchangeRate", Value: "7.3"}).Error)
			require.NoError(t, db.Create(&User{Id: 1, Username: "blocked", Password: "x", Quota: 100}).Error)
			require.NoError(t, tc.insert(db))
			_, err := MigrateLegacyUSDLedgerToCNY("7.3", true)
			assert.ErrorContains(t, err, tc.wantErr)
			assert.False(t, db.Migrator().HasTable(&BillingLedgerVersion{}))
			var user User
			require.NoError(t, db.First(&user, 1).Error)
			assert.Equal(t, 100, user.Quota)
		})
	}
}

func TestLegacyUSDLedgerMigrationRejectsUnsafePaymentHistory(t *testing.T) {
	for _, tc := range []struct {
		name     string
		provider string
		wantErr  string
	}{
		{name: "ambiguous provider", provider: PaymentProviderCreem, wantErr: `payment provider "creem" has no safely inferable historical currency`},
		{name: "unknown provider", provider: "future-provider", wantErr: `payment provider "future-provider" has no safely inferable historical currency`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := useBillingMigrationSQLite(t)
			migrateFixtureSchema(t, db)
			require.NoError(t, db.Create(&Option{Key: "USDExchangeRate", Value: "7.3"}).Error)
			require.NoError(t, db.Create(&TopUp{UserId: 1, TradeNo: tc.name, PaymentProvider: tc.provider, Status: common.TopUpStatusFailed}).Error)

			_, err := MigrateLegacyUSDLedgerToCNY("7.3", true)
			assert.ErrorContains(t, err, tc.wantErr)
			assert.False(t, db.Migrator().HasTable(&BillingLedgerVersion{}))
		})
	}
}

func TestLegacyUSDLedgerMigrationAddsPaymentSnapshotColumnsOnlyOnApply(t *testing.T) {
	db := useBillingMigrationSQLite(t)
	require.NoError(t, db.Exec(`CREATE TABLE options (key TEXT PRIMARY KEY, value TEXT)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO options (key, value) VALUES ('USDExchangeRate', '7.3'), ('QuotaPerUnit', '1000000')`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE subscription_orders (
		id INTEGER PRIMARY KEY, user_id INTEGER, plan_id INTEGER, money REAL, trade_no TEXT,
		payment_method TEXT, payment_provider TEXT, status TEXT, create_time INTEGER,
		complete_time INTEGER, provider_payload TEXT
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE top_ups (
		id INTEGER PRIMARY KEY, user_id INTEGER, amount INTEGER, money REAL, trade_no TEXT,
		payment_method TEXT, payment_provider TEXT, create_time INTEGER, complete_time INTEGER, status TEXT
	)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO subscription_orders
		(id, user_id, plan_id, money, trade_no, payment_provider, status)
		VALUES (1, 1, 1, 10, 'sub-epay', 'epay', 'success')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO top_ups
		(id, user_id, amount, money, trade_no, payment_provider, status)
		VALUES (1, 1, 2, 14.6, 'wallet-epay', 'epay', 'success')`).Error)

	_, err := MigrateLegacyUSDLedgerToCNY("", false)
	require.NoError(t, err)
	assert.False(t, db.Migrator().HasColumn(&TopUp{}, "Currency"))
	assert.False(t, db.Migrator().HasColumn(&TopUp{}, "CreditedQuota"))
	assert.False(t, db.Migrator().HasColumn(&SubscriptionOrder{}, "Currency"))

	_, err = MigrateLegacyUSDLedgerToCNY("7.3", true)
	require.NoError(t, err)
	assert.True(t, db.Migrator().HasColumn(&TopUp{}, "Currency"))
	assert.True(t, db.Migrator().HasColumn(&TopUp{}, "CreditedQuota"))
	assert.True(t, db.Migrator().HasColumn(&SubscriptionOrder{}, "Currency"))
	var topup TopUp
	require.NoError(t, db.First(&topup, 1).Error)
	assert.Equal(t, "CNY", topup.Currency)
	assert.Equal(t, int64(7_300_000), topup.CreditedQuota)
	var order SubscriptionOrder
	require.NoError(t, db.First(&order, 1).Error)
	assert.Equal(t, "CNY", order.Currency)
	assert.InDelta(t, 10, order.Money, 1e-12)
}

func TestLegacyUSDLedgerMigrationRollsBackInvalidExpression(t *testing.T) {
	db := useBillingMigrationSQLite(t)
	migrateFixtureSchema(t, db)
	require.NoError(t, db.Create(&Option{Key: "USDExchangeRate", Value: "7.3"}).Error)
	require.NoError(t, db.Create(&Option{Key: "billing_setting.billing_expr", Value: `{"unsafe":"p * 2"}`}).Error)
	require.NoError(t, db.Create(&User{Id: 1, Username: "legacy", Password: "x", Quota: 100}).Error)
	_, err := MigrateLegacyUSDLedgerToCNY("7.3", true)
	assert.ErrorContains(t, err, "no tier")
	var user User
	require.NoError(t, db.First(&user, 1).Error)
	assert.Equal(t, 100, user.Quota)
	assert.False(t, db.Migrator().HasTable(&BillingLedgerVersion{}))
}

func TestLegacyUSDLedgerMigrationScalesNestedLogBillingData(t *testing.T) {
	expression := `tier("base", p * 0.25)`
	encodedExpression := base64.StdEncoding.EncodeToString([]byte(expression))
	other := `{"expr_b64":"` + encodedExpression + `","audio_input_seperate_price":0.5,"tool_surcharges":[{"name":"search","count":2,"price":0.1,"quota":20}],"nested":[[{"actual_quota":40}]]}`

	scaled, err := scaleLogOther(other, decimal.RequireFromString("3.65"), decimal.RequireFromString("7.3"))
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, common.UnmarshalJsonStr(scaled, &result))
	assert.InDelta(t, 3.65, result["audio_input_seperate_price"], 1e-12)
	surcharges := result["tool_surcharges"].([]any)
	surcharge := surcharges[0].(map[string]any)
	assert.InDelta(t, 0.73, surcharge["price"], 1e-12)
	assert.Equal(t, float64(73), surcharge["quota"])
	assert.Equal(t, float64(2), surcharge["count"])
	nested := result["nested"].([]any)[0].([]any)[0].(map[string]any)
	assert.Equal(t, float64(146), nested["actual_quota"])
	decoded, decodeErr := base64.StdEncoding.DecodeString(result["expr_b64"].(string))
	require.NoError(t, decodeErr)
	assert.Equal(t, `tier("base", p * 0.25 * 7.3)`, string(decoded))
}

func TestLegacyUSDLedgerMigrationRollsBackWalletOverflow(t *testing.T) {
	db := useBillingMigrationSQLite(t)
	migrateFixtureSchema(t, db)
	require.NoError(t, db.Create(&Option{Key: "USDExchangeRate", Value: "7.3"}).Error)
	originalQuota := common.MaxWalletQuota / 2
	require.NoError(t, db.Create(&User{Id: 1, Username: "overflow", Password: "x", Quota: originalQuota}).Error)

	_, err := MigrateLegacyUSDLedgerToCNY("7.3", true)
	assert.ErrorContains(t, err, "scaled wallet quota exceeds")
	var user User
	require.NoError(t, db.First(&user, 1).Error)
	assert.Equal(t, originalQuota, user.Quota)
	assert.False(t, db.Migrator().HasTable(&BillingLedgerVersion{}))
}

func TestBillingLedgerStartupGuardRejectsExistingUnmarkedDatabase(t *testing.T) {
	db := useBillingMigrationSQLite(t)
	require.NoError(t, db.AutoMigrate(&User{}))
	_, err := ensureBillingLedgerReady(db, billingLedgerMainScope, "users")
	assert.True(t, errors.Is(err, ErrBillingLedgerMigrationRequired))
}

func TestBillingLedgerSeparateLogRetryRequiresSameFactors(t *testing.T) {
	for _, tc := range []struct {
		name       string
		markerRate string
		wantError  bool
	}{
		{name: "matching marker resumes main migration", markerRate: "7.3"},
		{name: "mismatched marker is rejected", markerRate: "8", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mainDB := useBillingMigrationSQLite(t)
			migrateFixtureSchema(t, mainDB)
			require.NoError(t, mainDB.Create(&Option{Key: "USDExchangeRate", Value: "7.3"}).Error)
			require.NoError(t, mainDB.Create(&User{Id: 1, Username: "retry", Password: "x", Quota: 100}).Error)

			logDB, err := gorm.Open(sqlite.Open("file:"+t.Name()+"-log?mode=memory&cache=shared"), &gorm.Config{})
			require.NoError(t, err)
			require.NoError(t, logDB.AutoMigrate(&BillingLedgerVersion{}, &Log{}))
			require.NoError(t, writeBillingLedgerMarker(logDB, billingLedgerLogScope, tc.markerRate, "500000"))
			LOG_DB = logDB
			common.SetLogDatabaseType(common.DatabaseTypeSQLite)
			t.Cleanup(func() { closeGormDB(t, logDB) })

			report, err := MigrateLegacyUSDLedgerToCNY("7.3", true)
			if tc.wantError {
				assert.ErrorContains(t, err, "separate log ledger was migrated with legacy rate")
				assert.False(t, mainDB.Migrator().HasTable(&BillingLedgerVersion{}))
				return
			}
			require.NoError(t, err)
			assert.True(t, report.LogAlreadyMigrated)
			var user User
			require.NoError(t, mainDB.First(&user, 1).Error)
			assert.Equal(t, 730, user.Quota)
			assertLedgerMarker(t, mainDB, billingLedgerMainScope)
		})
	}
}

func TestLegacyBillingFactorsAcceptDecimalRate(t *testing.T) {
	db := useBillingMigrationSQLite(t)
	require.NoError(t, db.AutoMigrate(&Option{}))
	require.NoError(t, db.Create(&Option{Key: "USDExchangeRate", Value: "7.3"}).Error)
	rate, oldQuotaPerUnit, quotaFactor, _, err := legacyBillingFactors(db)
	require.NoError(t, err)
	assert.Equal(t, "7.3", rate.String())
	assert.Equal(t, "500000", oldQuotaPerUnit.String())
	assert.Equal(t, "7.3", quotaFactor.String())
}
