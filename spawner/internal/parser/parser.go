package parser

import (
	"fmt"
	"io"

	spawner "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
	"gopkg.in/yaml.v3"
)

// ParseConfig parse the config file
func ParseConfig(file io.Reader) (spawner.Config, error) {
	conf := spawner.Config{}

	configFile, err := io.ReadAll(file)
	if err != nil {
		return spawner.Config{}, fmt.Errorf("failed to read the config file: %+w", err)
	}
	err = yaml.Unmarshal(configFile, &conf)
	if err != nil {
		return spawner.Config{}, err
	}
	if err := ValidateConfig(conf); err != nil {
		return spawner.Config{}, err
	}

	return conf, nil
}
