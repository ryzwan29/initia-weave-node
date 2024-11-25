package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/initia-labs/weave/config"
	"github.com/initia-labs/weave/flags"
)

var Version string

func Execute() error {
	rootCmd := &cobra.Command{
		Version: Version,
		Use:     "weave",
		Long:    "Weave is the CLI for managing Initia deployments.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.AutomaticEnv()
			viper.SetEnvPrefix("weave")
			if err := config.InitializeConfig(); err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.AddCommand(InitCommand())
	rootCmd.AddCommand(InitiaCommand())
	rootCmd.AddCommand(GasStationCommand())

	if flags.IsEnabled(flags.MinitiaLaunch) {
		rootCmd.AddCommand(MinitiaCommand())
	}

	if flags.IsEnabled(flags.OPInitBots) {
		rootCmd.AddCommand(OPInitBotsCommand())
	}

	if flags.IsEnabled(flags.Relayer) {
		rootCmd.AddCommand(RelayerCommand())
	}

	return rootCmd.ExecuteContext(context.Background())
}
