package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

type generator struct {
	signingMethod jwt.SigningMethod
}

func NewGenerator(signingMethod jwt.SigningMethod) *generator {
	return &generator{
		signingMethod: signingMethod,
	}
}

func (g *generator) GenerateSignedJWT(claims jwt.Claims, secret []byte) (string, error) {
	token := jwt.NewWithClaims(g.signingMethod, claims)

	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return signedToken, nil
}
