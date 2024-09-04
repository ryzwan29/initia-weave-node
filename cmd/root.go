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

			states.GetGlobalStorage() // Ensure global storage is initialized

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			// Initialize state transitions and ensure they're set up correctly
			states.GetHomePage() // Initialize homepage state

			states.SetHomePageTransitions()           // Set homepage transitions lazily after initialization
			states.SetInitiaInitTransitions()         // Initialize transitions for InitiaInit
			states.SetRunL1NodeTransitions()          // Initialize transitions for RunL1Node
			states.SetRunL1NodeTextInputTransitions() // Initialize transitions for RunL1NodeTextInput

			_, err := tea.NewProgram(states.GetHomePage()).Run()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return rootCmd.ExecuteContext(context.Background())
}
