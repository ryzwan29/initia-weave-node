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
