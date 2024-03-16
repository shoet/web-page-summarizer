package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/shoet/webpagesummary/pkg/config"
)

func LoadCognitoConfigForTest(t *testing.T) (*config.CognitoConfig, error) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	projectRootDir, err := GetProjectRootDir(cwd, 10)
	if err != nil {
		t.Fatalf("failed to get project root dir: %v", err)
	}

	dotEnvPath := filepath.Join(projectRootDir, ".env")
	if err := godotenv.Load(dotEnvPath); err != nil {
		t.Fatalf("failed to load .env: %v", err)
	}
	cfg, err := config.NewCognitoConfig()
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	return cfg, nil
}
