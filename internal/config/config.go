package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	FireflyURL   string
	FireflyToken string
	Secret       []byte
	Host         string
	Port         string
	LogLevel     slog.Level
}

func BuildConfigFromEnv() (Config, error) {
	c := Config{}
	c.FireflyURL = os.Getenv("FIREFLY_URL")
	if c.FireflyURL == "" {
		return Config{}, fmt.Errorf("FIREFLY_URL should be present")
	}

	c.FireflyToken = os.Getenv("FIREFLY_TOKEN")
	if c.FireflyToken == "" {
		return Config{}, fmt.Errorf("FIREFLY_TOKEN should be present")
	}

	secretStr := os.Getenv("WEBHOOK_SECRET")
	if secretStr == "" {
		return Config{}, fmt.Errorf("WEBHOOK_SECRET should be present")
	}
	c.Secret = []byte(secretStr)

	c.Host = os.Getenv("FIREFLY_BOC_FIXER_HOST")
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}

	c.LogLevel = parseLogLevel(os.Getenv("LOG_LEVEL"))

	c.Port = os.Getenv("FIREFLY_BOC_FIXER_PORT")
	if c.Port == "" {
		c.Port = "3000"
	}
	return c, nil
}

func parseLogLevel(str string) slog.Level {
	str = strings.ToUpper(str)
	switch str {
	case "ERROR":
		return slog.LevelError
	case "WARN":
		return slog.LevelWarn
	case "INFO":
		return slog.LevelInfo
	case "DEBUG":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}
