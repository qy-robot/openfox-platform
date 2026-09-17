package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

type SubscriptionCreemPayRequest struct {
	PlanId int `json:"plan_id"`
}

func SubscriptionRequestCreemPay(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}
	common.ApiErrorMsg(c, "Creem 暂不支持人民币订阅支付")
}
