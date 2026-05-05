package oidc

import (
	"context"
	"fmt"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/config"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCProvider interface {
	GetProviderType() string
	ExchangeAuthCodeForAccessToken(ctx context.Context, authCode string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	WithNonce(nonce string) OIDCProvider
	WithState(state string) OIDCProvider
	VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
	BuildAuthCodeURL() string
}

type oidcProvider struct {
	providerType string
	provider     *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
	state        string
	nonce        string
}

func NewOIDCProvider(ctx context.Context, cfg config.OIDCProviderConfig) (*oidcProvider, error) {
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

	return &oidcProvider{
		providerType: cfg.ProviderType,
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
	}, nil
}

func (p *oidcProvider) WithNonce(nonce string) OIDCProvider {
	p.nonce = nonce
	return p
}

func (p *oidcProvider) WithState(state string) OIDCProvider {
	p.state = state
	return p
}

func (p *oidcProvider) ExchangeAuthCodeForAccessToken(ctx context.Context, authCode string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return p.oauth2Config.Exchange(ctx, authCode, opts...)
}

func (p *oidcProvider) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return p.verifier.Verify(ctx, rawIDToken)
}

func (p *oidcProvider) BuildAuthCodeURL() string {
	return p.oauth2Config.AuthCodeURL(p.state, oauth2.SetAuthURLParam("nonce", p.nonce))
}

func (p *oidcProvider) GetProviderType() string {
	return p.providerType
}
