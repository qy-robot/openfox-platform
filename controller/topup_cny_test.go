package controller

import (
	"maps"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func TestGetPayMoneyUsesCNYAmountDirectly(t *testing.T) {
	originalDiscounts := make(map[int]float64, len(operation_setting.GetPaymentSetting().AmountDiscount))
	maps.Copy(originalDiscounts, operation_setting.GetPaymentSetting().AmountDiscount)
	originalRatios := common.TopupGroupRatio2JSONString()
	t.Cleanup(func() {
		operation_setting.GetPaymentSetting().AmountDiscount = originalDiscounts
		require.NoError(t, common.UpdateTopupGroupRatioByJSONString(originalRatios))
	})

	operation_setting.GetPaymentSetting().AmountDiscount = map[int]float64{100: 0.8}
	require.NoError(t, common.UpdateTopupGroupRatioByJSONString(`{"default":1,"vip":1.25}`))

	require.Equal(t, float64(80), getPayMoney(100, "default"))
	require.Equal(t, float64(100), getPayMoney(100, "vip"), "group ratio and discount remain intentional")
	require.Equal(t, float64(50), getPayMoney(50, "default"))
}

func TestSubscriptionEpayRequiresMatchingCNYPayment(t *testing.T) {
	db := modelManagementDB(t, "sqlite", "")
	order := model.SubscriptionOrder{TradeNo: "cny-subscription-payment", Money: 30,
		Currency: "CNY", PaymentProvider: model.PaymentProviderEpay}
	require.NoError(t, db.Create(&order).Error)
	require.NoError(t, validateSubscriptionEpayPayment(order.TradeNo, "30.000"))
	require.Error(t, validateSubscriptionEpayPayment(order.TradeNo, "29.99"))
	require.Error(t, validateSubscriptionEpayPayment(order.TradeNo, "29.999"))
	require.Error(t, validateSubscriptionEpayPayment("missing", "30.00"))
	require.NoError(t, db.Model(&order).Update("currency", "USD").Error)
	require.Error(t, validateSubscriptionEpayPayment(order.TradeNo, "30.00"))
}

func TestEpayCallbackAmountMustMatchCNYOrder(t *testing.T) {
	require.NoError(t, validateEpayPaidAmount(30, "30.00"))
	require.NoError(t, validateEpayPaidAmount(30.1, "30.10"))
	require.Error(t, validateEpayPaidAmount(30, "29.99"))
	require.Error(t, validateEpayPaidAmount(30, "29.999"))
	require.NoError(t, validateEpayPaidAmount(30, "30.000"))
	require.Error(t, validateEpayPaidAmount(30, "USD 30"))
}
