package billingexpr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaleCurrencyPreservesTierAndRequestSemantics(t *testing.T) {
	original := `(len <= 32000 ? tier("short", fixed(0.01)) : tier("long", p * 2 + c * 8)) * (param("fast") == true ? 2 : 1)`
	scaled, err := ScaleCurrency(original, 7.3)
	require.NoError(t, err)

	for _, tc := range []struct {
		params  TokenParams
		request RequestInput
	}{
		{params: TokenParams{P: 100, C: 50, Len: 1000}},
		{params: TokenParams{P: 100, C: 50, Len: 1000}, request: RequestInput{Body: []byte(`{"fast":true}`)}},
		{params: TokenParams{P: 100, C: 50, Len: 50000}},
	} {
		before, _, runErr := RunExprWithRequest(original, tc.params, tc.request)
		require.NoError(t, runErr)
		after, _, runErr := RunExprWithRequest(scaled, tc.params, tc.request)
		require.NoError(t, runErr)
		assert.InDelta(t, before*7.3, after, 1e-9)
	}
}

func TestScaleCurrencyRejectsUnsafeOrInvalidInput(t *testing.T) {
	_, err := ScaleCurrency(`p * 2 + c * 8`, 7.3)
	assert.ErrorContains(t, err, "no tier")
	_, err = ScaleCurrency(`tier("base", p * 2)`, 0)
	assert.ErrorContains(t, err, "finite and positive")
}
