package email_verification

import (
	"fmt"
	"time"
)

type emailContentGenerator struct{}

func NewEmailContentGenerator() *emailContentGenerator {
	return &emailContentGenerator{}
}

func (*emailContentGenerator) GenerateVerificationEmailContent(code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0; padding:0; background-color:#f4f4f7; font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="background-color:#f4f4f7; padding:40px 20px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" border="0" style="max-width:480px; background-color:#ffffff; border-radius:12px; box-shadow:0 2px 8px rgba(0,0,0,0.04); overflow:hidden;">
          <tr>
            <td style="padding:40px 40px 24px 40px; text-align:center;">
              <h1 style="margin:0 0 8px 0; font-size:24px; font-weight:600; color:#111827;">Verify your email</h1>
              <p style="margin:0; font-size:15px; line-height:1.5; color:#6b7280;">
                Enter the code below to confirm your email address.
              </p>
            </td>
          </tr>
          <tr>
            <td style="padding:8px 40px 32px 40px;" align="center">
              <div style="display:inline-block; padding:20px 32px; background-color:#f3f4f6; border-radius:10px; font-family:'SF Mono','Menlo','Consolas',monospace; font-size:32px; font-weight:600; letter-spacing:8px; color:#111827;">
                %s
              </div>
            </td>
          </tr>
          <tr>
            <td style="padding:0 40px 32px 40px; text-align:center;">
              <p style="margin:0; font-size:13px; line-height:1.5; color:#9ca3af;">
                This code will expire in 15 minutes.<br>
                If you didn't create an account, you can safely ignore this email.
              </p>
            </td>
          </tr>
        </table>
        <p style="margin:24px 0 0 0; font-size:12px; color:#9ca3af;">
          &copy; %d Your Company. All rights reserved.
        </p>
      </td>
    </tr>
  </table>
</body>
</html>`, code, time.Now().Year())
}
