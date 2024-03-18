package adapter

import (
	"context"
	"fmt"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/shoet/webpagesummary/pkg/infrastracture/entities"
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

func (c *CognitoService) Login(ctx context.Context, email, password string) (*entities.LoginSession, error) {
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

	session := &entities.LoginSession{
		IdToken:      *res.AuthenticationResult.IdToken,
		AccessToken:  *res.AuthenticationResult.AccessToken,
		RefreshToken: *res.AuthenticationResult.RefreshToken,
	}

	return session, nil
}

func (c *CognitoService) GetUserInfo(ctx context.Context, accessToken string) (*entities.User, error) {
	output, err := c.client.GetUser(ctx, &cognitoidentityprovider.GetUserInput{
		AccessToken: &accessToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	fmt.Println(output)
	return &entities.User{
		Email:    "test",
		Username: "test",
	}, nil
}

func (c *CognitoService) VeryfyToken(ctx context.Context, accessToken string) error {
	if _, err := c.GetUserInfo(ctx, accessToken); err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}
	return nil
}
