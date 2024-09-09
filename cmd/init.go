package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models/weaveinit"
)

func InitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:  "init",
		Long: "Init for initializing the weave CLI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := tea.NewProgram(weaveinit.NewWeaveInit()).Run()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return initCmd
}
