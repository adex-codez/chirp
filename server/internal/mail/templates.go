package mail

import "fmt"

// Brand tokens mirror design.md so the mail reads as Chirp:
// primary #0F2A44, page #F2F5F8, card #FFFFFF, tint #C3D3E2.
const (
	brandNavy = "#0F2A44"
	brandDeep = "#071627"
	brandTint = "#C3D3E2"
	pageBg    = "#F2F5F8"
	cardBg    = "#FFFFFF"
	inkBody   = "#334155"
	inkMuted  = "#64748B"
)

func shell(preheader, eyebrow, title, body, code, footnote string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
</head>
<body style="margin:0;padding:0;background-color:%s;">
<div style="display:none;max-height:0;overflow:hidden;opacity:0;">%s</div>
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:%s;padding:32px 16px;">
<tr><td align="center">
<table role="presentation" width="560" cellpadding="0" cellspacing="0" style="max-width:560px;width:100%%;background-color:%s;border-radius:16px;overflow:hidden;">
<tr><td style="background-color:%s;padding:28px 32px;">
<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:22px;font-weight:700;color:#ffffff;letter-spacing:-0.3px;">&#128038; Chirp</div>
<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:12px;color:%s;letter-spacing:2px;text-transform:uppercase;margin-top:6px;">%s</div>
</td></tr>
<tr><td style="padding:32px 32px 8px 32px;">
<h1 style="margin:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:24px;line-height:32px;font-weight:700;color:%s;">%s</h1>
<p style="margin:12px 0 0 0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:16px;line-height:28px;color:%s;">%s</p>
</td></tr>
<tr><td align="center" style="padding:20px 32px 8px 32px;">
<div style="display:inline-block;border:2px dashed %s;border-radius:12px;padding:18px 36px;background-color:%s;">
<div style="font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:36px;font-weight:700;letter-spacing:10px;color:%s;padding-left:10px;">%s</div>
</div>
<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:12px;line-height:20px;color:%s;margin-top:12px;">Valid for 20 minutes. It expires after 5 wrong tries.</div>
</td></tr>
<tr><td style="padding:20px 32px 32px 32px;">
<p style="margin:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:12px;line-height:20px;color:%s;">%s</p>
</td></tr>
<tr><td style="background-color:%s;padding:16px 32px;border-top:1px solid %s;">
<p style="margin:0;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;font-size:12px;line-height:20px;color:%s;">Sent by Chirp &middot; every participant has a verified identity.</p>
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`,
		title,
		pageBg, preheader,
		pageBg, cardBg,
		brandNavy, brandTint, eyebrow,
		brandDeep, title,
		inkBody, body,
		brandNavy, pageBg,
		brandNavy, code,
		inkMuted,
		inkMuted, footnote,
		pageBg, brandTint,
		inkMuted,
	)
}

// VerifyEmailHTML renders the signup/resend verification message.
func VerifyEmailHTML(code string) string {
	return shell(
		"Your Chirp verification code is inside — it expires in 20 minutes.",
		"Verified identity",
		"Verify your Email",
		"Welcome to Chirp — enter this code in the app to prove this Email is yours and finish creating your account.",
		code,
		"If you did not create a Chirp account, you can safely ignore this Email.",
	)
}

// ResetPasswordHTML renders the password-reset message.
func ResetPasswordHTML(code string) string {
	return shell(
		"Your Chirp password-reset code is inside — it expires in 20 minutes.",
		"Password reset",
		"Reset your password",
		"Enter this code in the app to choose a new password. Your existing sessions stay signed in until the reset completes.",
		code,
		"If you did not ask for a reset, you can safely ignore this Email — your password stays unchanged.",
	)
}
