package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/analytics"
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
			analytics.TrackRunEvent(cmd, analytics.InitComponent)
			if config.IsFirstTimeSetup() {
				ctx := weavecontext.NewAppContext(models.NewExistingCheckerState())

				// Capture both the final model and the error from Run()
				if finalModel, err := tea.NewProgram(models.NewGasStationMethodSelect(ctx), tea.WithAltScreen()).Run(); err != nil {
					return err
				} else {
					fmt.Println(finalModel.View())
					if _, ok := finalModel.(*models.WeaveAppSuccessfullyInitialized); !ok {
						analytics.TrackCompletedEvent(cmd, analytics.InitComponent)
						return nil
					}
				}
			}

			if finalModel, err := tea.NewProgram(weaveinit.NewWeaveInit(), tea.WithAltScreen()).Run(); err != nil {
				return err
			} else {
				analytics.TrackCompletedEvent(cmd, analytics.InitComponent)
				fmt.Println(finalModel.View())
				return nil
			}
		},
	}

	return initCmd
}
