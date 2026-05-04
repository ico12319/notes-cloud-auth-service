package oidc

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"log"
)

type idTokenVerifier struct {
	provider *Provider
}

func NewIDTokenVerifier(provider *Provider) *idTokenVerifier {
	return &idTokenVerifier{
		provider: provider,
	}
}

func (i *idTokenVerifier) Verify(ctx context.Context, exchangedToken *oauth2.Token, session *OAuthSession) error {
	rawIDToken, ok := exchangedToken.Extra("id_token").(string)
	if !ok {
		return fmt.Errorf("id token is missing")
	}

	idToken, err := i.provider.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Printf("failed to verify id token: %v", err)

		return fmt.Errorf("failed to verify id token %w", err)
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

		return fmt.Errorf("invalid id token claims")
	}

	if claims.Nonce != session.Nonce {
		return fmt.Errorf("invalid nonce in id token claims")
	}

	if !claims.EmailVerified {
		return fmt.Errorf("email should be verified")
	}

	return nil
}
