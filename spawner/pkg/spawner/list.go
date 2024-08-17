package spawner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"text/tabwriter"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"golang.org/x/sync/errgroup"
)

// ListVMs lists running VMs on specified farms in the config file.
func List(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) error {
	var vms []vmInfo
	var lock sync.Mutex
	var wg sync.WaitGroup

	for _, farm := range cfg.Farms {
		wg.Add(1)
		go func(farm uint64) {
			defer wg.Done()

			name := fmt.Sprintf("vm/%d", farm)
			contracts, err := tfPluginClient.ContractsGetter.ListContractsOfProjectName(name, true)
			if err != nil {
				fmt.Printf("error listing contracts for farm %d: %v\n", farm, err)
				return
			}
			if len(contracts.NodeContracts) == 0 {
				log.Warn().Msgf("no VMs found for farm %d with project name %s", farm, name)
				return
			}

			var farmWg sync.WaitGroup
			var farmGroup errgroup.Group

			for _, contract := range contracts.NodeContracts {
				farmWg.Add(1)
				contract := contract

				farmGroup.Go(func() error {
					defer farmWg.Done()

					contractID, err := strconv.ParseUint(contract.ContractID, 10, 64)
					if err != nil {
						return err
					}

					nodeID := contract.NodeID

					nodeClient, err := tfPluginClient.State.NcPool.GetNodeClient(tfPluginClient.State.Substrate, nodeID)
					if err != nil {
						return err
					}

					dl, err := nodeClient.DeploymentGet(ctx, contractID)
					if err != nil {
						return err
					}

					var metadata deploymentMetadata
					err = json.Unmarshal([]byte(dl.Metadata), &metadata)
					if err != nil {
						return err
					}

					if metadata.Type == "vm" {
						lock.Lock()
						vms = append(vms, vmInfo{
							Farm:        farm,
							Node:        nodeID,
							Name:        metadata.Name,
							Contract:    contractID,
							ProjectName: name,
						})
						lock.Unlock()
					}

					return nil
				})
			}

			if err := farmGroup.Wait(); err != nil {
				fmt.Printf("Error processing contracts for farm %d: %v\n", farm, err)
			}

		}(farm)
	}

	wg.Wait()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Farm\tNode\tName\tContract\tProjectName")
	for _, vm := range vms {
		fmt.Fprintf(w, "%d\t%d\t%s\t%d\t%s\n", vm.Farm, vm.Node, vm.Name, vm.Contract, vm.ProjectName)
	}

	w.Flush()

	return nil
}
