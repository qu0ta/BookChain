package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
}

func Load() (*Config, error) {
	if err := godotenv.Load("../.env.example"); err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := &Config{
		HTTPPort:    os.Getenv("HTTP_PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if strings.TrimSpace(cfg.HTTPPort) == "" {
		return nil, fmt.Errorf("HTTP_PORT env variable not set")
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, fmt.Errorf("DATABASE_URL env variable not set")
	}

	return cfg, nil
}
