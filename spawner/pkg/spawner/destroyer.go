package spawner

import (
	"context"
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)


// DestroyVms destroys VMs on specified farms in config file
func DestroyVms(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) error {

	for _, farm := range cfg.Farms {
		name := fmt.Sprintf("vm/%d", farm)

		err := tfPluginClient.CancelByProjectName(name, true)
		if err != nil {
			return err
		}
	}

	return nil
}
