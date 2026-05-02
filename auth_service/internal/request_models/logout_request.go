package request_models

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}
