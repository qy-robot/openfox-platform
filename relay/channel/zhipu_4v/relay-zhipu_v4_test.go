package zhipu_4v

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestRequestOpenAI2ZhipuMaxTokens(t *testing.T) {
	tests := []struct {
		name       string
		maxTokens  *uint
		wantTokens *uint
	}{
		{
			name:       "omitted max_tokens is not injected",
			maxTokens:  nil,
			wantTokens: nil,
		},
		{
			name:       "in-range max_tokens is preserved",
			maxTokens:  lo.ToPtr[uint](4096),
			wantTokens: lo.ToPtr[uint](4096),
		},
		{
			name:       "max_tokens at the ceiling is preserved",
			maxTokens:  lo.ToPtr[uint](131072),
			wantTokens: lo.ToPtr[uint](131072),
		},
		{
			name:       "max_tokens above the ceiling is clamped",
			maxTokens:  lo.ToPtr[uint](256000),
			wantTokens: lo.ToPtr[uint](131072),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requestOpenAI2Zhipu(dto.GeneralOpenAIRequest{
				Model:     "glm-5.3-flash",
				Messages:  []dto.Message{{Role: "user", Content: "hi"}},
				MaxTokens: tt.maxTokens,
			})

			if tt.wantTokens == nil {
				assert.Nil(t, got.MaxTokens)
				return
			}
			assert.Equal(t, *tt.wantTokens, *got.MaxTokens)
		})
	}
}
