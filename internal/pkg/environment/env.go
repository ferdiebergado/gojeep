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
