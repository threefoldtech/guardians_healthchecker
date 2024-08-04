package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/guardians_healthchecker/spawner/internal/parser"
	spawner "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "destroy VMs on specified farms",
	RunE: func(cmd *cobra.Command, args []string) error {
		// It doesn't have a subcommand
		if len(cmd.Flags().Args()) != 0 {
			return fmt.Errorf("'destroy' and %v cannot be used together, please use one command at a time", cmd.Flags().Args())
		}

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return fmt.Errorf("error in configuration file: %w", err)
		}

		if configPath == "" {
			return fmt.Errorf("required configuration file path is empty")
		}

		configFile, err := os.Open(configPath)
		if err != nil {
			return fmt.Errorf("failed to open configuration file '%s' with error: %w", configPath, err)
		}
		defer configFile.Close()

		yamlFmt := filepath.Ext(configPath) == ".yaml"
		if !yamlFmt {
			return fmt.Errorf("unsupported configuration file format '%s', should be .yaml", configPath)
		}

		cfg, err := parser.ParseConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to parse configuration file '%s' with error: %w", configPath, err)
		}

		tfPluginClient, err := setup(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		err = spawner.DestroyVms(ctx, cfg, tfPluginClient)
		if err != nil {
			return fmt.Errorf("failed to destroy VMs with error: %w", err)
		}

		return nil
	},
}
