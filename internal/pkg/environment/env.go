package environment

import (
	"fmt"

	"github.com/ferdiebergado/gopherkit/env"
)

func LoadEnv(appEnv string) error {
	const (
		envDev  = ".env"
		envTest = ".env.testing"
	)
	var envFile string

	switch appEnv {
	case "development":
		envFile = envDev
	case "testing":
		envFile = envTest
	default:
		return fmt.Errorf("unrecognized environment: %s", appEnv)
	}

	if err := env.Load(envFile); err != nil {
		return fmt.Errorf("cannot load env file %s, %w", envFile, err)
	}

	return nil
}

func Setup() (string, error) {
	appEnv := env.Get("ENV", "development")
	if appEnv != "production" {
		if err := LoadEnv(appEnv); err != nil {
			return "", fmt.Errorf("load env: %w", err)
		}
	}
	return appEnv, nil
}
