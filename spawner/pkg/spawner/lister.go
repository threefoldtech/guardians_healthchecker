package spawner

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

func ListVMs(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) ([]vmInfo, error) {
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

		for _, contract := range contracts.NodeContracts {
			contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
			if err != nil {
				return nil, err
			}

			nodeID := contract.NodeID

			nodeClient, err := tfPluginClient.State.NcPool.GetNodeClient(tfPluginClient.State.Substrate, nodeID)
			if err != nil {
				return nil, err
			}

			dl, err := nodeClient.DeploymentGet(ctx, contractID)
			if err != nil {
				return nil, err
			}

			var metadata deploymentMetadata
			err = json.Unmarshal([]byte(dl.Metadata), &metadata)
			if err != nil {
				return nil, err
			}

			if metadata.Type == "vm" {
				vms = append(vms, vmInfo{
					Farm:        farm,
					Node:        nodeID,
					Name:        metadata.Name,
					Contract:    contractID,
					ProjectName: metadata.ProjectName,
				})
			}
		}
	}
	return vms, nil
}
