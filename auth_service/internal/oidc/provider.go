package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"
	"golang.org/x/oauth2"
)

type Provider struct {
	OIDCProvider *oidc.Provider
	OAuth2Config *oauth2.Config
	Verifier     *oidc.IDTokenVerifier
}

func NewProvider(ctx context.Context, cfg config.GoogleOIDC) (*Provider, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to create provider for %s: %w", cfg.IssuerURL, err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       cfg.Scopes,
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	return &Provider{
		OIDCProvider: provider,
		OAuth2Config: oauth2Config,
		Verifier:     verifier,
	}, nil
}
