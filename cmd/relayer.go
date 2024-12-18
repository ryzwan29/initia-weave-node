package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/common"
	"github.com/initia-labs/weave/config"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/models/relayer"
	"github.com/initia-labs/weave/service"
)

func RelayerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "relayer",
		Short:                      "Relayer subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		relayerInitCommand(),
		relayerStartCommand(),
		relayerStopCommand(),
		relayerRestartCommand(),
		relayerLogCommand(),
	)

	return cmd
}

func relayerInitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize and configure your Hermes relayer for IBC",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := weavecontext.NewAppContext(relayer.NewRelayerState())
			minitiaHome, _ := cmd.Flags().GetString(FlagMinitiaHome)
			ctx = weavecontext.SetMinitiaHome(ctx, minitiaHome)

			if config.IsFirstTimeSetup() {
				checkerCtx := weavecontext.NewAppContext(models.NewExistingCheckerState())
				finalModel, err := tea.NewProgram(models.NewExistingAppChecker(checkerCtx, relayer.NewRollupSelect(ctx))).Run()
				if err != nil {
					return err
				}

				// Check if the final model is of type WeaveAppSuccessfullyInitialized
				if _, ok := finalModel.(*models.WeaveAppSuccessfullyInitialized); !ok {
					// If the model is not of the expected type, return or handle as needed
					return nil
				}
			}

			_, err := tea.NewProgram(relayer.NewRollupSelect(ctx)).Run()
			if err != nil {
				return err
			}
			return nil
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("cannot get user home directory: %v", err))
	}

	initCmd.Flags().String(FlagMinitiaHome, filepath.Join(homeDir, common.MinitiaDirectory), "Minitia application directory to fetch artifacts from if existed")

	return initCmd
}

func relayerStartCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Hermes relayer application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			updateClient, err := cmd.Flags().GetString(FlagUpdateClient)
			if err != nil {
				return err
			}

			switch updateClient {
			case "true":
				relayer.UpdateClientFromConfig()
				fmt.Println("Client updated successfully.")
			case "false":
			default:
				panic("invalid update client flag")
			}
			s, err := service.NewService(service.Relayer)
			if err != nil {
				return err
			}
			err = s.Start()
			if err != nil {
				return err
			}
			fmt.Println("Started Hermes relayer application. You can see the logs with `weave relayer log`")
			return nil
		},
	}
	startCmd.Flags().String(FlagUpdateClient, "true", "Update the client configuration before starting the relayer")

	return startCmd
}

func relayerStopCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the Hermes relayer application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := service.NewService(service.Relayer)
			if err != nil {
				return err
			}
			err = s.Stop()
			if err != nil {
				return err
			}
			fmt.Println("Stopped Hermes relayer application.")
			return nil
		},
	}

	return startCmd
}

func relayerRestartCommand() *cobra.Command {
	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the Hermes relayer application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := service.NewService(service.Relayer)
			if err != nil {
				return err
			}
			err = s.Restart()
			if err != nil {
				return err
			}

			fmt.Println("Started Hermes relayer application. You can see the logs with `weave relayer log`")
			return nil
		},
	}

	return restartCmd
}

func relayerLogCommand() *cobra.Command {
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Stream the logs of the Hermes relayer application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := cmd.Flags().GetInt(FlagN)
			if err != nil {
				return err
			}

			s, err := service.NewService(service.Relayer)
			if err != nil {
				return err
			}
			return s.Log(n)
		},
	}

	logCmd.Flags().IntP(FlagN, FlagN, 100, "previous log lines to show")

	return logCmd
}
