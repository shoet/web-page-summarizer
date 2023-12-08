package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	QueueUrl       string `env:"QUEUE_URL,required"`
	BrowserPath    string `env:"BROWSER_PATH"`
	OpenAIApiKey   string `env:"OPENAI_API_KEY"`
	MaxTaskExecute int    `env:"MAX_TASK_EXECUTE" envDefault:"20"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed Parse config: %w", err)
	}
	return cfg, nil
}
