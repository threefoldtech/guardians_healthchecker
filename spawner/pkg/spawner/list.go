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
		vms  []vmInfo
		lock sync.Mutex
		wg   sync.WaitGroup
	)

	for _, farm := range cfg.Farms {
		wg.Add(1)
		go func(farm uint64) {
			defer wg.Done()
			processFarm(ctx, farm, &vms, &lock, tfPluginClient)
		}(farm)
	}

	wg.Wait()

	displayVMs(vms)

	return nil
}

// processFarm processes all contracts for a given farm and appends VMs to the vms slice.
func processFarm(ctx context.Context, farm uint64, vms *[]vmInfo, lock *sync.Mutex, tfPluginClient deployer.TFPluginClient) {
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

	var farmGroup errgroup.Group
	for _, contract := range contracts.NodeContracts {
		contract := contract
		farmGroup.Go(func() error {
			return processContract(ctx, contract, farm, name, vms, lock, tfPluginClient)
		})
	}

	if err := farmGroup.Wait(); err != nil {
		fmt.Printf("Error processing contracts for farm %d: %v\n", farm, err)
	}
}

// processContract processes a single contract and appends the VM info to the vms slice.
func processContract(
	ctx context.Context,
	contract graphql.Contract,
	farm uint64,
	name string,
	vms *[]vmInfo,
	lock *sync.Mutex,
	tfPluginClient deployer.TFPluginClient,
) error {
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
		*vms = append(*vms, vmInfo{
			Farm:        farm,
			Node:        nodeID,
			Name:        metadata.Name,
			Contract:    contractID,
			ProjectName: name,
		})
		lock.Unlock()
	}

	return nil
}

// displayVMs prints the list of VMs in a tabular format.
func displayVMs(vms []vmInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Farm\tNode\tName\tContract\tProjectName")
	for _, vm := range vms {
		fmt.Fprintf(w, "%d\t%d\t%s\t%d\t%s\n", vm.Farm, vm.Node, vm.Name, vm.Contract, vm.ProjectName)
	}
	w.Flush()
}
