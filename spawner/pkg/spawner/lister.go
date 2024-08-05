package spawner

import (
	"context"
	"fmt"
	"strconv"

	// "time"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

type vmInfo struct {
	Name string
	// RunningTime time.Duration
}

func ListVMs(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) ([]vmInfo, error) {
	var deployments []workloads.Deployment
	var vms []vmInfo

	for _, farm := range cfg.Farms {
		name := fmt.Sprintf("vm/%d", farm)
		contracts, err := tfPluginClient.ContractsGetter.ListContractsOfProjectName(name, true)
		if err != nil {
			return nil, err
		}
		if len(contracts.NodeContracts) == 0 {
			return nil, fmt.Errorf("couldn't find any contracts with name %s", name)
		}

		var nodeID uint32

		for _, contract := range contracts.NodeContracts {
			contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
			if err != nil {
				return nil, err
			}
			nodeID = contract.NodeID
			checkIfExistAndAppend(tfPluginClient, nodeID, contractID)
			deployment, err := tfPluginClient.State.LoadDeploymentFromGrid(ctx, nodeID, name)
			if err != nil {
				return nil, err
			}
			deployments = append(deployments, deployment)
		}

		for _, deployment := range deployments {
			vms = append(vms, vmInfo{
				Name: deployment.Name,
				// RunningTime: deployment.,
			})

		}
	}

	return vms, nil
}

func checkIfExistAndAppend(t deployer.TFPluginClient, node uint32, contractID uint64) {
	for _, n := range t.State.CurrentNodeDeployments[node] {
		if n == contractID {
			return
		}
	}

	t.State.CurrentNodeDeployments[node] = append(t.State.CurrentNodeDeployments[node], contractID)
}
