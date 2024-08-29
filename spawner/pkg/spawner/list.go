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
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	"golang.org/x/sync/errgroup"
)

// List lists running VMs on specified farms in the config file.
func List(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) error {
	var (
		vms []vmInfo
		wg  sync.WaitGroup
		mu  sync.Mutex
	)

	for _, farm := range cfg.Farms {
		wg.Add(1)
		go func(farm uint64) {
			defer wg.Done()
			farmVMs := processFarm(ctx, farm, tfPluginClient)
			mu.Lock()
			vms = append(vms, farmVMs...)
			mu.Unlock()
		}(farm)
	}

	wg.Wait()

	displayVMs(vms)

	return nil
}

// processFarm processes all contracts for a given farm and returns a slice of VMs.
func processFarm(ctx context.Context, farm uint64, tfPluginClient deployer.TFPluginClient) []vmInfo {
	name := fmt.Sprintf("vm/%d", farm)
	contracts, err := tfPluginClient.ContractsGetter.ListContractsOfProjectName(name, true)
	if err != nil {
		fmt.Printf("error listing contracts for farm %d: %v\n", farm, err)
		return nil
	}
	if len(contracts.NodeContracts) == 0 {
		log.Warn().Msgf("no VMs found for farm %d with project name %s", farm, name)
		return nil
	}

	var (
		farmGroup errgroup.Group
		vms       []vmInfo
		mu        sync.Mutex
	)

	for _, contract := range contracts.NodeContracts {
		contract := contract
		farmGroup.Go(func() error {
			vm, err := processContract(ctx, contract, farm, name, tfPluginClient)
			if err != nil {
				return err
			}
			if vm != nil {
				mu.Lock()
				vms = append(vms, *vm)
				mu.Unlock()
			}
			return nil
		})
	}

	if err := farmGroup.Wait(); err != nil {
		fmt.Printf("Error processing contracts for farm %d: %v\n", farm, err)
	}

	return vms
}

// processContract processes a single contract and returns the VM info.
func processContract(
	ctx context.Context,
	contract graphql.Contract,
	farm uint64,
	name string,
	tfPluginClient deployer.TFPluginClient,
) (*vmInfo, error) {
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
		return &vmInfo{
			Farm:        farm,
			Node:        nodeID,
			Name:        metadata.Name,
			Contract:    contractID,
			ProjectName: name,
		}, nil
	}

	return nil, nil
}

// displayVMs prints the list of VMs in a tabular format.
func displayVMs(vms []vmInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Farm\tNode\tName\tContract\tProjectName")
	for _, vm := range vms {
		fmt.Fprintf(w, "%d\t%d\t%s\t%d\t%s\n", vm.Farm, vm.Node, vm.Name, vm.Contract, vm.ProjectName)
	}
	w.Flush()
}
