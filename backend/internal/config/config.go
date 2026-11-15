package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string
}

func Load() (*Config, error) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := &Config{
		HTTPPort: os.Getenv("HTTP_PORT"),
	}
	if cfg.HTTPPort == "" {
		return nil, fmt.Errorf("HTTP_PORT env variable not set")
	}
	return cfg, nil
}
