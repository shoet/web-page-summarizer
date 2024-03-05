package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	QueueUrl            string `env:"QUEUE_URL,required"`
	BrowserPath         string `env:"BROWSER_PATH"`
	OpenAIApiKey        string `env:"OPENAI_API_KEY"`
	ExecTimeout         int    `env:"EXEC_TIMEOUT_SEC" envDefault:"300"`
	BrowserDownloadPath string `env:"BROWSER_DOWNLOAD_PATH" envDefault:"/tmp/playwright/browser"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed Parse config: %w", err)
	}
	return cfg, nil
}
