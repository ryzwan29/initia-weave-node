package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/config"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/models/weaveinit"
)

func InitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Weave CLI, funding gas station and setting up config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.IsFirstTimeSetup() {
				ctx := weavecontext.NewAppContext(models.NewExistingCheckerState())

				// Capture both the final model and the error from Run()
				finalModel, err := tea.NewProgram(models.NewExistingAppChecker(ctx, weaveinit.NewWeaveInit())).Run()
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
