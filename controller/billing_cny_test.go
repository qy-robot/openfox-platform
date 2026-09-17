package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupBillingCNYTest(t *testing.T) *model.Token {
	t.Helper()

	previousDB := model.DB
	previousDisplayTokenStat := common.DisplayTokenStatEnabled
	previousQuotaPerUnit := common.QuotaPerUnit
	previousRedisEnabled := common.RedisEnabled

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Token{}))
	model.DB = db
	common.DisplayTokenStatEnabled = true
	common.QuotaPerUnit = 500000
	common.RedisEnabled = false

	token := &model.Token{Id: 901, UserId: 902, Key: "billing-cny", RemainQuota: 1_000_000, UsedQuota: 250_000}
	require.NoError(t, db.Create(token).Error)

	t.Cleanup(func() {
		model.DB = previousDB
		common.DisplayTokenStatEnabled = previousDisplayTokenStat
		common.QuotaPerUnit = previousQuotaPerUnit
		common.RedisEnabled = previousRedisEnabled
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	return token
}

func billingCNYRequest(t *testing.T, tokenID int, handler gin.HandlerFunc) map[string]any {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set("token_id", tokenID)
	handler(ctx)
	require.Equal(t, http.StatusOK, recorder.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	return body
}

func TestCompatibleBillingResponsesDeclareCNY(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupBillingCNYTest(t)

	subscription := billingCNYRequest(t, token.Id, GetSubscription)
	require.Equal(t, operation_setting.BillingCurrency, subscription["currency"])
	require.Equal(t, float64(2.5), subscription["hard_limit_usd"])

	usage := billingCNYRequest(t, token.Id, GetUsage)
	require.Equal(t, operation_setting.BillingCurrency, usage["currency"])
	require.Equal(t, float64(50), usage["total_usage"])
}

func TestCompatibleBillingCurrencyIsOptionalWhenParsingUpstream(t *testing.T) {
	var subscription OpenAISubscriptionResponse
	require.NoError(t, json.Unmarshal([]byte(`{"object":"billing_subscription","hard_limit_usd":12.5}`), &subscription))
	require.Empty(t, subscription.Currency)

	var usage OpenAIUsageResponse
	require.NoError(t, json.Unmarshal([]byte(`{"object":"list","total_usage":125}`), &usage))
	require.Empty(t, usage.Currency)
}
