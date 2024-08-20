package parser

import (
	"strings"
	"testing"

	// "github.com/stretchr/testify/assert"
	types "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
	"gopkg.in/yaml.v3"
	"gotest.tools/assert"
)

func TestParseConfig(t *testing.T) {
	confStruct := types.Config{
		Farms:              []uint64{1, 2, 3},
		DeploymentStrategy: 1.0,
		GridEndpoints: types.Endpoints{
			GraphQl:      "https://graphql.dev.grid.tf/graphql",
			Proxy:        "https://gridproxy.dev.grid.tf/",
			Relay:        "wss://relay.dev.grid.tf",
			SubstrateURL: "wss://tfchain.dev.grid.tf/ws",
		},
		Mnemonic:        "rival oyster defense garbage fame disease mask mail family wire village vibrant index fuel dolphin",
		FailureStrategy: "retry",
		SSHKey:          "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSUGPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XAt3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/EnmZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbxNrRFi9wrf+M7Q== schacon@mylaptop.local",
		Influx: types.InfluxConfig{
			URL:    "http://influx.example.com",
			Org:    "example_org",
			Token:  "example_token",
			Bucket: "example_bucket",
		},
	}
	t.Run("valid config", func(t *testing.T) {
		conf := confStruct

		data, err := yaml.Marshal(conf)
		assert.NilError(t, err)
		println(string(data))

		configFile := strings.NewReader(string(data))

		config, err := ParseConfig(configFile)
		assert.NilError(t, err)
		assert.DeepEqual(t, conf, config)
	})
	t.Run("invalid yaml format", func(t *testing.T) {
		configFile := strings.NewReader("invalid yaml format")

		_, err := ParseConfig(configFile)
		assert.Error(t, err, err.Error())
	})
	t.Run("invalid mnemonic", func(t *testing.T) {
		conf := confStruct
		conf.Mnemonic = "invalid mnemonic"

		data, err := yaml.Marshal(conf)
		assert.NilError(t, err)

		configFile := strings.NewReader(string(data))

		_, err = ParseConfig(configFile)
		assert.Error(t, err, err.Error())
	})
	t.Run("invalid deployment strategy", func(t *testing.T) {
		conf := confStruct
		conf.DeploymentStrategy = 2.0

		data, err := yaml.Marshal(conf)
		assert.NilError(t, err)

		configFile := strings.NewReader(string(data))

		_, err = ParseConfig(configFile)
		assert.Error(t, err, err.Error())
	})
	t.Run("invalid grid endpoints", func(t *testing.T) {
		conf := confStruct
		conf.GridEndpoints.GraphQl = "invalid url"

		data, err := yaml.Marshal(conf)
		assert.NilError(t, err)

		configFile := strings.NewReader(string(data))

		_, err = ParseConfig(configFile)
		assert.Error(t, err, err.Error())
	})
	t.Run("invalid influx config", func(t *testing.T) {
		conf := confStruct
		conf.Influx.URL = "invalid url"

		data, err := yaml.Marshal(conf)
		assert.NilError(t, err)

		configFile := strings.NewReader(string(data))

		_, err = ParseConfig(configFile)
		assert.Error(t, err, err.Error())
	})
}
