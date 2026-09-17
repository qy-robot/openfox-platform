package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegistrationVerificationEmail(t *testing.T) {
	body := RegistrationVerificationEmail("RoboCodingAI", "012345", 10)
	assert.Contains(t, body, "欢迎加入 RoboCoding")
	assert.Contains(t, body, "by擎云机器人")
	assert.Contains(t, body, ">012345</div>")
	assert.Contains(t, body, ">10 分钟</strong>")
	assert.NotContains(t, body, "<img")
	assert.NotContains(t, body, "<script")
}

func TestRegistrationVerificationEmailEscapesDynamicContent(t *testing.T) {
	body := RegistrationVerificationEmail(`<img src=x onerror=alert(1)>`, `<script>`, 7)
	assert.NotContains(t, body, "<img")
	assert.NotContains(t, body, "<script>")
	assert.Contains(t, body, "&lt;script&gt;")
	assert.Contains(t, body, ">7 分钟</strong>")
}
