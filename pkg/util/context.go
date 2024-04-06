package util

import (
	"context"
	"fmt"
)

type TokenSubContextKey struct{}
type HasAPIKeyContextKey struct{}

func GetUserSub(ctx context.Context) (string, error) {
	userSub, ok := ctx.Value(TokenSubContextKey{}).(string)
	if !ok {
		return "", fmt.Errorf("failed to get user sub from context")
	}

	hasAPIKey, _ := ctx.Value(HasAPIKeyContextKey{}).(bool)
	if !ok {
		return "", fmt.Errorf("failed to get user sub from context")
	}

	if hasAPIKey {
		userSub = "apikey"
	}
	return userSub, nil
}
