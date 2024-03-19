package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/shoet/webpagesummary/pkg/config"
)

func LoadCognitoConfigForTest(t *testing.T) (*config.CognitoConfig, error) {
	t.Helper()
	return LoadCognitoConfigLocal()
}

func LoadCognitoConfigLocal() (*config.CognitoConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}

	projectRootDir, err := GetProjectRootDir(cwd, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get project root directory: %v", err)
	}

	dotEnvPath := filepath.Join(projectRootDir, ".env")
	if err := godotenv.Load(dotEnvPath); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}
	cfg, err := config.NewCognitoConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load cognito config: %v", err)
	}
	return cfg, nil
}
