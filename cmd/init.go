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
		Use:   "init",
		Short: "Init for initializing the weave CLI.",
		Long:  "Init for initializing the weave CLI.",
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

			_, err := tea.NewProgram(weaveinit.NewWeaveInit()).Run()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return initCmd
}
