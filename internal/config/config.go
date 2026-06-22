package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port            int           `env:"PORT" env-default:"8080"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`
	BaseURL         string        `env:"BASE_URL" env-default:"http://localhost:8080"`
	Storage         string        `env:"STORAGE" env-default:"memory"`
	ShortCode       ShortCode
	Database        Database
}

type ShortCode struct {
	Length     int    `env:"SHORTCODE_LENGTH" env-default:"10"`
	Alphabet   string `env:"SHORTCODE_ALPHABET" env-default:"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_"`
	MaxRetries int    `env:"SHORTCODE_MAX_RETRIES" env-default:"5"`
}

type Database struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
	Database string `env:"POSTGRES_DATABASE" env-default:"postgres"`
}

func Load() (*Config, error) {
	cfg := Config{}

	if _, err := os.Stat(".env"); err == nil {
		if err = cleanenv.ReadConfig(".env", &cfg); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		return &cfg, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to access dotenv file: %w", err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}
