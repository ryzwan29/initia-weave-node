package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/models/minitia"
	"github.com/initia-labs/weave/utils"
)

func MinitiaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "minitia",
		Short:                      "Minitia subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		minitiaLaunchCommand(),
	)

	return cmd
}

func minitiaLaunchCommand() *cobra.Command {
	launchCmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch for initializing a new L2 node.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if utils.IsFirstTimeSetup() {
				// Capture both the final model and the error from Run()
				finalModel, err := tea.NewProgram(models.NewExistingAppChecker()).Run()
				if err != nil {
					return err
				}

				// Check if the final model is of type WeaveAppSuccessfullyInitialized
				if _, ok := finalModel.(*models.WeaveAppSuccessfullyInitialized); !ok {
					// If the model is not of the expected type, return or handle as needed
					return nil
				}
			}

			_, err := tea.NewProgram(minitia.NewExistingMinitiaChecker(minitia.NewLaunchState())).Run()
			if err != nil {
				return err
			}
			return nil
		},
	}

	return launchCmd
}
