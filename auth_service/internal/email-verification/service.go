package email_verification

import (
	"context"
	"fmt"
	"github.com/resend/resend-go/v2"
	"text/template"
)

type ResendService struct {
	client      *resend.Client
	fromAddress string
	frontendURL string
}

func NewResendService(apiKey, fromAddress, frontendURL string) *ResendService {
	return &ResendService{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
		frontendURL: frontendURL,
	}
}

func (s *ResendService) SendVerificationEmail(
	ctx context.Context,
	to string,
	token string,
) error {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.frontendURL, template.URLQueryEscaper(token))

	htmlBody := fmt.Sprintf(`
		<h2>Verify your email</h2>
		<p>Click the link below to verify your email address:</p>
		<p><a href="%s">Verify email</a></p>
		<p>If you did not create this account, you can ignore this email.</p>
	`, verifyURL)

	params := &resend.SendEmailRequest{
		From:    s.fromAddress,
		To:      []string{to},
		Subject: "Verify your email address",
		Html:    htmlBody,
	}

	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}
