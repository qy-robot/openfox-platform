package operation_setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBillingPresentationIsAlwaysCNY(t *testing.T) {
	setting := GetGeneralSetting()
	original := *setting
	t.Cleanup(func() { *setting = original })

	setting.QuotaDisplayType = QuotaDisplayTypeUSD
	setting.CustomCurrencySymbol = "$"
	setting.CustomCurrencyExchangeRate = 7.3

	require.Equal(t, QuotaDisplayTypeCNY, GetQuotaDisplayType())
	require.Equal(t, "¥", GetCurrencySymbol())
	require.Equal(t, float64(1), GetUsdToCurrencyRate(7.3))
	require.True(t, IsCurrencyDisplay())
	require.True(t, IsCNYDisplay())
}

func TestValidateBillingCurrencyOptionAllowsOnlyFixedCompatibilityValues(t *testing.T) {
	valid := map[string]string{
		"Price":                                         "1",
		"USDExchangeRate":                               "1.0",
		"QuotaPerUnit":                                  "500000",
		"DisplayInCurrencyEnabled":                      "true",
		"general_setting.quota_display_type":            "CNY",
		"general_setting.custom_currency_symbol":        "¥",
		"general_setting.custom_currency_exchange_rate": "1",
	}
	for key, value := range valid {
		require.NoError(t, ValidateBillingCurrencyOption(key, value), key)
	}

	invalid := map[string]string{
		"Price":                                         "7.3",
		"USDExchangeRate":                               "7.3",
		"QuotaPerUnit":                                  "1000000",
		"DisplayInCurrencyEnabled":                      "false",
		"general_setting.quota_display_type":            "USD",
		"general_setting.custom_currency_symbol":        "$",
		"general_setting.custom_currency_exchange_rate": "6.9",
	}
	for key, value := range invalid {
		require.Error(t, ValidateBillingCurrencyOption(key, value), key)
	}
}
