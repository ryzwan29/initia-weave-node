package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/models/weaveinit"
	"github.com/initia-labs/weave/utils"
)

func InitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:  "init",
		Long: "Init for initializing the weave CLI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if utils.IsFirstTimeSetup() {
				_, err := tea.NewProgram(models.NewExistingAppChecker()).Run()
				if err != nil {
					return err
				}
			}
			_, err := tea.NewProgram(weaveinit.NewWeaveInit()).Run()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return initCmd
}

func InitiaInitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:  "initia init",
		Long: "Init for initializing the initia CLI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if utils.IsFirstTimeSetup() {
				_, err := tea.NewProgram(models.NewExistingAppChecker()).Run()
				if err != nil {
					return err
				}
			}

			_, err := tea.NewProgram(weaveinit.NewRunL1NodeNetworkSelect(weaveinit.NewRunL1NodeState())).Run()
			if err != nil {
				return err
			}
			return nil
		},
	}

	return initCmd
}
