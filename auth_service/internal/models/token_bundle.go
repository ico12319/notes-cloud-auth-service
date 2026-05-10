package models

import "github.com/notes-in-the-cloud/notes-cloud-jwt-utils/accesstoken"

type TokenBundle struct {
	AccessToken  *accesstoken.Token `json:"accessToken"`
	RefreshToken string             `json:"refreshToken"`
}
