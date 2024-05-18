//go:build e2e_tests
// +build e2e_tests

package sanity

import (
	"fmt"
	"net/url"
	"os"
)

type config struct {
	ServiceAddr string
}

func getConfig() (config, error) {
	cfg := config{}

	const defaultServiceAddr = "http://localhost:8080"
	serviceAddr, ok := os.LookupEnv("SHORTIK_HOST")
	if !ok {
		serviceAddr = defaultServiceAddr
	}

	var err error
	cfg.ServiceAddr, err = url.JoinPath(serviceAddr, "v1")
	if err != nil {
		return cfg, fmt.Errorf("failed to get the service address base URL: %w", err)
	}

	return cfg, nil
}
