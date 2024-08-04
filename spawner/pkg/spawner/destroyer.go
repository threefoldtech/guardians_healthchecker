package spawner

import (
	"context"
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

func DestroyVms(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) error {

	for _, farm := range cfg.Farms {
		nodes, err := getNodes(ctx, tfPluginClient, farm)
		if err != nil {
			return err
		}

		for _, node := range nodes {
			vmName := fmt.Sprintf("vm/%s", node.ID)
			networkName := fmt.Sprintf("net/%s", node.ID)

			err = tfPluginClient.CancelByProjectName(vmName, true)
			if err != nil {
				return err
			}

			err = tfPluginClient.CancelByProjectName(networkName, true)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
