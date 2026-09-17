package common

import (
	"fmt"
	"html"
)

// RegistrationVerificationEmail renders a self-contained email without remote assets.
func RegistrationVerificationEmail(systemName, code string, validMinutes int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>注册邮箱验证</title></head>
<body style="margin:0;padding:0;background-color:#f4f5f7;color:#202124;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','Microsoft YaHei',Arial,sans-serif;">
<div style="display:none;max-height:0;overflow:hidden;mso-hide:all;">欢迎注册 %[1]s，请使用邮件中的验证码完成邮箱验证。</div>
<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="background-color:#f4f5f7;">
<tr><td align="center" style="padding:36px 16px;">
<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="max-width:560px;">
<tr><td style="padding:0 8px 24px;">
<div style="font-size:26px;line-height:34px;font-weight:700;letter-spacing:-0.7px;">%[1]s</div>
<div style="margin-top:4px;font-size:12px;line-height:20px;color:#656b75;">by擎云机器人</div>
</td></tr>
<tr><td style="padding:32px 24px;background-color:#ffffff;border:1px solid #e4e7ec;border-top:4px solid #202124;border-radius:12px;">
<div style="font-size:12px;line-height:20px;font-weight:600;letter-spacing:2px;color:#777e89;">邮箱验证</div>
<h1 style="margin:12px 0 16px;font-size:24px;line-height:34px;font-weight:700;">欢迎加入 %[1]s</h1>
<p style="margin:0;font-size:15px;line-height:26px;color:#555c66;">距离开启机器人开发之旅，只差一步。<br>请在注册页面输入以下验证码，完成邮箱验证。</p>
<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="margin-top:24px;background-color:#f4f5f7;border:1px solid #e8eaee;border-radius:8px;">
<tr><td align="center" style="padding:22px 8px;">
<div style="font-size:12px;line-height:20px;color:#656b75;">你的注册验证码</div>
<div dir="ltr" style="margin-top:8px;font-family:Consolas,'Courier New',monospace;font-size:34px;line-height:44px;font-weight:700;letter-spacing:6px;color:#17191d;white-space:nowrap;">%[2]s</div>
<div style="margin-top:8px;font-size:13px;line-height:22px;color:#656b75;">有效期为 <strong style="color:#202124;">%[3]d 分钟</strong>，请及时使用</div>
</td></tr></table>
<p style="margin:24px 0 0;font-size:13px;line-height:23px;color:#656b75;">为保护账号安全，请勿将验证码转发或透露给他人。</p>
<p style="margin:12px 0 0;font-size:13px;line-height:23px;color:#656b75;">如果这不是你的操作，请忽略此邮件。</p>
</td></tr>
<tr><td align="center" style="padding:24px 12px 8px;font-size:12px;line-height:22px;color:#777e89;">让机器人开发更简单<br>此邮件由系统自动发送，请勿直接回复。</td></tr>
</table>
</td></tr></table>
</body></html>`, html.EscapeString(NormalizeSystemName(systemName)), html.EscapeString(code), validMinutes)
}
