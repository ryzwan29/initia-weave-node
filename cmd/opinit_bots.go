package cmd

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models/opinit_bots"
	"github.com/initia-labs/weave/utils"
)

func OPInitBotsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "opinit_bots",
		Short:                      "OPInit bots subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(OPInitBotsKeysSetupCommand())
	cmd.AddCommand(OPInitBotsInitCommand())

	return cmd
}

func OPInitBotsKeysSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup keys for OPInit bots",
		RunE: func(cmd *cobra.Command, args []string) error {
			versions := utils.ListBinaryReleases("https://api.github.com/repos/initia-labs/opinit-bots/releases")
			userHome, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			binaryPath := filepath.Join(userHome, utils.WeaveDataDirectory, "opinitd")
			currentVersion, _ := utils.GetBinaryVersion(binaryPath)
			_, err = tea.NewProgram(opinit_bots.NewOPInitBotVersionSelector(opinit_bots.NewOPInitBotsState(), versions, currentVersion)).Run()
			return err
		},
	}
	return cmd
}

func OPInitBotsInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init for OPInit bots",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := tea.NewProgram(opinit_bots.NewOPInitBotInitSelector(opinit_bots.NewOPInitBotsState())).Run()
			return err
		},
	}
	return cmd
}
