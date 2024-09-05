package cmd

import (
	"context"
	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() error {
	rootCmd := &cobra.Command{
		Version: "v1.0.0",
		Use:     "weave",
		Long:    "Weave is a CLI for managing Initia deployments.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.AutomaticEnv()
			viper.SetEnvPrefix("weave")

			if err := utils.InitializeConfig(); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := tea.NewProgram(models.NewHomepage()).Run()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return rootCmd.ExecuteContext(context.Background())
}
