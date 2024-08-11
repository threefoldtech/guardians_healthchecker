package spawner

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
)

func RunSpawner(ctx context.Context, cfg Config, tfPluginClient deployer.TFPluginClient) error {
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
		log.Info().Uint64("Farm", farm).Msg("Running deployment")

		nodes, err := getNodes(ctx, tfPluginClient, farm)
		if err != nil {
			return err
		}
		vmCount := calculateVMCount(nodes, cfg.DeploymentStrategy)
		err = spawn(ctx, tfPluginClient, cfg, nodes, vmCount)
		if err != nil {
			return err
		}
	}
	endTime := time.Since(deploymentStart)
	log.Info().Msgf("Deployment took %s", endTime)

	return nil
}

func getNodes(ctx context.Context, tfPluginClient deployer.TFPluginClient, farm uint64) ([]types.Node, error) {
	trueVal := true
	freeMRU := uint64(4 * gb)
	freeSRU := uint64(20 * gb)

	filter := types.NodeFilter{
		Status:  []string{"up"},
		Healthy: &trueVal,
		FreeMRU: &freeMRU,
		FreeSRU: &freeSRU,
		FarmIDs: []uint64{farm},
	}
	nodes, err := deployer.FilterNodes(ctx, tfPluginClient, filter, []uint64{freeSRU}, nil, nil)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	return nodes, nil
}

func spawn(ctx context.Context, tfPluginClient deployer.TFPluginClient, cfg Config, nodes []types.Node, vmCount int) error {
	var networks []*workloads.ZNet
	var vms []*workloads.Deployment

	for i := 0; i < vmCount; i++ {
		node := nodes[i]

		network := workloads.ZNet{
			Name:  fmt.Sprintf("network_%d", node.NodeID),
			Nodes: []uint32{uint32(node.NodeID)},
			IPRange: gridtypes.NewIPNet(net.IPNet{
				IP:   net.IPv4(10, 20, 0, 0),
				Mask: net.CIDRMask(16, 32),
			}),
			AddWGAccess:  false,
			SolutionType: fmt.Sprintf("vm/%d", node.FarmID),
		}
		vm := workloads.VM{
			Name:        fmt.Sprintf("vm_%d", node.NodeID),
			Flist:       "https://hub.grid.tf/amryassir.3bot/benchmark.flist",
			CPU:         2,
			Planetary:   true,
			Memory:      1024,
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
			fmt.Sprintf("vm/%d", node.FarmID),
			nil,
			network.Name,
			nil,
			nil,
			[]workloads.VM{vm},
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

func calculateVMCount(nodes []types.Node, strategy string) int {
	totalNodes := len(nodes)

	strategy = strings.TrimSuffix(strategy, "%")
	percent, err := strconv.ParseFloat(strategy, 64)
	if err != nil || percent < 0 || percent > 100 {
		percent = 100
	}

	return int(float64(totalNodes) * (percent / 100))
}

func handleFailure(ctx context.Context, err error, cfg Config, tfPluginClient deployer.TFPluginClient, networks []*workloads.ZNet, vms []*workloads.Deployment) error {
	switch cfg.FailureStrategy {
	case "retry":
		for i := 0; i < defaultMaxRetries; i++ {
			fmt.Printf("Retrying deployment... Attempt %d/%d\n", i+1, defaultMaxRetries)
			err := tfPluginClient.NetworkDeployer.BatchDeploy(ctx, networks)
			if err == nil {
				err = tfPluginClient.DeploymentDeployer.BatchDeploy(ctx, vms)
			}
			if err == nil {
				return nil
			}
		}
		fmt.Println("Retry attempts failed.")
		return err

	case "destroy-all":
		fmt.Println("Destroying all VMs due to failure...")
		for _, vm := range vms {
			err := tfPluginClient.DeploymentDeployer.Cancel(ctx, vm)
			if err != nil {
				fmt.Printf("Failed to destroy VM: %s\n", vm.Name)
			}
		}
		return err

	case "destroy-failing":
		fmt.Println("Destroying VMs in farms with failing nodes...")
		for _, vm := range vms {
			if destroyErr := tfPluginClient.DeploymentDeployer.Cancel(ctx, vm); destroyErr != nil {
				fmt.Printf("Destroyed failing VM: %s\n", vm.Name)
			}
		}
		return err

	case "stop":
		fmt.Println("Stopping operation due to failure...")
		return err

	default:
		fmt.Println("Unknown failure strategy")
		return fmt.Errorf("unknown failure strategy: %s", cfg.FailureStrategy)
	}
}
