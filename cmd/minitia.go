package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/models/minitia"
	"github.com/initia-labs/weave/service"
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
		minitiaStartCommand(),
		minitiaStopCommand(),
		minitiaRestartCommand(),
		minitiaLogCommand(),
	)

	return cmd
}

func minitiaLaunchCommand() *cobra.Command {
	launchCmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch a new Minitia from scratch",
		RunE: func(cmd *cobra.Command, args []string) error {
			if utils.IsFirstTimeSetup() {
				finalModel, err := tea.NewProgram(models.NewExistingAppChecker(minitia.NewExistingMinitiaChecker(minitia.NewLaunchState()))).Run()
				if err != nil {
					return err
				}

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

func minitiaStartCommand() *cobra.Command {
	launchCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Minitia full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := service.NewService(service.Minitia)
			if err != nil {
				return err
			}
			err = s.Start()
			if err != nil {
				return err
			}
			fmt.Println("Started Minitia full node application. You can see the logs with `minitia log`")
			return nil
		},
	}

	return launchCmd
}

func minitiaStopCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the Minitia full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := service.NewService(service.Minitia)
			if err != nil {
				return err
			}
			err = s.Stop()
			if err != nil {
				return err
			}
			fmt.Println("Stopped the Minitia full node application.")
			return nil
		},
	}

	return startCmd
}

func minitiaRestartCommand() *cobra.Command {
	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the Minitia full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := service.NewService(service.Minitia)
			if err != nil {
				return err
			}
			err = s.Restart()
			if err != nil {
				return err
			}

			fmt.Println("Restart Minitia full node application. You can see the logs with `minitia log`")
			return nil
		},
	}

	return restartCmd
}

func minitiaLogCommand() *cobra.Command {
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Stream the logs of the Minitia full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := cmd.Flags().GetInt(FlagN)
			if err != nil {
				return err
			}

			s, err := service.NewService(service.Minitia)
			if err != nil {
				return err
			}
			return s.Log(n)
		},
	}

	logCmd.Flags().IntP(FlagN, FlagN, 100, "previous log lines to show")

	return logCmd
}
