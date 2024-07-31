package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spawner",
	Short: "A tool used for spawning and destroying benchmark VMs",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(spawnCmd)
	// rootCmd.AddCommand(destroyCmd)
	// rootCmd.AddCommand(listCmd)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}

func init() {
	spawnCmd.Flags().BoolP("debug", "d", false, "allow debug logs")
	spawnCmd.Flags().StringP("config", "c", "", "path to config file")

	// destroyCmd.Flags().BoolP("debug", "d", false, "allow debug logs")
	// destroyCmd.Flags().StringP("config", "c", "", "path to config file")

	// listCmd.Flags().BoolP("debug", "d", false, "allow debug logs")
	// listCmd.Flags().StringP("config", "c", "", "path to config file")
}
