package models

type TokenBundle struct {
	AccessToken  *AccessToken `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
}
