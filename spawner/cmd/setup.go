package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/guardians_healthchecker/spawner/internal/parser"
	spawner "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

// setup sets up a new TFPluginClient
func setup(cfg spawner.Config) (deployer.TFPluginClient, error) {
	mnemonic := cfg.Mnemonic

	opts := []deployer.PluginOpt{
		deployer.WithTwinCache(),
		deployer.WithRMBTimeout(30),
		deployer.WithProxyURL(cfg.GridEndpoints.Proxy),
		deployer.WithRelayURL(cfg.GridEndpoints.Relay),
		deployer.WithSubstrateURL(cfg.GridEndpoints.SubstrateURL),
		deployer.WithGraphQlURL(cfg.GridEndpoints.GraphQl),
	}

	return deployer.NewTFPluginClient(mnemonic, opts...)
}

// loadConfigAndSetup loads and parses the configuration file, sets up the tfPluginClient, and returns the context, config, and client.
func loadConfigAndSetup(cmd *cobra.Command) (context.Context, spawner.Config, deployer.TFPluginClient, error) {
	if len(cmd.Flags().Args()) != 0 {
		return nil, spawner.Config{}, deployer.TFPluginClient{}, fmt.Errorf("command '%s' does not support additional arguments: %v", cmd.Name(), cmd.Flags().Args())
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, spawner.Config{}, deployer.TFPluginClient{}, fmt.Errorf("error in configuration file path: %w", err)
	}

	if configPath == "" {
		return nil, spawner.Config{}, deployer.TFPluginClient{}, fmt.Errorf("required configuration file path is empty")
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, spawner.Config{}, deployer.TFPluginClient{}, fmt.Errorf("failed to open configuration file '%s' with error: %w", configPath, err)
	}
	defer configFile.Close()

	yamlFmt := filepath.Ext(configPath) == ".yaml"
	if !yamlFmt {
		return nil, spawner.Config{}, deployer.TFPluginClient{}, fmt.Errorf("unsupported configuration file format '%s', should be .yaml", configPath)
	}

	cfg, err := parser.ParseConfig(configFile)
	if err != nil {
		return nil, spawner.Config{}, deployer.TFPluginClient{}, fmt.Errorf("failed to parse configuration file '%s' with error: %w", configPath, err)
	}

	tfPluginClient, err := setup(cfg)
	if err != nil {
		return nil, spawner.Config{}, deployer.TFPluginClient{}, err
	}

	ctx := context.Background()

	return ctx, cfg, tfPluginClient, nil
}
