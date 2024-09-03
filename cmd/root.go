package cmd

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/initia-labs/weave/states"
)

func Execute() error {
	rootCmd := &cobra.Command{
		Version: "v1.0.0",
		Use:     "weave",
		Long:    "Weave is a CLI for managing Initia deployments.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			viper.AutomaticEnv()
			viper.SetEnvPrefix("weave")

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := tea.NewProgram(states.NewHomePage(states.DefaultStates())).Run()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return rootCmd.ExecuteContext(context.Background())
}
