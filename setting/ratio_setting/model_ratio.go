package ratio_setting

import (
	"maps"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	hostreasoning "github.com/QuantumNous/new-api/setting/reasoning"
	"github.com/QuantumNous/new-api/types"
)

// from songquanpeng/one-api
// Catalog values are fixed RMB selling prices; no exchange-rate conversion.

// modelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// TODO: when a new api is enabled, check the pricing here
// 1 quota per token === ￥2 / 1M tokens (500000 quota per yuan).

var defaultModelRatio = map[string]float64{
	//"midjourney":                50,
	"gpt-4-gizmo-*":                            109.5,
	"gpt-4o-gizmo-*":                           18.25,
	"gpt-4-all":                                109.5,
	"gpt-4o-all":                               109.5,
	"gpt-4":                                    109.5,
	"gpt-4-0613":                               109.5,
	"gpt-4-32k":                                219,
	"gpt-4-32k-0613":                           219,
	"gpt-4-1106-preview":                       36.5,
	"gpt-4-0125-preview":                       36.5,
	"gpt-4-turbo-preview":                      36.5,
	"gpt-4-vision-preview":                     36.5,
	"gpt-4-1106-vision-preview":                36.5,
	"chatgpt-4o-latest":                        18.25,
	"gpt-4o":                                   9.125,
	"gpt-4o-audio-preview":                     9.125,
	"gpt-4o-audio-preview-2024-10-01":          9.125,
	"gpt-4o-2024-05-13":                        18.25,
	"gpt-4o-2024-08-06":                        9.125,
	"gpt-4o-2024-11-20":                        9.125,
	"gpt-4o-realtime-preview":                  18.25,
	"gpt-4o-realtime-preview-2024-10-01":       18.25,
	"gpt-4o-realtime-preview-2024-12-17":       18.25,
	"gpt-4o-mini-realtime-preview":             2.19,
	"gpt-4o-mini-realtime-preview-2024-12-17":  2.19,
	"gpt-4.1":                                  7.3,
	"gpt-4.1-2025-04-14":                       7.3,
	"gpt-4.1-mini":                             1.46,
	"gpt-4.1-mini-2025-04-14":                  1.46,
	"gpt-4.1-nano":                             0.365,
	"gpt-4.1-nano-2025-04-14":                  0.365,
	"gpt-image-1":                              18.25,
	"o1":                                       54.75,
	"o1-2024-12-17":                            54.75,
	"o1-preview":                               54.75,
	"o1-preview-2024-09-12":                    54.75,
	"o1-mini":                                  4.015,
	"o1-mini-2024-09-12":                       4.015,
	"o1-pro":                                   547.5,
	"o1-pro-2025-03-19":                        547.5,
	"o3-mini":                                  4.015,
	"o3-mini-2025-01-31":                       4.015,
	"o3-mini-high":                             4.015,
	"o3-mini-2025-01-31-high":                  4.015,
	"o3-mini-low":                              4.015,
	"o3-mini-2025-01-31-low":                   4.015,
	"o3-mini-medium":                           4.015,
	"o3-mini-2025-01-31-medium":                4.015,
	"o3":                                       7.3,
	"o3-2025-04-16":                            7.3,
	"o3-pro":                                   73,
	"o3-pro-2025-06-10":                        73,
	"o3-deep-research":                         36.5,
	"o3-deep-research-2025-06-26":              36.5,
	"o4-mini":                                  4.015,
	"o4-mini-2025-04-16":                       4.015,
	"o4-mini-deep-research":                    7.3,
	"o4-mini-deep-research-2025-06-26":         7.3,
	"gpt-4o-mini":                              0.5475,
	"gpt-4o-mini-2024-07-18":                   0.5475,
	"gpt-4-turbo":                              36.5,
	"gpt-4-turbo-2024-04-09":                   36.5,
	"gpt-4.5-preview":                          273.75,
	"gpt-4.5-preview-2025-02-27":               273.75,
	"gpt-5":                                    4.5625,
	"gpt-5-2025-08-07":                         4.5625,
	"gpt-5-chat-latest":                        4.5625,
	"gpt-5-mini":                               0.9125,
	"gpt-5-mini-2025-08-07":                    0.9125,
	"gpt-5-nano":                               0.1825,
	"gpt-5-nano-2025-08-07":                    0.1825,
	"gpt-5.5":                                  18.25,
	"gpt-5.6-sol":                              18.25,
	"gpt-5.6-terra":                            9.125,
	"gpt-5.6-luna":                             3.65,
	"gpt-3.5-turbo":                            1.825,
	"gpt-3.5-turbo-0613":                       5.475,
	"gpt-3.5-turbo-16k":                        10.95,
	"gpt-3.5-turbo-16k-0613":                   10.95,
	"gpt-3.5-turbo-instruct":                   5.475,
	"gpt-3.5-turbo-1106":                       3.65,
	"gpt-3.5-turbo-0125":                       1.825,
	"text-ada-001":                             1.46,
	"text-babbage-001":                         1.825,
	"text-curie-001":                           7.3,
	"text-davinci-edit-001":                    73,
	"code-davinci-edit-001":                    73,
	"whisper-1":                                109.5,
	"tts-1":                                    54.75,
	"tts-1-1106":                               54.75,
	"tts-1-hd":                                 109.5,
	"tts-1-hd-1106":                            109.5,
	"davinci":                                  73,
	"curie":                                    73,
	"text-embedding-3-small":                   0.073,
	"text-embedding-3-large":                   0.4745,
	"text-embedding-ada-002":                   0.365,
	"text-search-ada-doc-001":                  73,
	"text-moderation-stable":                   0.73,
	"text-moderation-latest":                   0.73,
	"claude-3-haiku-20240307":                  0.9125,
	"claude-3-5-haiku-20241022":                3.65,
	"claude-haiku-4-5-20251001":                3.65,
	"claude-3-sonnet-20240229":                 10.95,
	"claude-3-5-sonnet-20240620":               10.95,
	"claude-3-5-sonnet-20241022":               10.95,
	"claude-3-7-sonnet-20250219":               10.95,
	"claude-3-7-sonnet-20250219-thinking":      10.95,
	"claude-sonnet-4-20250514":                 10.95,
	"claude-sonnet-4-5-20250929":               10.95,
	"claude-opus-4-5-20251101":                 18.25,
	"claude-opus-4-6":                          18.25,
	"claude-opus-4-6-max":                      18.25,
	"claude-opus-4-6-high":                     18.25,
	"claude-opus-4-6-medium":                   18.25,
	"claude-opus-4-6-low":                      18.25,
	"claude-opus-4-7":                          18.25,
	"claude-opus-4-7-max":                      18.25,
	"claude-opus-4-7-xhigh":                    18.25,
	"claude-opus-4-7-high":                     18.25,
	"claude-opus-4-7-medium":                   18.25,
	"claude-opus-4-7-low":                      18.25,
	"claude-opus-4-8":                          18.25,
	"claude-opus-4-8-max":                      18.25,
	"claude-opus-4-8-xhigh":                    18.25,
	"claude-opus-4-8-high":                     18.25,
	"claude-opus-4-8-medium":                   18.25,
	"claude-opus-4-8-low":                      18.25,
	"claude-3-opus-20240229":                   54.75,
	"claude-opus-4-20250514":                   54.75,
	"claude-opus-4-1-20250805":                 54.75,
	"ERNIE-4.0-8K":                             60,
	"ERNIE-3.5-8K":                             6,
	"ERNIE-3.5-8K-0205":                        12,
	"ERNIE-3.5-8K-1222":                        6,
	"ERNIE-Bot-8K":                             12,
	"ERNIE-3.5-4K-0205":                        6,
	"ERNIE-Speed-8K":                           2,
	"ERNIE-Speed-128K":                         2,
	"ERNIE-Lite-8K-0922":                       4,
	"ERNIE-Lite-8K-0308":                       1.5,
	"ERNIE-Tiny-8K":                            0.5,
	"BLOOMZ-7B":                                2,
	"Embedding-V1":                             1,
	"bge-large-zh":                             1,
	"bge-large-en":                             1,
	"tao-8k":                                   1,
	"PaLM-2":                                   7.3,
	"gemini-1.5-pro-latest":                    9.125,
	"gemini-1.5-flash-latest":                  0.5475,
	"gemini-2.0-flash":                         0.365,
	"gemini-2.5-pro-exp-03-25":                 4.5625,
	"gemini-2.5-pro-preview-03-25":             4.5625,
	"gemini-2.5-pro":                           4.5625,
	"gemini-2.5-flash-preview-04-17":           0.5475,
	"gemini-2.5-flash-preview-04-17-thinking":  0.5475,
	"gemini-2.5-flash-preview-05-20":           0.5475,
	"gemini-2.5-flash-preview-05-20-thinking":  0.5475,
	"gemini-2.5-flash-thinking-*":              0.5475, // 用于为后续所有2.5 flash thinking budget 模型设置默认倍率
	"gemini-2.5-pro-thinking-*":                4.5625, // 用于为后续所有2.5 pro thinking budget 模型设置默认倍率
	"gemini-2.5-flash-lite-preview-thinking-*": 0.365,
	"gemini-2.5-flash-lite-preview-06-17":      0.365,
	"gemini-2.5-flash":                         1.095,
	"gemini-robotics-er-1.5-preview":           1.095,
	"gemini-embedding-001":                     0.5475,
	"text-embedding-004":                       0.0073,
	"chatglm_turbo":                            2.60756, // ￥0.005 / 1k tokens
	"chatglm_pro":                              5.21439, // ￥0.01 / 1k tokens
	"chatglm_std":                              2.60756, // ￥0.005 / 1k tokens
	"chatglm_lite":                             1.04317, // ￥0.002 / 1k tokens
	"glm-4":                                    52.1439, // ￥0.1 / 1k tokens
	"glm-4v":                                   25,      // ￥0.05 / 1k tokens
	"glm-4-alltools":                           50,      // ￥0.1 / 1k tokens
	"glm-3-turbo":                              2.60756,
	"glm-4-plus":                               25,
	"glm-4-0520":                               50,
	"glm-4-air":                                0.5,
	"glm-4-airx":                               5,
	"glm-4-long":                               0.5,
	"glm-4-flash":                              0,
	"glm-4v-plus":                              5,
	"qwen-turbo":                               6.25756, // ￥0.012 / 1k tokens
	"qwen-plus":                                73,      // ￥0.14 / 1k tokens
	"text-embedding-v1":                        0.365,   // ￥0.0007 / 1k tokens
	"SparkDesk-v1.1":                           9.38634, // ￥0.018 / 1k tokens
	"SparkDesk-v2.1":                           9.38634, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1":                           9.38634, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5":                           9.38634, // ￥0.018 / 1k tokens
	"SparkDesk-v4.0":                           9.38634,
	"hunyuan":                                  52.1439, // ¥0.1 / 1k tokens  // https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
	// https://platform.lingyiwanwu.com/docs#-计费单元
	// Fixed RMB catalog values.
	"yi-34b-chat-0205":       1.314,
	"yi-34b-chat-200k":       6.3072,
	"yi-vl-plus":             3.1536,
	"yi-large":               10,
	"yi-medium":              1.25,
	"yi-vision":              3,
	"yi-medium-200k":         6,
	"yi-spark":               0.5,
	"yi-large-rag":           12.5,
	"yi-large-turbo":         6,
	"yi-large-preview":       10,
	"yi-large-rag-preview":   12.5,
	"command":                3.65,
	"command-nightly":        3.65,
	"command-light":          3.65,
	"command-light-nightly":  3.65,
	"command-r":              1.825,
	"command-r-plus":         10.95,
	"command-r-08-2024":      0.5475,
	"command-r-plus-08-2024": 9.125,
	"deepseek-chat":          0.9855,
	"deepseek-coder":         0.9855,
	"deepseek-reasoner":      2.0075, // 0.55 / 1k tokens
	// Perplexity online 模型对搜索额外收费，有需要应自行调整，此处不计入搜索费用
	"llama-3-sonar-small-32k-chat":   0.73,
	"llama-3-sonar-small-32k-online": 0.73,
	"llama-3-sonar-large-32k-chat":   3.65,
	"llama-3-sonar-large-32k-online": 3.65,
	// grok
	"grok-3-beta":           10.95,
	"grok-3-mini-beta":      1.095,
	"grok-2":                7.3,
	"grok-2-vision":         7.3,
	"grok-beta":             18.25,
	"grok-vision-beta":      18.25,
	"grok-3-fast-beta":      18.25,
	"grok-3-mini-fast-beta": 2.19,
	// submodel
	"NousResearch/Hermes-4-405B-FP8":          5.84,
	"Qwen/Qwen3-235B-A22B-Thinking-2507":      4.38,
	"Qwen/Qwen3-Coder-480B-A35B-Instruct-FP8": 5.84,
	"Qwen/Qwen3-235B-A22B-Instruct-2507":      2.19,
	"zai-org/GLM-4.5-FP8":                     5.84,
	"openai/gpt-oss-120b":                     3.65,
	"deepseek-ai/DeepSeek-R1-0528":            5.84,
	"deepseek-ai/DeepSeek-R1":                 5.84,
	"deepseek-ai/DeepSeek-V3-0324":            5.84,
	"deepseek-ai/DeepSeek-V3.1":               5.84,
}

var defaultModelPrice = map[string]float64{
	"suno_music":                     0.73,
	"suno_lyrics":                    0.073,
	"dall-e-3":                       0.292,
	"imagen-3.0-generate-002":        0.219,
	"black-forest-labs/flux-1.1-pro": 0.292,
	"gpt-4-gizmo-*":                  0.73,
	"mj_video":                       5.84,
	"mj_imagine":                     0.73,
	"mj_edits":                       0.73,
	"mj_variation":                   0.73,
	"mj_reroll":                      0.73,
	"mj_blend":                       0.73,
	"mj_modal":                       0.73,
	"mj_zoom":                        0.73,
	"mj_shorten":                     0.73,
	"mj_high_variation":              0.73,
	"mj_low_variation":               0.73,
	"mj_pan":                         0.73,
	"mj_inpaint":                     0,
	"mj_custom_zoom":                 0,
	"mj_describe":                    0.365,
	"mj_upscale":                     0.365,
	"swap_face":                      0.365,
	"mj_upload":                      0.365,
	"sora-2":                         2.19,
	"sora-2-pro":                     3.65,
	"gpt-4o-mini-tts":                2.19,
	"veo-3.0-generate-001":           2.92,
	"veo-3.0-fast-generate-001":      1.095,
	"veo-3.1-generate-preview":       2.92,
	"veo-3.1-fast-generate-preview":  1.095,
}

var defaultAudioRatio = map[string]float64{
	"gpt-4o-audio-preview":         16,
	"gpt-4o-mini-audio-preview":    66.67,
	"gpt-4o-realtime-preview":      8,
	"gpt-4o-mini-realtime-preview": 16.67,
	"gpt-4o-mini-tts":              25,
}

var defaultAudioCompletionRatio = map[string]float64{
	"gpt-4o-realtime":      2,
	"gpt-4o-mini-realtime": 2,
	"gpt-4o-mini-tts":      1,
	"tts-1":                0,
	"tts-1-hd":             0,
	"tts-1-1106":           0,
	"tts-1-hd-1106":        0,
}

var modelPriceMap = types.NewRWMap[string, float64]()
var modelRatioMap = types.NewRWMap[string, float64]()
var completionRatioMap = types.NewRWMap[string, float64]()

var defaultCompletionRatio = map[string]float64{
	"gpt-4-gizmo-*":  2,
	"gpt-4o-gizmo-*": 3,
	"gpt-4-all":      2,
	"gpt-image-1":    8,
}

// InitRatioSettings initializes all model related settings maps
func InitRatioSettings() {
	modelPriceMap.AddAll(defaultModelPrice)
	modelRatioMap.AddAll(defaultModelRatio)
	completionRatioMap.AddAll(defaultCompletionRatio)
	cacheRatioMap.AddAll(defaultCacheRatio)
	createCacheRatioMap.AddAll(defaultCreateCacheRatio)
	imageRatioMap.AddAll(defaultImageRatio)
	audioRatioMap.AddAll(defaultAudioRatio)
	audioCompletionRatioMap.AddAll(defaultAudioCompletionRatio)
}

func GetModelPriceMap() map[string]float64 {
	return modelPriceMap.ReadAll()
}

func ModelPrice2JSONString() string {
	return modelPriceMap.MarshalJSONString()
}

func UpdateModelPriceByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(modelPriceMap, jsonStr, InvalidateExposedDataCache)
}

// GetModelPrice 返回模型的价格，如果模型不存在则返回-1，false
func GetModelPrice(name string, printErr bool) (float64, bool) {
	name = FormatMatchingModelName(name)

	if price, ok := modelPriceMap.Get(name); ok {
		return price, true
	}

	if printErr {
		common.SysError("model price not found: " + name)
	}
	return -1, false
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(modelRatioMap, jsonStr, InvalidateExposedDataCache)
}

// 处理带有思考预算的模型名称，方便统一定价
func handleThinkingBudgetModel(name, prefix, wildcard string) string {
	if strings.HasPrefix(name, prefix) && strings.Contains(name, "-thinking-") {
		return wildcard
	}
	return name
}

func GetModelRatio(name string) (float64, bool, string) {
	name = FormatMatchingModelName(name)

	ratio, ok := modelRatioMap.Get(name)
	if !ok {
		return 37.5, operation_setting.SelfUseModeEnabled, name
	}
	return ratio, true, name
}

func DefaultModelRatio2JSONString() string {
	jsonBytes, err := common.Marshal(defaultModelRatio)
	if err != nil {
		common.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func GetDefaultModelRatioMap() map[string]float64 {
	return defaultModelRatio
}

func GetDefaultModelPriceMap() map[string]float64 {
	return defaultModelPrice
}

// GetDefaultPricingMaps returns independent copies for model-level reset and
// first-write initialization; callers cannot mutate the built-in defaults.
func GetDefaultPricingMaps() map[string]map[string]float64 {
	defaults := map[string]map[string]float64{
		"ModelPrice": defaultModelPrice, "ModelRatio": defaultModelRatio,
		"CompletionRatio": defaultCompletionRatio, "CacheRatio": defaultCacheRatio,
		"CreateCacheRatio": defaultCreateCacheRatio, "ImageRatio": defaultImageRatio,
		"AudioRatio": defaultAudioRatio, "AudioCompletionRatio": defaultAudioCompletionRatio,
	}
	result := make(map[string]map[string]float64, len(defaults))
	for key, values := range defaults {
		result[key] = make(map[string]float64, len(values))
		maps.Copy(result[key], values)
	}
	return result
}

func CompletionRatio2JSONString() string {
	return completionRatioMap.MarshalJSONString()
}

func UpdateCompletionRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(completionRatioMap, jsonStr, InvalidateExposedDataCache)
}

func GetCompletionRatio(name string) float64 {
	return GetCompletionRatioInfo(name).Ratio
}

type CompletionRatioInfo struct {
	Ratio  float64 `json:"ratio"`
	Locked bool    `json:"locked"`
}

func GetCompletionRatioInfo(name string) CompletionRatioInfo {
	name = FormatMatchingModelName(name)
	var configured *float64
	if ratio, ok := completionRatioMap.Get(name); ok {
		configured = &ratio
	}
	return ResolveCompletionRatio(name, configured)
}

// ResolveCompletionRatio applies relay's enforced and fallback ratios to a
// configuration snapshot or draft without consulting mutable saved settings.
func ResolveCompletionRatio(name string, configured *float64) CompletionRatioInfo {
	name = FormatMatchingModelName(name)
	if strings.Contains(name, "/") && configured != nil {
		return CompletionRatioInfo{Ratio: *configured}
	}

	hardCodedRatio, locked := getHardcodedCompletionModelRatio(name)
	if locked {
		return CompletionRatioInfo{
			Ratio:  hardCodedRatio,
			Locked: true,
		}
	}

	if configured != nil {
		return CompletionRatioInfo{
			Ratio:  *configured,
			Locked: false,
		}
	}

	return CompletionRatioInfo{
		Ratio:  hardCodedRatio,
		Locked: false,
	}
}

func getHardcodedCompletionModelRatio(name string) (float64, bool) {

	isReservedModel := strings.HasSuffix(name, "-all") || strings.HasSuffix(name, "-gizmo-*")
	if isReservedModel {
		return 2, false
	}

	if strings.HasPrefix(name, "gpt-") {
		if strings.HasPrefix(name, "gpt-4o") {
			if name == "gpt-4o-2024-05-13" {
				return 3, true
			}
			if strings.HasPrefix(name, "gpt-4o-mini-tts") {
				return 20, false
			}
			return 4, false
		}
		// gpt-5 匹配
		if strings.HasPrefix(name, "gpt-5") {
			if !strings.Contains(name, ".") {
				return 8, true
			}
			if strings.HasPrefix(name, "gpt-5.4") {
				if strings.HasPrefix(name, "gpt-5.4-nano") {
					return 6.25, true
				}
				return 6, true
			}
			// gpt-5.5 and later models are unlocked
			return 6, false
		}
		// gpt-4.5-preview匹配
		if strings.HasPrefix(name, "gpt-4.5-preview") {
			return 2, true
		}
		if strings.HasPrefix(name, "gpt-4-turbo") || strings.HasSuffix(name, "gpt-4-1106") || strings.HasSuffix(name, "gpt-4-1105") {
			return 3, true
		}
		// 没有特殊标记的 gpt-4 模型默认倍率为 2
		return 2, false
	}
	if strings.HasPrefix(name, "o1") || strings.HasPrefix(name, "o3") {
		return 4, true
	}
	if name == "chatgpt-4o-latest" {
		return 3, true
	}

	if strings.Contains(name, "claude-3") {
		return 5, true
	} else if strings.Contains(name, "claude-sonnet-4") || strings.Contains(name, "claude-opus-4") || strings.Contains(name, "claude-haiku-4") {
		return 5, true
	}

	if strings.HasPrefix(name, "gpt-3.5") {
		if name == "gpt-3.5-turbo" || strings.HasSuffix(name, "0125") {
			// https://openai.com/blog/new-embedding-models-and-api-updates
			// Updated GPT-3.5 Turbo model and lower pricing
			return 3, true
		}
		if strings.HasSuffix(name, "1106") {
			return 2, true
		}
		return 4.0 / 3.0, true
	}
	if strings.HasPrefix(name, "mistral-") {
		return 3, true
	}
	if strings.HasPrefix(name, "gemini-") {
		if strings.HasPrefix(name, "gemini-1.5") {
			return 4, true
		} else if strings.HasPrefix(name, "gemini-2.0") {
			return 4, true
		} else if strings.HasPrefix(name, "gemini-2.5-pro") { // 移除preview来增加兼容性，这里假设正式版的倍率和preview一致
			return 8, false
		} else if strings.HasPrefix(name, "gemini-2.5-flash") { // 处理不同的flash模型倍率
			if strings.HasPrefix(name, "gemini-2.5-flash-preview") {
				return 3.5 / 0.15, false
			}
			if strings.HasPrefix(name, "gemini-2.5-flash-lite") {
				return 4, false
			}
			return 2.5 / 0.3, false
		} else if strings.HasPrefix(name, "gemini-robotics-er-1.5") {
			return 2.5 / 0.3, false
		} else if strings.HasPrefix(name, "gemini-3-pro") {
			if strings.HasPrefix(name, "gemini-3-pro-image") {
				return 60, false
			}
			return 6, false
		}
		return 4, false
	}
	if strings.HasPrefix(name, "command") {
		switch name {
		case "command-r":
			return 3, true
		case "command-r-plus":
			return 5, true
		case "command-r-08-2024":
			return 4, true
		case "command-r-plus-08-2024":
			return 4, true
		default:
			return 4, false
		}
	}
	// hint 只给官方上4倍率，由于开源模型供应商自行定价，不对其进行补全倍率进行强制对齐
	if strings.HasPrefix(name, "ERNIE-Speed-") {
		return 2, true
	} else if strings.HasPrefix(name, "ERNIE-Lite-") {
		return 2, true
	} else if strings.HasPrefix(name, "ERNIE-Character") {
		return 2, true
	} else if strings.HasPrefix(name, "ERNIE-Functions") {
		return 2, true
	}
	switch name {
	case "llama2-70b-4096":
		return 0.8 / 0.64, true
	case "llama3-8b-8192":
		return 2, true
	case "llama3-70b-8192":
		return 0.79 / 0.59, true
	}
	return 1, false
}

func GetAudioRatio(name string) float64 {
	name = FormatMatchingModelName(name)
	if ratio, ok := audioRatioMap.Get(name); ok {
		return ratio
	}
	return 1
}

func GetAudioCompletionRatio(name string) float64 {
	name = FormatMatchingModelName(name)
	if ratio, ok := audioCompletionRatioMap.Get(name); ok {
		return ratio
	}
	return 1
}

func ContainsAudioRatio(name string) bool {
	name = FormatMatchingModelName(name)
	_, ok := audioRatioMap.Get(name)
	return ok
}

func ContainsAudioCompletionRatio(name string) bool {
	name = FormatMatchingModelName(name)
	_, ok := audioCompletionRatioMap.Get(name)
	return ok
}

func ModelRatio2JSONString() string {
	return modelRatioMap.MarshalJSONString()
}

var defaultImageRatio = map[string]float64{
	"gpt-image-1": 2,
}
var imageRatioMap = types.NewRWMap[string, float64]()
var audioRatioMap = types.NewRWMap[string, float64]()
var audioCompletionRatioMap = types.NewRWMap[string, float64]()

func ImageRatio2JSONString() string {
	return imageRatioMap.MarshalJSONString()
}

func UpdateImageRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(imageRatioMap, jsonStr)
}

const DefaultImageRatio = 1.0

func GetImageRatio(name string) (float64, bool) {
	ratio, ok := imageRatioMap.Get(name)
	if !ok {
		return DefaultImageRatio, false
	}
	return ratio, true
}

func AudioRatio2JSONString() string {
	return audioRatioMap.MarshalJSONString()
}

func UpdateAudioRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(audioRatioMap, jsonStr, InvalidateExposedDataCache)
}

func AudioCompletionRatio2JSONString() string {
	return audioCompletionRatioMap.MarshalJSONString()
}

func UpdateAudioCompletionRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(audioCompletionRatioMap, jsonStr, InvalidateExposedDataCache)
}

func GetModelRatioCopy() map[string]float64 {
	return modelRatioMap.ReadAll()
}

func GetModelPriceCopy() map[string]float64 {
	return modelPriceMap.ReadAll()
}

func GetCompletionRatioCopy() map[string]float64 {
	return completionRatioMap.ReadAll()
}

func GetImageRatioCopy() map[string]float64 {
	return imageRatioMap.ReadAll()
}

func GetAudioRatioCopy() map[string]float64 {
	return audioRatioMap.ReadAll()
}

func GetAudioCompletionRatioCopy() map[string]float64 {
	return audioCompletionRatioMap.ReadAll()
}

// RoutingMatchModelName returns the name used for channel-ability and token-limit
// fallback matching: strip @ modifiers and legacy aliases first, then apply
// wildcard normalization.
func RoutingMatchModelName(name string) string {
	return FormatMatchingModelName(hostreasoning.BaseModelName(name))
}

// HasConfiguredModelRatio reports whether name has an explicit ratio entry
// after wildcard normalization. Self-use fallback does not count.
func HasConfiguredModelRatio(name string) bool {
	name = FormatMatchingModelName(name)
	_, ok := modelRatioMap.Get(name)
	return ok
}

// 转换模型名，减少渠道必须配置各种带参数模型
func FormatMatchingModelName(name string) string {
	if strings.HasPrefix(name, "gemini-2.5-flash-lite") {
		name = handleThinkingBudgetModel(name, "gemini-2.5-flash-lite", "gemini-2.5-flash-lite-thinking-*")
	} else if strings.HasPrefix(name, "gemini-2.5-flash") {
		name = handleThinkingBudgetModel(name, "gemini-2.5-flash", "gemini-2.5-flash-thinking-*")
	} else if strings.HasPrefix(name, "gemini-2.5-pro") {
		name = handleThinkingBudgetModel(name, "gemini-2.5-pro", "gemini-2.5-pro-thinking-*")
	}

	if strings.HasPrefix(name, "gpt-4-gizmo") {
		name = "gpt-4-gizmo-*"
	}
	if strings.HasPrefix(name, "gpt-4o-gizmo") {
		name = "gpt-4o-gizmo-*"
	}
	return name
}

// result: 倍率or价格， usePrice， exist
func GetModelRatioOrPrice(model string) (float64, bool, bool) { // price or ratio
	price, usePrice := GetModelPrice(model, false)
	if usePrice {
		return price, true, true
	}
	modelRatio, success, _ := GetModelRatio(model)
	if success {
		return modelRatio, false, true
	}
	return 37.5, false, false
}
