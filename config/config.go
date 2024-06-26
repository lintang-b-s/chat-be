package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App      `yaml:"app"`
		HTTP     `yaml:"http"`
		Log      `yaml:"logger"`
		Redis    `yaml:"redis"`
		Postgres `yaml:"postgres"`
		EdenAi   `yaml:"edenAi"`
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
	}

	Redis struct {
		Address  string `env-required:"true" yaml:"server_address" env:"REDIS_ADDRESS" `
		Password string `env-required:"true" yaml:"password" env:"REDIS_PASSWORD"`
	}

	Postgres struct {
		Host string `env-required:"true"  env:"POSTGRES_HOST"`
		Username string `env-required:"true" yaml:"username" env:"POSTGRES_USERNAME"`
		Password string `env-required:"true" yaml:"password" env:"POSTGRES_PASSWORD"`
	}

	EdenAi struct {
		ApiKey string `env-required:"true" env:"EDENAI_APIKEY" env-default:"EDENAI_APIKEY"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}


	err := cleanenv.ReadConfig("./.env", cfg)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
