package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/models/opinit_bots"
	"github.com/spf13/cobra"
)

func OPInitBotsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "opinit_bots",
		Short:                      "OPInit bots subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(OPInitBotsKeysCommand())

	return cmd
}

func OPInitBotsKeysCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "OPInit bots keys subcommands",
	}

	cmd.AddCommand(OPInitBotsKeysSetupCommand())

	return cmd
}

func OPInitBotsKeysSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup keys for OPInit bots",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := tea.NewProgram(opinit_bots.NewSetupBotCheckbox(opinit_bots.NewOPInitBotsState())).Run()
			return err
		},
	}
	return cmd
}
