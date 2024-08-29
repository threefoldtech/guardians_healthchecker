package parser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cosmos/go-bip39"
	types "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
)

// validateMnemonic validates that mnemonic is valid
func validateMnemonic(mnemonic string) error {
	if !bip39.IsMnemonicValid(mnemonic) {
		return fmt.Errorf("invalid mnemonic: '%s'", mnemonic)
	}
	return nil
}

// validateFarms ensures all farm IDs are positive integers
func validateFarms(farms []uint64) error {
	for _, farm := range farms {
		if farm <= 0 {
			return fmt.Errorf("invalid farm ID: %d, must be a positive integer", farm)
		}
	}
	return nil
}

// validateDeploymentStrategy checks if the deployment strategy is a float between 0 and 1
func validateDeploymentStrategy(strategy float64) error {
	if strategy < 0 || strategy > 1 {
		return fmt.Errorf("invalid deployment strategy: %f, must be between 0 and 1", strategy)
	}
	return nil
}

// validateGridEndpoints checks if all grid endpoint URLs are valid
func validateGridEndpoints(endpoints types.Endpoints) error {
	if !isValidURL(endpoints.GraphQl, false) {
		return fmt.Errorf("invalid GraphQl endpoint URL: %s", endpoints.GraphQl)
	}
	if !isValidURL(endpoints.Proxy, false) {
		return fmt.Errorf("invalid Proxy endpoint URL: %s", endpoints.Proxy)
	}
	if !isValidURL(endpoints.Relay, true) {
		return fmt.Errorf("invalid Relay endpoint URL: %s", endpoints.Relay)
	}
	if !isValidURL(endpoints.SubstrateURL, true) {
		return fmt.Errorf("invalid Substrate URL: %s", endpoints.SubstrateURL)
	}

	return nil
}

// validateFailureStrategy ensures the failure strategy is one of the allowed values
func validateFailureStrategy(strategy string) error {
	validStrategies := map[string]bool{
		"retry":           true,
		"stop":            true,
		"destroy-all":     true,
		"destroy-failing": true,
	}

	if !validStrategies[strategy] {
		return fmt.Errorf("invalid failure strategy: %s, must be one of %v", strategy, validStrategies)
	}
	return nil
}

// validateInfluxConfig validates InfluxDB configuration fields
func validateInfluxConfig(config types.InfluxConfig) error {
	if !isValidURL(config.URL, false) {
		return fmt.Errorf("invalid influx URL: %s", config.URL)
	}
	if strings.TrimSpace(config.Org) == "" {
		return fmt.Errorf("influx organization cannot be empty")
	}
	if strings.TrimSpace(config.Token) == "" {
		return fmt.Errorf("influx token cannot be empty")
	}
	if strings.TrimSpace(config.Bucket) == "" {
		return fmt.Errorf("influx bucket cannot be empty")
	}
	return nil
}

// ValidateConfig performs all validations on the provided configuration
func ValidateConfig(cfg types.Config) error {
	if err := validateMnemonic(cfg.Mnemonic); err != nil {
		return err
	}
	if err := validateFarms(cfg.Farms); err != nil {
		return err
	}
	if err := validateDeploymentStrategy(cfg.DeploymentStrategy); err != nil {
		return err
	}
	if err := validateGridEndpoints(cfg.GridEndpoints); err != nil {
		return err
	}
	if err := validateFailureStrategy(cfg.FailureStrategy); err != nil {
		return err
	}

	if err := validateInfluxConfig(cfg.Influx); err != nil {
		return err
	}
	return nil
}

// isValidURL checks if the URL is valid and optionally ensures it uses the WebSocket protocol
func isValidURL(u string, requireWebSocket bool) bool {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}

	if requireWebSocket {
		if parsedURL.Scheme != "ws" && parsedURL.Scheme != "wss" {
			return false
		}
	}

	return true
}
