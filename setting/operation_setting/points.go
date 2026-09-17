package operation_setting

import (
	"github.com/QuantumNous/new-api/common"
	"math"
)

const BillingCurrency = "CNY"
const PointsPerCNY = 10

// QuotaPerPoint keeps the existing integer ledger precision while exposing
// the product rule that one yuan equals ten points.
func QuotaPerPoint() float64 {
	if common.QuotaPerUnit <= 0 || math.IsNaN(common.QuotaPerUnit) || math.IsInf(common.QuotaPerUnit, 0) {
		return 0
	}
	value := common.QuotaPerUnit / PointsPerCNY
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return 0
	}
	return value
}
