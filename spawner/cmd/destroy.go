package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	spawner "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "destroy VMs on specified farms",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, tfPluginClient, err := loadConfigAndSetup(cmd)
		if err != nil {
			return err
		}
		err = spawner.Destroy(context.Background(), cfg, tfPluginClient)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		return nil
	},
}
