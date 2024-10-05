package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/service"
	"github.com/initia-labs/weave/utils"
)

func InitiaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "initia",
		Short:                      "Initia subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		initiaInitCommand(),
		initiaStartCommand(),
		initiaStopCommand(),
		initiaRestartCommand(),
		initiaLogCommand(),
	)

	return cmd
}

func initiaInitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap your Initia full node",
		RunE: func(cmd *cobra.Command, args []string) error {
			if utils.IsFirstTimeSetup() {
				// Capture both the final model and the error from Run()
				finalModel, err := tea.NewProgram(models.NewExistingAppChecker(initia.NewRunL1NodeNetworkSelect(initia.NewRunL1NodeState()))).Run()
				if err != nil {
					return err
				}

				// Check if the final model is of type WeaveAppSuccessfullyInitialized
				if _, ok := finalModel.(*models.WeaveAppSuccessfullyInitialized); !ok {
					// If the model is not of the expected type, return or handle as needed
					return nil
				}
			}

			_, err := tea.NewProgram(initia.NewRunL1NodeNetworkSelect(initia.NewRunL1NodeState())).Run()
			if err != nil {
				return err
			}
			return nil
		},
	}

	return initCmd
}

func initiaStartCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the initiad full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := service.NewService(service.Initia)
			if err != nil {
				return err
			}
			err = s.Start()
			if err != nil {
				return err
			}
			fmt.Println("Started Initia full node application. You can see the logs with `initia log`")
			return nil
		},
	}

	return startCmd
}

func initiaStopCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the initiad full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := service.NewService(service.Initia)
			if err != nil {
				return err
			}
			err = s.Stop()
			if err != nil {
				return err
			}
			fmt.Println("Stopped Initia full node application.")
			return nil
		},
	}

	return startCmd
}

func initiaRestartCommand() *cobra.Command {
	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the initiad full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := service.NewService(service.Initia)
			if err != nil {
				return err
			}
			err = s.Restart()
			if err != nil {
				return err
			}

			fmt.Println("Started Initia full node application. You can see the logs with `initia log`")
			return nil
		},
	}

	return restartCmd
}

func initiaLogCommand() *cobra.Command {
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Stream the logs of the initiad full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := cmd.Flags().GetInt(FlagN)
			if err != nil {
				return err
			}

			s, err := service.NewService(service.Initia)
			if err != nil {
				return err
			}
			return s.Log(n)
		},
	}

	logCmd.Flags().IntP(FlagN, FlagN, 100, "previous log lines to show")

	return logCmd
}
