package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/initia-labs/weave/analytics"
	"github.com/initia-labs/weave/config"
)

var Version string

func Execute() error {
	rootCmd := &cobra.Command{
		Use:  "weave",
		Long: WeaveHelperText,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.AutomaticEnv()
			viper.SetEnvPrefix("weave")
			if err := config.InitializeConfig(); err != nil {
				return err
			}
			analytics.Initialize(Version)
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			analytics.Client.Flush()
			analytics.Client.Shutdown()
			return nil
		},
	}

	rootCmd.AddCommand(
		InitCommand(),
		InitiaCommand(),
		GasStationCommand(),
		VersionCommand(),
		// UpgradeCommand(),
		MinitiaCommand(),
		OPInitBotsCommand(),
		RelayerCommand(),
		AnalyticsCommand(),
	)

	return rootCmd.ExecuteContext(context.Background())
}
