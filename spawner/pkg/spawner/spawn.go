package spawner

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-retry"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

// Represents the configuration for the deployment
const (
	gb         = 1024 * 1024 * 1024
	cpuCount   = 4
	memorySize = 8
	rootSize   = 40
)

// Represents the deployment strategy
const (
	defaultMaxRetries = 5
	retrying          = "retry"
	destroyAll        = "destroy-all"
	destroyFailing    = "destroy-failing"
	stop              = "stop"
)

// Spawn given a list of farm IDs, it spawns VMs on all nodes in these farms
func Spawn(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	deploymentStart := time.Now()

	// close ctx on SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	for _, farm := range cfg.Farms {
		log.Info().Uint64("Farm", farm).Msg("running deployment")

		nodes, err := getNodes(ctx, tfPluginClient, farm)
		if err != nil {
			return err
		}
		vmCount := calculateVMCount(nodes, cfg.DeploymentStrategy)
		if vmCount == 0 {
			log.Warn().Msg("there is nothing to deploy")
			return nil
		}
		err = spawn(ctx, tfPluginClient, cfg, nodes, vmCount)
		if err != nil {
			return err
		}
	}
	endTime := time.Since(deploymentStart)
	log.Info().Msgf("deployment took %s", endTime)

	return nil
}

// getNodes returns all the nodes on a specified farm
func getNodes(ctx context.Context, tfPluginClient deployer.TFPluginClient, farm uint64) ([]types.Node, error) {
	trueVal := true
	freeMRU := uint64(memorySize * gb)
	freeSRU := uint64(rootSize * gb)

	filter := types.NodeFilter{
		Status:  []string{"up"},
		Healthy: &trueVal,
		FreeMRU: &freeMRU,
		FreeSRU: &freeSRU,
		FarmIDs: []uint64{farm},
	}
	nodes, err := deployer.FilterNodes(ctx, tfPluginClient, filter, nil, nil, []uint64{freeSRU})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// calculateVMCount calculates the number of VMs to deploy based on the deployment strategy
func calculateVMCount(nodes []types.Node, strategy float64) int {
	totalNodes := len(nodes)

	return int(float64(totalNodes) * strategy)
}

// spawn creates and deploys VMs on the specified nodes according to the provided configuration
func spawn(ctx context.Context, tfPluginClient deployer.TFPluginClient, cfg Config, nodes []types.Node, vmCount int) error {
	networks, vms, err := getDeployment(cfg, nodes, vmCount)
	if err != nil {
		return err
	}

	var resultErr *multierror.Error
	retryCount := 1

	err = retry.Do(ctx, retry.WithMaxRetries(defaultMaxRetries, retry.NewConstant(1*time.Second)), func(ctx context.Context) error {
		if retryCount != 1 {
			log.Info().Int("Retry", retryCount).Msg("Retrying deployment")
		}

		err := tfPluginClient.NetworkDeployer.BatchDeploy(ctx, networks)
		if err != nil {
			log.Debug().Err(err).Msg("Network deployment failed")
			resultErr = multierror.Append(resultErr, err)
		}

		err = tfPluginClient.DeploymentDeployer.BatchDeploy(ctx, vms)
		if err != nil {
			log.Debug().Err(err).Msg("VM deployment failed")
			resultErr = multierror.Append(resultErr, err)
		}

		if resultErr == nil {
			return nil
		}

		switch cfg.FailureStrategy {
		case stop:
			return resultErr

		case destroyAll:
			return Destroy(ctx, cfg, tfPluginClient)

		case retrying:
			vms, networks = identifyFailingResources(vms, networks)

			retryCount++
			return retry.RetryableError(resultErr.ErrorOrNil())

		case destroyFailing:
			_, failingNetworks := identifyFailingResources(vms, networks)

			destroyFailingNetworks(ctx, tfPluginClient, failingNetworks)
		}

		return nil
	})

	if err != nil {
		log.Error().Err(resultErr.ErrorOrNil()).Msg("Deployment failed after retries")
		return resultErr
	}

	return nil
}

// getDeployment creates the deployment configuration for the specified nodes
func getDeployment(cfg Config, nodes []types.Node, vmCount int) ([]*workloads.ZNet, []*workloads.Deployment, error) {
	var networks []*workloads.ZNet
	var vms []*workloads.Deployment

	for i := 0; i < vmCount; i++ {
		node := nodes[i]
		name := fmt.Sprintf("vm/%d", node.FarmID)

		network := workloads.ZNet{
			Name:  fmt.Sprintf("network_%d", node.NodeID),
			Nodes: []uint32{uint32(node.NodeID)},
			IPRange: gridtypes.NewIPNet(net.IPNet{
				IP:   net.IPv4(10, 20, 0, 0),
				Mask: net.CIDRMask(16, 32),
			}),
			AddWGAccess:  false,
			SolutionType: name,
		}
		vm := workloads.VM{
			Name:        fmt.Sprintf("vm_%d", node.NodeID),
			Flist:       "https://hub.grid.tf/amryassir.3bot/benchmark.flist",
			CPU:         cpuCount,
			Planetary:   true,
			Memory:      memorySize * 1024,
			RootfsSize:  rootSize * 1024,
			Entrypoint:  "/sbin/zinit init",
			NetworkName: network.Name,
			EnvVars: map[string]string{
				"INFLUX_URL":    cfg.Influx.URL,
				"INFLUX_ORG":    cfg.Influx.Org,
				"INFLUX_TOKEN":  cfg.Influx.Token,
				"INFLUX_BUCKET": cfg.Influx.Bucket,
				"NODE_ID":       fmt.Sprintf("%d", node.NodeID),
				"FARM_ID":       fmt.Sprintf("%d", node.FarmID),
				"SSH_KEY":       cfg.SSHKey,
			},
		}
		dl := workloads.NewDeployment(
			fmt.Sprintf("vm_%d", node.NodeID),
			uint32(node.NodeID),
			name,
			nil,
			network.Name,
			nil,
			nil,
			[]workloads.VM{vm},
			nil,
			nil,
		)

		networks = append(networks, &network)
		vms = append(vms, &dl)
	}

	return networks, vms, nil
}

// identifyFailingResources identifies the failing resources based on the error
func identifyFailingResources(vms []*workloads.Deployment, networks []*workloads.ZNet) ([]*workloads.Deployment, []*workloads.ZNet) {
	var failingVMs []*workloads.Deployment
	var failingNetworks []*workloads.ZNet

	for idx, vm := range vms {
		if vm.ContractID == 0 || len(networks[idx].NodeDeploymentID) == 0 {
			failingVMs = append(failingVMs, vm)
			failingNetworks = append(failingNetworks, networks[idx])
		}
	}

	return failingVMs, failingNetworks
}
