package operation_setting

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPointsPreserveCNYValueAndLedgerPrecision(t *testing.T) {
	oldQuota := common.QuotaPerUnit
	t.Cleanup(func() { common.QuotaPerUnit = oldQuota })
	common.QuotaPerUnit = 500000
	assert.Equal(t, float64(50000), QuotaPerPoint())
	assert.Equal(t, float64(10), common.QuotaPerUnit/QuotaPerPoint(), "one CNY must display as ten points")

	common.QuotaPerUnit = 0
	assert.Zero(t, QuotaPerPoint())
}
