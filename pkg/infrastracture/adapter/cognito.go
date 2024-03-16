package adapter

import (
	"context"
	"fmt"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type CognitoService struct {
	client     *cognitoidentityprovider.Client
	clientID   string
	userPoolID string
}

func NewCognitoService(ctx context.Context, clientId string, userPoolId string) (*CognitoService, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := cognitoidentityprovider.NewFromConfig(cfg)
	return &CognitoService{
		client:     client,
		clientID:   clientId,
		userPoolID: userPoolId,
	}, nil
}

type LoginSession struct {
	IdToken      string `json:"idToken"`
	AccessToken  string `jsonn:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (c *CognitoService) Login(ctx context.Context, email, password string) (*LoginSession, error) {
	res, err := c.client.AdminInitiateAuth(ctx, &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow:   types.AuthFlowTypeAdminUserPasswordAuth,
		ClientId:   &c.clientID,
		UserPoolId: &c.userPoolID,
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initiate auth: %w", err)
	}

	if res.AuthenticationResult == nil {
		return nil, fmt.Errorf("authentication result is nil: %w", err)
	}

	session := &LoginSession{
		IdToken:      *res.AuthenticationResult.IdToken,
		AccessToken:  *res.AuthenticationResult.AccessToken,
		RefreshToken: *res.AuthenticationResult.RefreshToken,
	}

	return session, nil
}
