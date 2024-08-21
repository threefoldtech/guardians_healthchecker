package spawner

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

const (
	defaultMaxRetries = 5
	gb                = 1024 * 1024 * 1024
	cpuCount          = 4
	memorySize        = 8
	rootSize          = 40
)

// RunSpawner given a list of farm IDs, it spawns VMs on all nodes in these farms
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
			log.Fatal().Msg("there is nothing to deploy")
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
	nodes, err := deployer.FilterNodes(ctx, tfPluginClient, filter, []uint64{freeSRU}, nil, nil)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// spawn creates and deploys VMs on the specified nodes according to the provided configuration
func spawn(ctx context.Context, tfPluginClient deployer.TFPluginClient, cfg Config, nodes []types.Node, vmCount int) error {
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
	err := tfPluginClient.NetworkDeployer.BatchDeploy(ctx, networks)
	if err != nil {
		err = handleFailure(ctx, err, cfg, tfPluginClient, networks, vms)
		if err != nil {
			return err
		}
	}
	err = tfPluginClient.DeploymentDeployer.BatchDeploy(ctx, vms)
	if err != nil {
		err = handleFailure(ctx, err, cfg, tfPluginClient, networks, vms)
		if err != nil {
			return err
		}
	}

	return nil
}

// calculateVMCount calculates the number of VMs to deploy based on the deployment strategy
func calculateVMCount(nodes []types.Node, strategy float64) int {
	totalNodes := len(nodes)

	return int(float64(totalNodes) * strategy)
}

// handleFailure handles deployment failures according to the specified failure strategy
func handleFailure(ctx context.Context, err error, cfg Config, tfPluginClient deployer.TFPluginClient, networks []*workloads.ZNet, vms []*workloads.Deployment) error {
	log.Error().Err(err).Msg("deployment failed")

	switch cfg.FailureStrategy {
	case "retry":
		for i := 0; i < defaultMaxRetries; i++ {
			log.Info().Msgf("retrying deployment... Attempt %d/%d\n", i+1, defaultMaxRetries)
			err := deployResources(ctx, tfPluginClient, networks, vms)
			if err == nil {
				return nil
			}
			log.Error().Err(err).Msg("retry failed")
		}
		log.Error().Msg("all retry attempts failed.")
		return err

	case "destroy-all":
		if destroyErr := destroyResources(ctx, tfPluginClient, networks, vms); destroyErr != nil {
			log.Error().Err(destroyErr).Msg("failed to destroy resources")
			return destroyErr
		}
		return err

	case "destroy-failing":
		failingVMs, failingNetworks := identifyFailingResources(vms, networks)
		if destroyErr := destroyResources(ctx, tfPluginClient, failingNetworks, failingVMs); destroyErr != nil {
			log.Error().Err(destroyErr).Msg("failed to destroy failing resources")
			return destroyErr
		}
		return err

	case "stop":
		log.Info().Msg("stopping operation due to failure")
		return err

	default:
		log.Error().Msg("unknown failure strategy")
		return fmt.Errorf("unknown failure strategy: %s", cfg.FailureStrategy)
	}
}

// deployResources deploys the specified networks and VMs
func deployResources(ctx context.Context, tfPluginClient deployer.TFPluginClient, networks []*workloads.ZNet, vms []*workloads.Deployment) error {
	err := tfPluginClient.NetworkDeployer.BatchDeploy(ctx, networks)
	if err != nil {
		return err
	}
	err = tfPluginClient.DeploymentDeployer.BatchDeploy(ctx, vms)
	if err != nil {
		return err
	}

	return nil
}

// destroyResources destroys the specified networks and VMs
func destroyResources(ctx context.Context, tfPluginClient deployer.TFPluginClient, networks []*workloads.ZNet, vms []*workloads.Deployment) error {
	err := tfPluginClient.NetworkDeployer.BatchCancel(ctx, networks)
	if err != nil {
		return err
	}
	for _, vm := range vms {
		err := tfPluginClient.DeploymentDeployer.Cancel(ctx, vm)
		if err != nil {
			return err
		}
	}

	return nil
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
