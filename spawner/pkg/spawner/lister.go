package spawner

import (
	"context"
	"fmt"
	"time"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

type vmInfo struct {
	Name        string
	RunningTime time.Duration
}

func ListVMs(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) ([]vmInfo, error) {
	var vms []vmInfo

	for _, farm := range cfg.Farms {
		name := fmt.Sprintf("vm/%s", farm)
		contracts, err := tfPluginClient.ContractsGetter.ListContractsOfProjectName(name, true)
		if err != nil {
			return nil, err
		}

		nodeContracts := contracts.NodeContracts
		for _, contract := range nodeContracts {
			vmname := contract.DeploymentData //still searching for a way to extract the name and time from this string
			runningtime := time.Since(contract.DeploymentData)

			vms = append(vms, vmInfo{
				Name:        vmname,
				RunningTime: runningtime,
			})
		}
	}

	return vms, nil
}
