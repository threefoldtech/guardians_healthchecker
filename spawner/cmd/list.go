package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	spawner "github.com/threefoldtech/guardians_healthchecker/spawner/pkg/spawner"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list VMs and how long they have been running",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, tfPluginClient, err := loadConfigAndSetup(cmd)
		if err != nil {
			return err
		}
		err = spawner.List(context.Background(), cfg, tfPluginClient)
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		return nil
	},
}
