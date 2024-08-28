package spawner

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

// Destroy destroys VMs
func Destroy(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) error {
	names := []string{}

	for _, farm := range cfg.Farms {
		names = append(names, fmt.Sprintf("vm/%d", farm))
	}

	return destroy(tfPluginClient, names)
}

// destroy destroys VMs by names
func destroy(tfPluginClient deployer.TFPluginClient, names []string) error {
	var resultErr *multierror.Error

	for _, name := range names {
		err := tfPluginClient.CancelByProjectName(name, true)
		if err != nil {
			resultErr = multierror.Append(resultErr, err)
		}
	}

	return resultErr.ErrorOrNil()
}

// destroyFailingNetworks destroys failing networks
func destroyFailingNetworks(ctx context.Context, tfPluginClient deployer.TFPluginClient, failingNetworks []*workloads.ZNet) error {
	var failing []*workloads.ZNet

	for _, network := range failingNetworks {
		if len(network.NodeDeploymentID) != 0 {
			failing = append(failing, network)
		}
	}
	err := tfPluginClient.NetworkDeployer.BatchCancel(ctx, failing)
	if err != nil {
		return err
	}

	return nil
}
