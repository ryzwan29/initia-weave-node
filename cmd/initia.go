package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/common"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/service"
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
			initiaHome, err := cmd.Flags().GetString(FlagInitiaHome)
			if err != nil {
				return err
			}

			ctx := weavecontext.NewAppContext(initia.NewRunL1NodeState())
			ctx = weavecontext.SetInitiaHome(ctx, initiaHome)
			model, err := initia.NewRunL1NodeNetworkSelect(ctx)
			if err != nil {
				return err
			}
			if finalModel, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
				return err
			} else {
				fmt.Println(finalModel.View())
				return nil
			}
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("cannot get user home directory: %v", err))
	}

	initCmd.Flags().String(FlagInitiaHome, filepath.Join(homeDir, common.InitiaDirectory), "The Initia application home directory")

	return initCmd
}

func initiaStartCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the initiad full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, err := cmd.Flags().GetBool(FlagDaemon)
			if err != nil {
				return err
			}

			s, err := service.NewService(service.UpgradableInitia)
			if err != nil {
				return err
			}

			if daemon {
				err = s.Start()
				if err != nil {
					return err
				}
				fmt.Println("Started Initia full node application. You can see the logs with `weave initia log`")
				return nil
			}

			return service.NonDaemonStart(s)
		},
	}

	startCmd.Flags().BoolP(FlagDaemon, "d", false, "Run the initiad full node application as a daemon")

	return startCmd
}

func initiaStopCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the initiad full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := service.NewService(service.UpgradableInitia)
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
			s, err := service.NewService(service.UpgradableInitia)
			if err != nil {
				return err
			}
			err = s.Restart()
			if err != nil {
				return err
			}

			fmt.Println("Started Initia full node application. You can see the logs with `weave initia log`")
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

			s, err := service.NewService(service.UpgradableInitia)
			if err != nil {
				return err
			}
			return s.Log(n)
		},
	}

	logCmd.Flags().IntP(FlagN, FlagN, 100, "previous log lines to show")

	return logCmd
}
