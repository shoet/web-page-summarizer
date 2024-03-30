package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Env                    string `env:"ENV"`
	QueueUrl               string `env:"QUEUE_URL,required"`
	BrowserPath            string `env:"BROWSER_PATH"`
	OpenAIApiKey           string `env:"OPENAI_API_KEY"`
	ExecTimeout            int    `env:"EXEC_TIMEOUT_SEC" envDefault:"300"`
	BrowserDownloadPath    string `env:"BROWSER_DOWNLOAD_PATH" envDefault:"/tmp/playwright/browser"`
	CORSWhiteList          string `env:"CORS_WHITE_LIST"`
	CognitoJWKUrl          string `env:"COGNITO_JWK_URL"`
	RequestRateLimitMax    int    `env:"REQUEST_RATE_LIMIT_MAX" envDefault:"10"`
	RequestRateLimitTTLSec int    `env:"REQUEST_RATE_LIMIT_TTL_SEC" envDefault:"86400"`
}

func (c *Config) GetCORSWhiteList() []string {
	whiteList := strings.Split(c.CORSWhiteList, ",")
	return whiteList
}

type RDBConfig struct {
	RDBDsn string `env:"RDB_DSN,required"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed Parse config: %w", err)
	}
	return cfg, nil
}

func NewRDBConfig() (*RDBConfig, error) {
	cfg := &RDBConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed Parse config: %w", err)
	}
	return cfg, nil
}

type CognitoConfig struct {
	CognitoUserPoolID string `env:"COGNITO_USER_POOL_ID,required"`
	CognitoClientID   string `env:"COGNITO_CLIENT_ID,required"`
	CognitoIDPoolID   string `env:"COGNITO_ID_POOL_ID,required"`
}

func NewCognitoConfig() (*CognitoConfig, error) {
	var config CognitoConfig
	if err := env.Parse(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &config, nil
}
