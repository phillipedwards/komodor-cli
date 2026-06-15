package auth

import (
	"errors"
	"os"

	"github.com/phillipedwards/komodor-cli/internal/config"
)

const envKey = "KOMODOR_API_KEY"

var ErrNoAPIKey = errors.New("no API key configured — set KOMODOR_API_KEY or run 'komodor auth set-key <key>'")

// Resolve returns an API key in priority order: flag > env var > config file.
func Resolve(flagKey string, cfg *config.Config) (string, error) {
	if flagKey != "" {
		return flagKey, nil
	}
	if v := os.Getenv(envKey); v != "" {
		return v, nil
	}
	if cfg != nil && cfg.APIKey != "" {
		return cfg.APIKey, nil
	}
	return "", ErrNoAPIKey
}
