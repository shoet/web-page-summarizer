package util

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Key struct {
	Alg string `json:"alg"`
	E   string `json:"e"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	Use string `json:"use"`
}

func VerifyToken(ctx context.Context, keyUrl string, accessToken string) (tokenSub string, err error) {
	set, err := jwk.Fetch(ctx, keyUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch JWK: %v", err)
	}
	token, err := jwt.Parse([]byte(accessToken), jwt.WithKeySet(set))
	if err != nil {
		return "", fmt.Errorf("failed to parse JWT: %v", err)
	}
	sub, ok := token.Get("sub")
	if !ok {
		return "", fmt.Errorf("sub claim not found")
	}
	tokenSub, ok = sub.(string)
	if !ok {
		return "", fmt.Errorf("sub claim is not a string")
	}
	return tokenSub, nil
}
