package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"

	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

func GetRatioConfig(c *gin.Context) {
	c.Header("X-Billing-Currency", "CNY")
	c.Header("X-Quota-Per-Unit", strconv.FormatFloat(common.QuotaPerUnit, 'f', -1, 64))
	if !ratio_setting.IsExposeRatioEnabled() {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "倍率配置接口未启用",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    billing_setting.GetPricingSyncData(map[string]any(ratio_setting.GetExposedData())),
	})
}
