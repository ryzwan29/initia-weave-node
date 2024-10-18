package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models/opinit_bots"
	"github.com/initia-labs/weave/service"
	"github.com/initia-labs/weave/utils"
)

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func ValidateOPinitBotNameArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected exactly one argument, got %d", len(args))
	}
	if !contains([]string{"executor", "challenger"}, args[0]) {
		return fmt.Errorf("invalid bot name '%s'. Valid options are: [executor, challenger]", args[0])
	}
	return nil
}

func OPInitBotsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "opinit-bots",
		Short:                      "OPInit bots subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(OPInitBotsKeysSetupCommand())
	cmd.AddCommand(OPInitBotsInitCommand())
	cmd.AddCommand(OPInitBotsStartCommand())
	cmd.AddCommand(OPInitBotsStopCommand())
	cmd.AddCommand(OPInitBotsRestartCommand())
	cmd.AddCommand(OPInitBotsLogCommand())

	return cmd
}

func Setup() error {
	versions := utils.ListBinaryReleases("https://api.github.com/repos/initia-labs/opinit-bots/releases")
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	binaryPath := filepath.Join(userHome, utils.WeaveDataDirectory, "opinitd")
	currentVersion, _ := utils.GetBinaryVersion(binaryPath)
	_, err = tea.NewProgram(opinit_bots.NewOPInitBotVersionSelector(opinit_bots.NewOPInitBotsState(), versions, currentVersion)).Run()
	return err
}

func OPInitBotsKeysSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup keys for OPInit bots",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Setup()
		},
	}
	return cmd
}

func OPInitBotsInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init OPInit bots",
		RunE: func(cmd *cobra.Command, args []string) error {
			userHome, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			binaryPath := filepath.Join(userHome, utils.WeaveDataDirectory, "opinitd")
			_, err = utils.GetBinaryVersion(binaryPath)
			if err != nil {
				Setup()
			}
			_, err = tea.NewProgram(opinit_bots.NewOPInitBotInitSelector(opinit_bots.NewOPInitBotsState())).Run()
			return err
		},
	}
	return cmd
}

func OPInitBotsStartCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start [bot-name]",
		Short: "Start the OPinit bot process.",
		Long: `Use this command to start the OPinit bot, where the only argument required is the desired bot name. 
Valid options are [executor, challenger] eg. weave opinit-bots start executor
 `,
		Args: ValidateOPinitBotNameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			botName := args[0]
			bot := service.CommandName(botName)
			s, err := service.NewService(bot)
			if err != nil {
				return err
			}
			err = s.Start()
			if err != nil {
				return err
			}
			fmt.Println(fmt.Sprintf("Started the OPinit %[1]s bot. You can see the logs with `opinit-bots log %[1]s`", botName))
			return nil
		},
	}

	return startCmd
}

func OPInitBotsStopCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "stop [bot-name]",
		Short: "Stop the running OPinit bot process.",
		Long: `Use this command to stop the running OPinit bot, where the only argument required is the desired bot name.
Valid options are [executor, challenger] eg. weave opinit-bots stop challenger`,
		Args: ValidateOPinitBotNameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			botName := args[0]
			bot := service.CommandName(botName)
			s, err := service.NewService(bot)
			if err != nil {
				return err
			}
			err = s.Stop()
			if err != nil {
				return err
			}
			fmt.Println(fmt.Sprintf("Stopped the OPinit %s bot process.", botName))
			return nil
		},
	}

	return startCmd
}

func OPInitBotsRestartCommand() *cobra.Command {
	restartCmd := &cobra.Command{
		Use:   "restart [bot-name]",
		Short: "Restart the running OPinit bot process.",
		Long: `Use this command to restart the running OPinit bot, where the only argument required is the desired bot name.
Valid options are [executor, challenger] eg. weave opinit-bots restart executor`,
		Args: ValidateOPinitBotNameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			botName := args[0]
			bot := service.CommandName(botName)
			s, err := service.NewService(bot)
			if err != nil {
				return err
			}
			err = s.Restart()
			if err != nil {
				return err
			}
			fmt.Println(fmt.Sprintf("Restart the OPinit %[1]s bot process. You can see the logs with `opinit-bots log %[1]s`", botName))
			return nil
		},
	}

	return restartCmd
}

func OPInitBotsLogCommand() *cobra.Command {
	logCmd := &cobra.Command{
		Use:   "log [bot-name]",
		Short: "Stream the logs of the running OPinit bot process.",
		Long: `Stream the logs of the running OPinit bot. The only argument required is the desired bot name.
Valid options are [executor, challenger] eg. weave opinit-bots logs executor`,
		Args: ValidateOPinitBotNameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := cmd.Flags().GetInt(FlagN)
			if err != nil {
				return err
			}

			botName := args[0]
			bot := service.CommandName(botName)
			s, err := service.NewService(bot)
			if err != nil {
				return err
			}
			return s.Log(n)
		},
	}

	logCmd.Flags().IntP(FlagN, FlagN, 100, "previous log lines to show")

	return logCmd
}
