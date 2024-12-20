package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/initia-labs/weave/config"
)

var Version string

func Execute() error {
	rootCmd := &cobra.Command{
		Use:  "weave",
		Long: "Weave is the CLI for managing Initia deployments.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.AutomaticEnv()
			viper.SetEnvPrefix("weave")
			if err := config.InitializeConfig(); err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.AddCommand(
		InitCommand(),
		InitiaCommand(),
		GasStationCommand(),
		VersionCommand(),
		UpgradeCommand(),
		MinitiaCommand(),
		OPInitBotsCommand(),
		RelayerCommand(),
	)

	return rootCmd.ExecuteContext(context.Background())
}
