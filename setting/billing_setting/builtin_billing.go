package billing_setting

// Built-in token prices are fixed CNY selling prices per million tokens.
// These preserve the former default RMB catalog; operator prices override them.
var builtinBillingExpr = map[string]string{
	// https://developers.openai.com/api/docs/pricing (Standard, 2026-09-09).
	// The Images API reports image output in output_tokens, normalized to c.
	"gpt-image-2":            `tier("standard", p * 36.5 + cr * 9.125 + img * 58.4 + img_cr * 14.6 + c * 219)`,
	"gpt-image-2.5-sunburst": `tier("standard", p * 36.5 + cr * 9.125 + img * 58.4 + img_cr * 14.6 + c * 219)`,
	"gpt-image-2.5-flare":    `tier("standard", p * 36.5 + cr * 9.125 + img * 58.4 + img_cr * 14.6 + c * 219)`,
	// https://developers.openai.com/api/docs/models/gpt-6-astra
	// Standard pricing; the long-context rates apply to the whole request.
	// Do not infer service-tier discounts from incoming request parameters:
	// channels filter service_tier by default, so it may not reach the upstream.
	"gpt-6-astra": `len <= 272000 ? tier("standard", p * 73 + c * 365 + cr * 7.3 + cc * 91.25) : tier("long_context", p * 146 + c * 547.5 + cr * 14.6 + cc * 182.5)`,
}
