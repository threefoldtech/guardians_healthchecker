package cmd

import (
	"github.com/rs/zerolog/log"
	spawner "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

func setup(conf spawner.Config, debug bool) (deployer.TFPluginClient, error) {
	mnemonic := conf.Mnemonic
	log.Debug().Str("mnemonic", mnemonic).Send()

	opts := []deployer.PluginOpt{
		deployer.WithTwinCache(),
		deployer.WithRMBTimeout(30),
		deployer.WithProxyURL(conf.GridEndpoints.Proxy),
		deployer.WithRelayURL(conf.GridEndpoints.Relay),
		deployer.WithSubstrateURL(conf.GridEndpoints.SubsrateURL),
	}
	if debug {
		opts = append(opts, deployer.WithLogs())
	}

	return deployer.NewTFPluginClient(mnemonic, opts...)
}
