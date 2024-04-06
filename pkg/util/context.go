package util

import (
	"context"
	"fmt"
)

type TokenSubContextKey struct{}
type HasAPIKeyContextKey struct{}

const APIKeyUserSub = "apikey"

func GetUserSub(ctx context.Context) (string, error) {
	_, ok := ctx.Value(HasAPIKeyContextKey{}).(bool)
	if ok {
		return APIKeyUserSub, nil
	}

	userSub, ok := ctx.Value(TokenSubContextKey{}).(string)
	if !ok {
		return "", fmt.Errorf("failed to get user sub from context")
	}

	return userSub, nil
}
