package adapter_test

import (
	"context"
	"testing"

	"github.com/shoet/webpagesummary/pkg/infrastracture/adapter"
	"github.com/shoet/webpagesummary/pkg/testutil"
)

func Test_CognitoService_Login(t *testing.T) {
	ctx := context.Background()

	cfg, err := testutil.LoadCognitoConfigForTest(t)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	service, err := adapter.NewCognitoService(ctx, cfg.CognitoClientID, cfg.CognitoUserPoolID)
	if err != nil {
		t.Fatalf("failed to create cognito service: %v", err)
	}

	args := struct {
		email    string
		password string
	}{
		email:    "test@example.com",
		password: "123456",
	}

	session, err := service.Login(ctx, args.email, args.password)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	_ = session
}
