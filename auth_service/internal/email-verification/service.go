package email_verification

import (
	"context"
	"fmt"
	"github.com/resend/resend-go/v2"
)

type verificationEmailContentGenerator interface {
	GenerateVerificationEmailContent(code string) string
}

type service struct {
	client                            *resend.Client
	fromAddress                       string
	verificationEmailContentGenerator verificationEmailContentGenerator
}

func NewService(
	apiKey,
	fromAddress string,
	verificationEmailContentGenerator verificationEmailContentGenerator,
) *service {
	return &service{
		client:                            resend.NewClient(apiKey),
		fromAddress:                       fromAddress,
		verificationEmailContentGenerator: verificationEmailContentGenerator,
	}
}

func (s *service) SendVerificationEmail(
	ctx context.Context,
	to string,
	code string,
) error {
	params := &resend.SendEmailRequest{
		From:    s.fromAddress,
		To:      []string{to},
		Subject: "Email verification",
		Html:    s.verificationEmailContentGenerator.GenerateVerificationEmailContent(code),
	}

	if _, err := s.client.Emails.SendWithContext(ctx, params); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}
