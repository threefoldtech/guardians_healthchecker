package cmd

import (
	spawner "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

func setup(cfg spawner.Config) (deployer.TFPluginClient, error) {
	mnemonic := cfg.Mnemonic

	opts := []deployer.PluginOpt{
		deployer.WithNetwork("dev"),
		deployer.WithTwinCache(),
		deployer.WithRMBTimeout(30),
		deployer.WithProxyURL(cfg.GridEndpoints.Proxy),
		deployer.WithRelayURL(cfg.GridEndpoints.Relay),
		deployer.WithSubstrateURL(cfg.GridEndpoints.SubsrateURL),
		// deployer.WithGraphQlURL(cfg.GridEndpoints.GraphQl),
	}

	return deployer.NewTFPluginClient(mnemonic, opts...)
}
