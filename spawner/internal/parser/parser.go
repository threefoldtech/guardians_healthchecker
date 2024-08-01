package parser

import (
	"fmt"
	"io"

	spawner "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
	"gopkg.in/yaml.v3"
)

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
	if err := validateMnemonic(conf.Mnemonic); err != nil {
		return spawner.Config{}, err
	}

	return conf, nil
}
