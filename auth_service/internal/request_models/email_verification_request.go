package request_models

type EmailVerificationRequest struct {
	VerificationCode string `json:"verificationCode"`
}
