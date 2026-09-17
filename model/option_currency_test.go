package model

import (
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCNYLedgerIgnoresLegacyQuotaCaches(t *testing.T) {
	useUserCacheMiniRedis(t)
	user := User{Id: 4209, Username: "legacy-ledger", Quota: 100, AuthVersion: 1}
	require.NoError(t, populateUserCache(user))
	userFields, err := common.RDB.HGetAll(t.Context(), getUserCacheKey(user.Id)).Result()
	require.NoError(t, err)
	require.NoError(t, common.RDB.Del(t.Context(), getUserCacheKey(user.Id)).Err())
	require.NoError(t, common.RDB.HSet(t.Context(), "user:4209", userFields).Err())
	_, err = cacheGetUserBase(user.Id)
	require.Error(t, err, "legacy USD quota cache must not be interpreted as CNY")

	token := Token{Id: 4209, Key: "legacy-currency-cache", RemainQuota: 100}
	require.NoError(t, cacheSetTokenForTest(token))
	tokenFields, err := common.RDB.HGetAll(t.Context(), getTokenCacheKey(token.Key)).Result()
	require.NoError(t, err)
	require.NoError(t, common.RDB.Del(t.Context(), getTokenCacheKey(token.Key)).Err())
	require.NoError(t, common.RDB.HSet(t.Context(), "token:"+common.GenerateHMAC(token.Key), tokenFields).Err())
	_, err = cacheGetTokenByKey(token.Key)
	require.Error(t, err, "legacy USD token quota cache must not be interpreted as CNY")
}

func TestCNYLedgerIgnoresLegacySubscriptionCaches(t *testing.T) {
	truncateTables(t)
	server := useUserCacheMiniRedis(t)

	resetSubscriptionCaches := func() {
		subscriptionPlanCacheOnce = sync.Once{}
		subscriptionPlanInfoCacheOnce = sync.Once{}
		subscriptionPlanCache = nil
		subscriptionPlanInfoCache = nil
	}
	resetSubscriptionCaches()
	t.Cleanup(resetSubscriptionCaches)

	plan := SubscriptionPlan{
		Id: 4210, Title: "current CNY plan", PriceAmount: 20, Currency: "CNY",
		DurationUnit: SubscriptionDurationMonth, DurationValue: 1, TotalAmount: 500000,
	}
	require.NoError(t, DB.Create(&plan).Error)
	subscription := UserSubscription{Id: 4210, UserId: 4210, PlanId: plan.Id}
	require.NoError(t, DB.Create(&subscription).Error)

	legacyPlan, err := common.Marshal(SubscriptionPlan{
		Id: plan.Id, Title: "legacy USD plan", PriceAmount: 20, Currency: "USD", TotalAmount: 100,
	})
	require.NoError(t, err)
	legacyInfo, err := common.Marshal(SubscriptionPlanInfo{PlanId: plan.Id, PlanTitle: "legacy USD plan"})
	require.NoError(t, err)
	require.NoError(t, common.RDB.Set(t.Context(), "new-api:subscription_plan:v1:4210", legacyPlan, 0).Err())
	require.NoError(t, common.RDB.Set(t.Context(), "new-api:subscription_plan_info:v1:sub:4210", legacyInfo, 0).Err())
	assert.True(t, server.Exists("new-api:subscription_plan:v1:4210"))
	assert.True(t, server.Exists("new-api:subscription_plan_info:v1:sub:4210"))

	gotPlan, err := GetSubscriptionPlanById(plan.Id)
	require.NoError(t, err)
	assert.Equal(t, "CNY", gotPlan.Currency)
	assert.Equal(t, "current CNY plan", gotPlan.Title)
	assert.EqualValues(t, 500000, gotPlan.TotalAmount)

	gotInfo, err := GetSubscriptionPlanInfoByUserSubscriptionId(subscription.Id)
	require.NoError(t, err)
	assert.Equal(t, plan.Id, gotInfo.PlanId)
	assert.Equal(t, "current CNY plan", gotInfo.PlanTitle)
}

func TestValidateOptionValueFreezesCNYLedgerSettings(t *testing.T) {
	require.NoError(t, validateOptionValue("QuotaPerUnit", "500000"))
	require.NoError(t, validateOptionValue("USDExchangeRate", "1"))
	require.NoError(t, validateOptionValue("general_setting.quota_display_type", "CNY"))

	require.Error(t, validateOptionValue("QuotaPerUnit", "1000000"))
	require.Error(t, validateOptionValue("USDExchangeRate", "7.3"))
	require.Error(t, validateOptionValue("general_setting.quota_display_type", "USD"))
}
