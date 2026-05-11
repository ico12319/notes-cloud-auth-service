package oauth

import (
	"log"
)

type randomGenerator interface {
	GenerateRandomString(length int) (string, error)
}

type sessionBuilder struct {
	randomGenerator randomGenerator
}

func NewOauthSessionBuilder(randomGenerator randomGenerator) *sessionBuilder {
	return &sessionBuilder{
		randomGenerator: randomGenerator,
	}
}

func (s *sessionBuilder) Build() (*OAuthSession, error) {
	state, err := s.randomGenerator.GenerateRandomString(32)
	if err != nil {
		log.Printf("failed to generate state: %v", err)

		return nil, err
	}

	nonce, err := s.randomGenerator.GenerateRandomString(32)
	if err != nil {
		log.Printf("failed to generate nonce: %v", err)

		return nil, err
	}

	return &OAuthSession{
		State: state,
		Nonce: nonce,
	}, nil
}
