package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	spawner "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list VMs and how long they have been running",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cfg, tfPluginClient, err := loadConfigAndSetup(cmd)
		if err != nil {
			return err
		}
		err = spawner.List(ctx, cfg, tfPluginClient)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		return nil
	},
}
