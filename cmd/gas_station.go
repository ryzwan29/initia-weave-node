package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/utils"
)

func GasStationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "gas-station",
		Short:                      "Gas Station subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		gasStationSetupCommand(),
		gasStationShowCommand(),
	)

	return cmd
}

func gasStationSetupCommand() *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup Gas Station account on Initia and Celestia for funding the OPinit-bots or relayer to send transactions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement this
			return nil
		},
	}

	return setupCmd
}

func gasStationShowCommand() *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show Initia and Celestia Gas Station addresses and balances",
		RunE: func(cmd *cobra.Command, args []string) error {
			if utils.IsFirstTimeSetup() {
				fmt.Println("Please setup Gas Station first, by running `gas-station setup`")
				return nil
			}

			// TODO: Implement this
			return nil
		},
	}

	return showCmd
}
