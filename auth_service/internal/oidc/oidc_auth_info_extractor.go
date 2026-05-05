package oidc

import (
	"context"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/oauth"
	"golang.org/x/oauth2"
	"log"
)

type oidcUserAuthInfoExtractor struct {
	provider *oidcProvider
}

func NewOIDCUserAuthInfoExtractor(provider *oidcProvider) *oidcUserAuthInfoExtractor {
	return &oidcUserAuthInfoExtractor{provider: provider}
}

func (o *oidcUserAuthInfoExtractor) Extract(ctx context.Context, token *oauth2.Token, session *oauth.OAuthSession) (*models.UserAuthInfo, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("id token is missing")
	}

	idToken, err := o.provider.VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		log.Printf("failed to verify id token: %v", err)
		return nil, fmt.Errorf("failed to verify id token %w", err)
	}

	var claims struct {
		Nonce         string `json:"nonce"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		FirstName     string `json:"given_name"`
		LastName      string `json:"family_name"`
		Subject       string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		log.Printf("failed to parse id token claims: %v", err)
		return nil, fmt.Errorf("invalid id token claims")
	}

	if claims.Nonce != session.Nonce {
		return nil, fmt.Errorf("invalid nonce in id token claims")
	}

	if !claims.EmailVerified {
		return nil, fmt.Errorf("email should be verified")
	}

	return &models.UserAuthInfo{
		UserOIDCInfo: models.UserOIDCInfo{
			ProviderUserID: claims.Subject,
			Provider:       o.provider.GetProviderType(),
		},
		UserPersonalInfo: models.UserPersonalInfo{
			FirstName: claims.FirstName,
			LastName:  claims.LastName,
			Email:     claims.Email,
		},
	}, nil
}
