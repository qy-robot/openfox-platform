package operation_setting

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

// 额度展示类型
const (
	QuotaDisplayTypeUSD    = "USD"
	QuotaDisplayTypeCNY    = "CNY"
	QuotaDisplayTypeTokens = "TOKENS"
	QuotaDisplayTypeCustom = "CUSTOM"
)

type GeneralSetting struct {
	DocsLink            string `json:"docs_link"`
	PingIntervalEnabled bool   `json:"ping_interval_enabled"`
	PingIntervalSeconds int    `json:"ping_interval_seconds"`
	// Legacy configuration fields are retained for compatibility; runtime values
	// are fixed to CNY, ¥ and 1 and cannot enable currency conversion.
	QuotaDisplayType           string  `json:"quota_display_type"`
	CustomCurrencySymbol       string  `json:"custom_currency_symbol"`
	CustomCurrencyExchangeRate float64 `json:"custom_currency_exchange_rate"`
}

// 默认配置
var generalSetting = GeneralSetting{
	DocsLink:                   "https://docs.newapi.pro",
	PingIntervalEnabled:        false,
	PingIntervalSeconds:        60,
	QuotaDisplayType:           QuotaDisplayTypeCNY,
	CustomCurrencySymbol:       "¥",
	CustomCurrencyExchangeRate: 1.0,
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("general_setting", &generalSetting)
}

func GetGeneralSetting() *GeneralSetting {
	return &generalSetting
}

// IsCurrencyDisplay reports the fixed monetary (CNY) display mode.
func IsCurrencyDisplay() bool {
	return true
}

// IsCNYDisplay 是否以人民币展示
func IsCNYDisplay() bool {
	return true
}

// GetQuotaDisplayType 返回额度展示类型
func GetQuotaDisplayType() string {
	return QuotaDisplayTypeCNY
}

// GetCurrencySymbol 返回当前展示类型对应符号
func GetCurrencySymbol() string {
	return "¥"
}

// GetUsdToCurrencyRate remains for source compatibility. RMB is the native
// ledger currency, so no runtime conversion is applied.
func GetUsdToCurrencyRate(_ float64) float64 {
	return 1
}

func fixedBillingCurrencyOptionValue(key string) (string, bool) {
	switch key {
	case "Price", "USDExchangeRate", "general_setting.custom_currency_exchange_rate":
		return "1", true
	case "QuotaPerUnit":
		return "500000", true
	case "DisplayInCurrencyEnabled":
		return "true", true
	case "general_setting.quota_display_type":
		return QuotaDisplayTypeCNY, true
	case "general_setting.custom_currency_symbol":
		return "¥", true
	default:
		return "", false
	}
}

func CanonicalBillingCurrencyOption(key string) (string, bool) {
	return fixedBillingCurrencyOptionValue(key)
}

func ValidateBillingCurrencyOption(key, value string) error {
	fixed, ok := fixedBillingCurrencyOptionValue(key)
	if !ok {
		return nil
	}
	if key == "Price" || key == "USDExchangeRate" || key == "general_setting.custom_currency_exchange_rate" || key == "QuotaPerUnit" {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		expected, _ := strconv.ParseFloat(fixed, 64)
		if err == nil && parsed == expected {
			return nil
		}
	} else if strings.TrimSpace(value) == fixed {
		return nil
	}
	return fmt.Errorf("%s is fixed to %s because billing currency is CNY", key, fixed)
}
