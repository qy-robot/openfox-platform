package model

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var errBillingMigrationDryRunRollback = errors.New("billing migration dry-run rollback")

func migrateMainLedger(db *gorm.DB, quotaFactor, priceFactor, oldQuotaPerUnit decimal.Decimal, includeLogs bool, report *BillingCurrencyMigrationReport) error {
	if err := db.AutoMigrate(&BillingLedgerVersion{}); err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := migrateMainLedgerTx(tx, quotaFactor, priceFactor, oldQuotaPerUnit, includeLogs, report); err != nil {
			return err
		}
		return writeBillingLedgerMarker(tx, billingLedgerMainScope, priceFactor.String(), oldQuotaPerUnit.String())
	})
}

func validateMainLedgerMigration(db *gorm.DB, quotaFactor, priceFactor, oldQuotaPerUnit decimal.Decimal, includeLogs bool, report *BillingCurrencyMigrationReport) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := migrateMainLedgerTx(tx, quotaFactor, priceFactor, oldQuotaPerUnit, includeLogs, report); err != nil {
			return err
		}
		return errBillingMigrationDryRunRollback
	})
	if errors.Is(err, errBillingMigrationDryRunRollback) {
		return nil
	}
	return err
}

func migrateMainLedgerTx(tx *gorm.DB, quotaFactor, priceFactor, oldQuotaPerUnit decimal.Decimal, includeLogs bool, report *BillingCurrencyMigrationReport) error {
	if err := migrateUsers(tx, quotaFactor, report); err != nil {
		return err
	}
	if err := migrateTokens(tx, quotaFactor, report); err != nil {
		return err
	}
	if err := migrateChannels(tx, quotaFactor, report); err != nil {
		return err
	}
	if err := migrateTeams(tx, quotaFactor, report); err != nil {
		return err
	}
	if err := migrateRedemptionsAndAwards(tx, quotaFactor, report); err != nil {
		return err
	}
	if err := migrateUsageAndTasks(tx, quotaFactor, priceFactor, report); err != nil {
		return err
	}
	if err := migrateSubscriptions(tx, quotaFactor, priceFactor, report); err != nil {
		return err
	}
	if err := migrateTopUps(tx, quotaFactor, oldQuotaPerUnit, report); err != nil {
		return err
	}
	if err := migrateBillingOptions(tx, quotaFactor, priceFactor, report); err != nil {
		return err
	}
	if includeLogs {
		if err := migrateLogsTx(tx, quotaFactor, priceFactor, report); err != nil {
			return err
		}
	}
	return nil
}

func migrateLogLedger(db *gorm.DB, quotaFactor, priceFactor, oldQuotaPerUnit decimal.Decimal, scope string, report *BillingCurrencyMigrationReport) error {
	if err := db.AutoMigrate(&BillingLedgerVersion{}); err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := migrateLogsTx(tx, quotaFactor, priceFactor, report); err != nil {
			return err
		}
		return writeBillingLedgerMarker(tx, scope, priceFactor.String(), oldQuotaPerUnit.String())
	})
}

func validateLogLedgerMigration(db *gorm.DB, quotaFactor, priceFactor decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := migrateLogsTx(tx, quotaFactor, priceFactor, report); err != nil {
			return err
		}
		return errBillingMigrationDryRunRollback
	})
	if errors.Is(err, errBillingMigrationDryRunRollback) {
		return nil
	}
	return err
}

func migrateUsers(tx *gorm.DB, rate decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if !tx.Migrator().HasTable(&User{}) {
		return nil
	}
	var rows []User
	if err := tx.Unscoped().Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		quota, err := scaleWalletQuota(rows[i].Quota, rate)
		if err != nil {
			return fmt.Errorf("users[%d].quota: %w", rows[i].Id, err)
		}
		used, err := scaleQuota(rows[i].UsedQuota, rate)
		if err != nil {
			return fmt.Errorf("users[%d].used_quota: %w", rows[i].Id, err)
		}
		aff, err := scaleWalletQuota(rows[i].AffQuota, rate)
		if err != nil {
			return fmt.Errorf("users[%d].aff_quota: %w", rows[i].Id, err)
		}
		history, err := scaleQuota(rows[i].AffHistoryQuota, rate)
		if err != nil {
			return fmt.Errorf("users[%d].aff_history: %w", rows[i].Id, err)
		}
		if err := tx.Unscoped().Model(&User{}).Where("id = ?", rows[i].Id).Updates(map[string]any{
			"quota": quota, "used_quota": used, "aff_quota": aff, "aff_history": history,
		}).Error; err != nil {
			return err
		}
	}
	report.UpdatedRows["users"] += int64(len(rows))
	return nil
}

func migrateTokens(tx *gorm.DB, rate decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if !tx.Migrator().HasTable(&Token{}) {
		return nil
	}
	var rows []Token
	if err := tx.Unscoped().Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		remain, err := scaleWalletQuota(rows[i].RemainQuota, rate)
		if err != nil {
			return fmt.Errorf("tokens[%d].remain_quota: %w", rows[i].Id, err)
		}
		used, err := scaleQuota(rows[i].UsedQuota, rate)
		if err != nil {
			return fmt.Errorf("tokens[%d].used_quota: %w", rows[i].Id, err)
		}
		if err := tx.Unscoped().Model(&Token{}).Where("id = ?", rows[i].Id).Updates(map[string]any{"remain_quota": remain, "used_quota": used}).Error; err != nil {
			return err
		}
	}
	report.UpdatedRows["tokens"] += int64(len(rows))
	return nil
}

func migrateChannels(tx *gorm.DB, rate decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if !tx.Migrator().HasTable(&Channel{}) {
		return nil
	}
	var rows []Channel
	if err := tx.Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		used, err := scaleQuota64(rows[i].UsedQuota, rate)
		if err != nil {
			return fmt.Errorf("channels[%d].used_quota: %w", rows[i].Id, err)
		}
		if err := tx.Model(&Channel{}).Where("id = ?", rows[i].Id).Update("used_quota", used).Error; err != nil {
			return err
		}
	}
	report.UpdatedRows["channels"] += int64(len(rows))
	return nil
}

func migrateTeams(tx *gorm.DB, rate decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if err := migrateOneIntColumn[Team](tx, "teams", "id", "quota", rate, report, func(row Team) (int, int) { return row.Id, row.Quota }); err != nil {
		return err
	}
	if err := migrateOneIntColumn[TeamMember](tx, "team_members", "id", "monthly_limit_quota", rate, report, func(row TeamMember) (int, int) { return row.Id, row.MonthlyLimitQuota }); err != nil {
		return err
	}
	if err := migrateOneIntColumn[TeamMonthlyUsage](tx, "team_monthly_usages", "id", "quota", rate, report, func(row TeamMonthlyUsage) (int, int) { return row.Id, row.Quota }); err != nil {
		return err
	}
	if err := migrateOneIntColumn[TeamQuotaReservation](tx, "team_quota_reservations", "id", "quota", rate, report, func(row TeamQuotaReservation) (int, int) { return row.Id, row.Quota }); err != nil {
		return err
	}
	return migrateOneIntColumn[TeamQuotaTransfer](tx, "team_quota_transfers", "id", "quota", rate, report, func(row TeamQuotaTransfer) (int, int) { return row.Id, row.Quota })
}

func migrateOneIntColumn[T any](tx *gorm.DB, table, idColumn, valueColumn string, rate decimal.Decimal, report *BillingCurrencyMigrationReport, values func(T) (int, int)) error {
	if !tx.Migrator().HasTable(table) {
		return nil
	}
	var rows []T
	if err := tx.Table(table).Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		id, value := values(row)
		scaler := scaleQuota
		if table == "teams" || table == "team_members" {
			scaler = scaleWalletQuota
		}
		scaled, err := scaler(value, rate)
		if err != nil {
			return fmt.Errorf("%s[%d].%s: %w", table, id, valueColumn, err)
		}
		if err := tx.Table(table).Where(idColumn+" = ?", id).Update(valueColumn, scaled).Error; err != nil {
			return err
		}
	}
	report.UpdatedRows[table] += int64(len(rows))
	return nil
}

func migrateRedemptionsAndAwards(tx *gorm.DB, rate decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if tx.Migrator().HasTable(&Redemption{}) {
		var rows []Redemption
		if err := tx.Unscoped().Find(&rows).Error; err != nil {
			return err
		}
		for i := range rows {
			scaled, err := scaleWalletQuota(rows[i].Quota, rate)
			if err != nil {
				return fmt.Errorf("redemptions[%d].quota: %w", rows[i].Id, err)
			}
			if err := tx.Unscoped().Model(&Redemption{}).Where("id = ?", rows[i].Id).Update("quota", scaled).Error; err != nil {
				return err
			}
		}
		report.UpdatedRows["redemptions"] += int64(len(rows))
	}
	if err := migrateOneIntColumn[Checkin](tx, "checkins", "id", "quota_awarded", rate, report, func(row Checkin) (int, int) { return row.Id, row.QuotaAwarded }); err != nil {
		return err
	}
	return migrateOneIntColumn[Midjourney](tx, "midjourneys", "id", "quota", rate, report, func(row Midjourney) (int, int) { return row.Id, row.Quota })
}

func migrateUsageAndTasks(tx *gorm.DB, quotaFactor, priceFactor decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if err := migrateOneIntColumn[QuotaData](tx, "quota_data", "id", "quota", quotaFactor, report, func(row QuotaData) (int, int) { return row.Id, row.Quota }); err != nil {
		return err
	}
	if !tx.Migrator().HasTable(&Task{}) {
		return nil
	}
	var tasks []Task
	if err := tx.Find(&tasks).Error; err != nil {
		return err
	}
	for i := range tasks {
		quota, err := scaleQuota(tasks[i].Quota, quotaFactor)
		if err != nil {
			return fmt.Errorf("tasks[%d].quota: %w", tasks[i].ID, err)
		}
		if err := scaleTaskPrivateData(&tasks[i].PrivateData, quotaFactor, priceFactor); err != nil {
			return fmt.Errorf("tasks[%d].private_data: %w", tasks[i].ID, err)
		}
		if err := tx.Model(&Task{}).Where("id = ?", tasks[i].ID).Updates(map[string]any{"quota": quota, "private_data": tasks[i].PrivateData}).Error; err != nil {
			return err
		}
	}
	report.UpdatedRows["tasks"] += int64(len(tasks))
	return nil
}

func scaleTaskPrivateData(private *TaskPrivateData, quotaFactor, priceFactor decimal.Decimal) error {
	if private == nil || private.BillingContext == nil {
		return nil
	}
	context := private.BillingContext
	var err error
	context.ModelPrice, err = scaleMoney(context.ModelPrice, priceFactor)
	if err != nil {
		return err
	}
	if context.TieredSnapshot == nil {
		context.ModelRatio, err = scaleMoney(context.ModelRatio, quotaFactor)
		return err
	}
	context.ModelRatio, err = scaleMoney(context.ModelRatio, quotaFactor)
	if err != nil {
		return err
	}
	snapshot := context.TieredSnapshot
	snapshot.ExprString, err = scaleExpression(snapshot.ExprString, priceFactor)
	if err != nil {
		return err
	}
	snapshot.ExprHash = billingexpr.ExprHashString(snapshot.ExprString)
	snapshot.QuotaPerUnit = 500000
	snapshot.EstimatedQuotaBeforeGroup, err = scaleMoney(snapshot.EstimatedQuotaBeforeGroup, quotaFactor)
	if err != nil {
		return err
	}
	snapshot.EstimatedQuotaAfterGroup, err = scaleQuota(snapshot.EstimatedQuotaAfterGroup, quotaFactor)
	if err != nil {
		return err
	}
	if snapshot.EstimatedFixedPrice != nil {
		value, scaleErr := scaleMoney(*snapshot.EstimatedFixedPrice, priceFactor)
		if scaleErr != nil {
			return scaleErr
		}
		snapshot.EstimatedFixedPrice = &value
	}
	return nil
}

func migrateSubscriptions(tx *gorm.DB, quotaFactor, priceFactor decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if tx.Migrator().HasTable(&SubscriptionPlan{}) {
		var plans []SubscriptionPlan
		if err := tx.Find(&plans).Error; err != nil {
			return err
		}
		for i := range plans {
			price, err := scaleMoney(plans[i].PriceAmount, priceFactor)
			if err != nil {
				return fmt.Errorf("subscription_plans[%d].price_amount: %w", plans[i].Id, err)
			}
			total, err := scaleQuota64(plans[i].TotalAmount, quotaFactor)
			if err != nil {
				return fmt.Errorf("subscription_plans[%d].total_amount: %w", plans[i].Id, err)
			}
			if err := tx.Model(&SubscriptionPlan{}).Where("id = ?", plans[i].Id).Updates(map[string]any{"price_amount": price, "currency": billingLedgerCurrency, "total_amount": total}).Error; err != nil {
				return err
			}
		}
		report.UpdatedRows["subscription_plans"] += int64(len(plans))
	}
	if tx.Migrator().HasTable(&SubscriptionOrder{}) {
		hasCurrencyColumn := tx.Migrator().HasColumn(&SubscriptionOrder{}, "Currency")
		var orders []SubscriptionOrder
		if err := tx.Find(&orders).Error; err != nil {
			return err
		}
		for i := range orders {
			currency, scale, err := legacySubscriptionOrderCurrency(orders[i].PaymentProvider)
			if err != nil {
				return fmt.Errorf("subscription_orders[%d]: %w", orders[i].Id, err)
			}
			money := orders[i].Money
			if scale {
				money, err = scaleMoney(money, priceFactor)
				if err != nil {
					return fmt.Errorf("subscription_orders[%d].money: %w", orders[i].Id, err)
				}
			}
			updates := map[string]any{"money": money}
			if hasCurrencyColumn {
				updates["currency"] = currency
			}
			if err := tx.Model(&SubscriptionOrder{}).Where("id = ?", orders[i].Id).Updates(updates).Error; err != nil {
				return err
			}
		}
		report.UpdatedRows["subscription_orders"] += int64(len(orders))
	}
	if tx.Migrator().HasTable(&UserSubscription{}) {
		var rows []UserSubscription
		if err := tx.Find(&rows).Error; err != nil {
			return err
		}
		for i := range rows {
			total, err := scaleQuota64(rows[i].AmountTotal, quotaFactor)
			if err != nil {
				return err
			}
			used, err := scaleQuota64(rows[i].AmountUsed, quotaFactor)
			if err != nil {
				return err
			}
			if err := tx.Model(&UserSubscription{}).Where("id = ?", rows[i].Id).Updates(map[string]any{"amount_total": total, "amount_used": used}).Error; err != nil {
				return err
			}
		}
		report.UpdatedRows["user_subscriptions"] += int64(len(rows))
	}
	if tx.Migrator().HasTable(&SubscriptionPreConsumeRecord{}) {
		var rows []SubscriptionPreConsumeRecord
		if err := tx.Find(&rows).Error; err != nil {
			return err
		}
		for i := range rows {
			value, err := scaleQuota64(rows[i].PreConsumed, quotaFactor)
			if err != nil {
				return err
			}
			if err := tx.Model(&SubscriptionPreConsumeRecord{}).Where("id = ?", rows[i].Id).Update("pre_consumed", value).Error; err != nil {
				return err
			}
		}
		report.UpdatedRows["subscription_pre_consume_records"] += int64(len(rows))
	}
	return nil
}

func legacySubscriptionOrderCurrency(provider string) (currency string, scale bool, err error) {
	switch strings.TrimSpace(provider) {
	case PaymentProviderEpay, PaymentProviderBalance:
		return billingLedgerCurrency, provider == PaymentProviderBalance, nil
	case PaymentProviderStripe:
		return "USD", false, nil
	case PaymentProviderCreem, PaymentProviderWaffo, PaymentProviderWaffoPancake, "":
		return "", false, fmt.Errorf("payment provider %q has no safely inferable historical currency", provider)
	default:
		return "", false, fmt.Errorf("unknown payment provider %q", provider)
	}
}

func migrateTopUps(tx *gorm.DB, quotaFactor, oldQuotaPerUnit decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if !tx.Migrator().HasTable(&TopUp{}) {
		return nil
	}
	subscriptionCurrencies := make(map[string]string)
	if tx.Migrator().HasTable(&SubscriptionOrder{}) {
		var orders []SubscriptionOrder
		if err := tx.Find(&orders).Error; err != nil {
			return err
		}
		for _, order := range orders {
			currency, _, err := legacySubscriptionOrderCurrency(order.PaymentProvider)
			if err != nil {
				return fmt.Errorf("subscription_orders[%d]: %w", order.Id, err)
			}
			subscriptionCurrencies[order.TradeNo] = currency
		}
	}
	hasCurrencyColumn := tx.Migrator().HasColumn(&TopUp{}, "Currency")
	hasCreditedQuotaColumn := tx.Migrator().HasColumn(&TopUp{}, "CreditedQuota")
	var rows []TopUp
	if err := tx.Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		currency := ""
		creditedQuota := int64(0)
		if subscriptionCurrency, ok := subscriptionCurrencies[rows[i].TradeNo]; ok {
			currency = subscriptionCurrency
		} else {
			var legacyQuota decimal.Decimal
			switch strings.TrimSpace(rows[i].PaymentProvider) {
			case PaymentProviderEpay:
				currency = billingLedgerCurrency
				legacyQuota = decimal.NewFromInt(rows[i].Amount).Mul(oldQuotaPerUnit)
			case PaymentProviderStripe:
				currency = "USD"
				legacyQuota = decimal.NewFromFloat(rows[i].Money).Mul(oldQuotaPerUnit)
			default:
				return fmt.Errorf("top_ups[%d]: payment provider %q has no safely inferable historical currency and credited quota", rows[i].Id, rows[i].PaymentProvider)
			}
			if rows[i].Status == common.TopUpStatusSuccess {
				var err error
				creditedQuota, err = scaleCreditedQuota(legacyQuota, quotaFactor)
				if err != nil {
					return fmt.Errorf("top_ups[%d].credited_quota: %w", rows[i].Id, err)
				}
			}
		}
		updates := make(map[string]any, 2)
		if hasCurrencyColumn {
			updates["currency"] = currency
		}
		if hasCreditedQuotaColumn {
			updates["credited_quota"] = creditedQuota
		}
		if len(updates) > 0 {
			if err := tx.Model(&TopUp{}).Where("id = ?", rows[i].Id).Updates(updates).Error; err != nil {
				return err
			}
		}
	}
	report.UpdatedRows["top_ups"] += int64(len(rows))
	return nil
}

func scaleCreditedQuota(legacyQuota, quotaFactor decimal.Decimal) (int64, error) {
	scaled := legacyQuota.Mul(quotaFactor).Round(0)
	limit := decimal.NewFromInt(int64(common.MaxWalletQuota))
	if scaled.IsNegative() || scaled.GreaterThan(limit) {
		return 0, fmt.Errorf("scaled credited quota must be between 0 and %d", common.MaxWalletQuota)
	}
	return scaled.IntPart(), nil
}

func migrateBillingOptions(tx *gorm.DB, quotaFactor, priceFactor decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if !tx.Migrator().HasTable(&Option{}) {
		return nil
	}
	var rows []Option
	if err := tx.Find(&rows).Error; err != nil {
		return err
	}
	integerKeys := map[string]bool{
		"QuotaForNewUser": true, "QuotaForInviter": true, "QuotaForInvitee": true,
		"QuotaRemindThreshold": true, "PreConsumedQuota": true,
	}
	priceMapKeys := map[string]bool{"ModelPrice": true, "tool_price_setting.prices": true}
	expressionMapKeys := map[string]bool{"billing_setting.billing_expr": true, "billing_setting.plugin_billing_expr": true}
	for _, row := range rows {
		value := row.Value
		var err error
		switch {
		case integerKeys[row.Key]:
			parsed, parseErr := strconv.Atoi(strings.TrimSpace(value))
			if parseErr != nil {
				return fmt.Errorf("option %s: %w", row.Key, parseErr)
			}
			parsed, err = scaleQuota(parsed, quotaFactor)
			value = strconv.Itoa(parsed)
		case priceMapKeys[row.Key]:
			value, err = scaleJSONNumberMap(value, priceFactor)
		case row.Key == "ModelRatio":
			value, err = scaleJSONNumberMap(value, quotaFactor)
		case expressionMapKeys[row.Key]:
			value, err = scaleJSONExpressionMap(value, priceFactor)
		case row.Key == "CreemProducts":
			value, err = scaleCreemProducts(value, quotaFactor)
		case row.Key == "checkin_setting.min_quota" || row.Key == "checkin_setting.max_quota":
			parsed, parseErr := strconv.Atoi(strings.TrimSpace(value))
			if parseErr != nil {
				return fmt.Errorf("option %s: %w", row.Key, parseErr)
			}
			parsed, err = scaleQuota(parsed, quotaFactor)
			value = strconv.Itoa(parsed)
		}
		if err != nil {
			return fmt.Errorf("option %s: %w", row.Key, err)
		}
		if value != row.Value {
			if err := tx.Model(&Option{}).Where(commonKeyCol+" = ?", row.Key).Update("value", value).Error; err != nil {
				return err
			}
			report.UpdatedRows["options"]++
		}
	}
	return freezeLegacyPricingDefaults(tx, quotaFactor, priceFactor, report)
}

func freezeLegacyPricingDefaults(tx *gorm.DB, quotaFactor, priceFactor decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	historicalDefault := decimal.NewFromFloat(legacyDefaultUSDExchangeRate)
	builtins := billing_setting.GetBuiltinBillingExprCopy()
	defaults := ratio_setting.GetDefaultPricingMaps()
	for _, config := range []struct {
		key    string
		factor decimal.Decimal
	}{
		{key: "ModelRatio", factor: quotaFactor},
		{key: "ModelPrice", factor: priceFactor},
	} {
		if config.factor.Equal(historicalDefault) {
			continue
		}
		values, err := readFloatOptionMap(tx, config.key)
		if err != nil {
			return err
		}
		for modelName, currentDefault := range defaults[config.key] {
			if _, isBuiltin := builtins[modelName]; isBuiltin {
				continue
			}
			if _, explicitlyConfigured := values[modelName]; explicitlyConfigured {
				continue
			}
			scaled := decimal.NewFromFloat(currentDefault).Mul(config.factor).Div(historicalDefault)
			values[modelName], _ = scaled.Float64()
		}
		if err := writeOptionJSON(tx, config.key, values); err != nil {
			return err
		}
		report.UpdatedRows["options"]++
	}
	if priceFactor.Equal(historicalDefault) {
		return nil
	}
	expressions, err := readStringOptionMap(tx, "billing_setting.billing_expr")
	if err != nil {
		return err
	}
	modes, err := readStringOptionMap(tx, "billing_setting.billing_mode")
	if err != nil {
		return err
	}
	configuredRatios, err := readFloatOptionMap(tx, "ModelRatio")
	if err != nil {
		return err
	}
	configuredPrices, err := readFloatOptionMap(tx, "ModelPrice")
	if err != nil {
		return err
	}
	factor := priceFactor.Div(historicalDefault)
	for modelName, currentExpression := range builtins {
		if _, configured := modes[modelName]; configured {
			continue
		}
		if _, configured := configuredRatios[modelName]; configured {
			continue
		}
		if _, configured := configuredPrices[modelName]; configured {
			continue
		}
		if _, configured := expressions[modelName]; configured {
			continue
		}
		scaled, scaleErr := scaleExpression(currentExpression, factor)
		if scaleErr != nil {
			return fmt.Errorf("freeze builtin %s: %w", modelName, scaleErr)
		}
		expressions[modelName] = scaled
		modes[modelName] = billing_setting.BillingModeTieredExpr
	}
	if err := writeOptionJSON(tx, "billing_setting.billing_expr", expressions); err != nil {
		return err
	}
	if err := writeOptionJSON(tx, "billing_setting.billing_mode", modes); err != nil {
		return err
	}
	report.UpdatedRows["options"] += 2
	return nil
}

func readFloatOptionMap(tx *gorm.DB, key string) (map[string]float64, error) {
	var row Option
	err := tx.Where(commonKeyCol+" = ?", key).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make(map[string]float64), nil
	}
	if err != nil {
		return nil, err
	}
	values := make(map[string]float64)
	if err := common.UnmarshalJsonStr(row.Value, &values); err != nil {
		return nil, fmt.Errorf("option %s: %w", key, err)
	}
	return values, nil
}

func readStringOptionMap(tx *gorm.DB, key string) (map[string]string, error) {
	var row Option
	err := tx.Where(commonKeyCol+" = ?", key).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return make(map[string]string), nil
	}
	if err != nil {
		return nil, err
	}
	values := make(map[string]string)
	if err := common.UnmarshalJsonStr(row.Value, &values); err != nil {
		return nil, fmt.Errorf("option %s: %w", key, err)
	}
	return values, nil
}

func writeOptionJSON(tx *gorm.DB, key string, value any) error {
	encoded, err := common.Marshal(value)
	if err != nil {
		return err
	}
	return tx.Save(&Option{Key: key, Value: string(encoded)}).Error
}

func scaleJSONNumberMap(value string, rate decimal.Decimal) (string, error) {
	var values map[string]float64
	if err := common.UnmarshalJsonStr(value, &values); err != nil {
		return "", err
	}
	for key, current := range values {
		scaled, err := scaleMoney(current, rate)
		if err != nil {
			return "", fmt.Errorf("%s: %w", key, err)
		}
		values[key] = scaled
	}
	encoded, err := common.Marshal(values)
	return string(encoded), err
}

func scaleJSONExpressionMap(value string, rate decimal.Decimal) (string, error) {
	var values map[string]string
	if err := common.UnmarshalJsonStr(value, &values); err != nil {
		return "", err
	}
	for key, expression := range values {
		scaled, err := scaleExpression(expression, rate)
		if err != nil {
			return "", fmt.Errorf("%s: %w", key, err)
		}
		values[key] = scaled
	}
	encoded, err := common.Marshal(values)
	return string(encoded), err
}

func scaleCreemProducts(value string, quotaFactor decimal.Decimal) (string, error) {
	var products []map[string]any
	if err := common.UnmarshalJsonStr(value, &products); err != nil {
		return "", err
	}
	for i := range products {
		raw, ok := products[i]["quota"]
		if !ok {
			continue
		}
		number, ok := raw.(float64)
		if !ok || math.Trunc(number) != number {
			return "", fmt.Errorf("product %d quota is not an integer", i)
		}
		scaled, err := scaleQuota64(int64(number), quotaFactor)
		if err != nil {
			return "", fmt.Errorf("product %d quota: %w", i, err)
		}
		products[i]["quota"] = scaled
	}
	encoded, err := common.Marshal(products)
	return string(encoded), err
}

func migrateLogsTx(tx *gorm.DB, quotaFactor, priceFactor decimal.Decimal, report *BillingCurrencyMigrationReport) error {
	if !tx.Migrator().HasTable(&Log{}) {
		return nil
	}
	var logs []Log
	if err := tx.Find(&logs).Error; err != nil {
		return err
	}
	for i := range logs {
		quota, err := scaleQuota(logs[i].Quota, quotaFactor)
		if err != nil {
			return fmt.Errorf("logs[%d].quota: %w", logs[i].Id, err)
		}
		other, err := scaleLogOther(logs[i].Other, quotaFactor, priceFactor)
		if err != nil {
			return fmt.Errorf("logs[%d].other: %w", logs[i].Id, err)
		}
		if err := tx.Model(&Log{}).Where("id = ?", logs[i].Id).Updates(map[string]any{"quota": quota, "other": other}).Error; err != nil {
			return err
		}
	}
	report.UpdatedRows["logs"] += int64(len(logs))
	return nil
}

func scaleLogOther(value string, quotaFactor, priceFactor decimal.Decimal) (string, error) {
	if strings.TrimSpace(value) == "" {
		return value, nil
	}
	var root map[string]any
	if err := common.UnmarshalJsonStr(value, &root); err != nil {
		return "", err
	}
	if err := scaleLogMap(root, quotaFactor, priceFactor); err != nil {
		return "", err
	}
	encoded, err := common.Marshal(root)
	return string(encoded), err
}

func scaleLogMap(values map[string]any, quotaFactor, priceFactor decimal.Decimal) error {
	priceKeys := map[string]bool{
		"model_price": true, "fixed_price": true, "audio_input_price": true,
		"audio_input_seperate_price": true, "price": true,
	}
	ratioKeys := map[string]bool{"model_ratio": true}
	quotaKeys := map[string]bool{
		"pre_consumed_quota": true, "actual_quota": true, "wallet_quota_deducted": true,
		"tool_call_surcharge_quota": true, "quota": true,
	}
	for key, raw := range values {
		switch nested := raw.(type) {
		case map[string]any:
			if err := scaleLogMap(nested, quotaFactor, priceFactor); err != nil {
				return err
			}
			continue
		case []any:
			if err := scaleLogArray(nested, quotaFactor, priceFactor); err != nil {
				return err
			}
			continue
		}
		if encoded, ok := raw.(string); ok && key == "expr_b64" {
			expression, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return fmt.Errorf("expr_b64: %w", err)
			}
			scaled, err := scaleExpression(string(expression), priceFactor)
			if err != nil {
				return fmt.Errorf("expr_b64: %w", err)
			}
			values[key] = base64.StdEncoding.EncodeToString([]byte(scaled))
			continue
		}
		if expression, ok := raw.(string); ok && (key == "billing_expr" || key == "expr_string") {
			scaled, err := scaleExpression(expression, priceFactor)
			if err != nil {
				return err
			}
			values[key] = scaled
			continue
		}
		number, ok := raw.(float64)
		if !ok {
			continue
		}
		if priceKeys[key] {
			scaled, err := scaleMoney(number, priceFactor)
			if err != nil {
				return err
			}
			values[key] = scaled
		} else if ratioKeys[key] {
			scaled, err := scaleMoney(number, quotaFactor)
			if err != nil {
				return err
			}
			values[key] = scaled
		} else if quotaKeys[key] {
			if math.Trunc(number) != number {
				return fmt.Errorf("%s is not an integer quota", key)
			}
			scaled, err := scaleQuota64(int64(number), quotaFactor)
			if err != nil {
				return err
			}
			values[key] = scaled
		}
	}
	return nil
}

func scaleLogArray(values []any, quotaFactor, priceFactor decimal.Decimal) error {
	for _, raw := range values {
		switch nested := raw.(type) {
		case map[string]any:
			if err := scaleLogMap(nested, quotaFactor, priceFactor); err != nil {
				return err
			}
		case []any:
			if err := scaleLogArray(nested, quotaFactor, priceFactor); err != nil {
				return err
			}
		}
	}
	return nil
}
