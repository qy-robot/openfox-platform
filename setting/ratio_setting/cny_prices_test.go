package ratio_setting

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuiltinPricesAreStoredAsCNY(t *testing.T) {
	// Domestic provider prices are already quoted in RMB: 0.12 yuan / 1K tokens.
	assert.InDelta(t, 60, GetDefaultModelRatioMap()["ERNIE-4.0-8K"], 1e-10)
	// Preserve the former displayed RMB selling price as a literal default.
	assert.InDelta(t, 9.125, GetDefaultModelRatioMap()["gpt-4o"], 1e-10)
	assert.InDelta(t, 0.292, GetDefaultModelPriceMap()["dall-e-3"], 1e-10)
}
