package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/common"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/cosmosutils"
	"github.com/initia-labs/weave/models/opinit_bots"
	"github.com/initia-labs/weave/service"
)

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func ValidateOPinitOptionalBotNameArgs(_ *cobra.Command, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("expected zero or one argument, got %d", len(args))
	}
	if len(args) == 1 && !contains([]string{"executor", "challenger"}, args[0]) {
		return fmt.Errorf("invalid bot name '%s'. Valid options are: [executor, challenger]", args[0])
	}
	return nil
}

func ValidateOPinitBotNameArgs(_ *cobra.Command, args []string) error {
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
	cmd.AddCommand(OPInitBotsResetCommand())

	return cmd
}

func Setup(minitiaHome, opInitHome string) (tea.Model, error) {
	versions, currentVersion := cosmosutils.GetOPInitVersions()

	// Initialize the context with OPInitBotsState
	ctx := weavecontext.NewAppContext(opinit_bots.NewOPInitBotsState())
	ctx = weavecontext.SetMinitiaHome(ctx, minitiaHome)
	ctx = weavecontext.SetOPInitHome(ctx, opInitHome)

	// Initialize the OPInitBotVersionSelector with the current context and versions
	versionSelector := opinit_bots.NewOPInitBotVersionSelector(ctx, versions, currentVersion)

	// Start the program
	finalModel, err := tea.NewProgram(versionSelector).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
	}
	return finalModel, err
}

func OPInitBotsKeysSetupCommand() *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup keys for OPInit bots",
		RunE: func(cmd *cobra.Command, args []string) error {
			minitiaHome, _ := cmd.Flags().GetString(FlagMinitiaHome)
			opInitHome, _ := cmd.Flags().GetString(FlagOPInitHome)

			_, err := Setup(minitiaHome, opInitHome)
			return err
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("cannot get user home directory: %v", err))
	}

	setupCmd.Flags().String(FlagMinitiaHome, filepath.Join(homeDir, common.MinitiaDirectory), "Minitia application directory to fetch artifacts from if existed")
	setupCmd.Flags().String(FlagOPInitHome, filepath.Join(homeDir, common.OPinitDirectory), "OPInit bots home directory")

	return setupCmd
}

func OPInitBotsInitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [bot-name]",
		Short: "Initialize OPinit bots",
		Long: `Initialize the OPinit bot. The argument is optional, as you will be prompted to select a bot if no bot name is provided.
Alternatively, you can specify a bot name as an argument to skip the selection. Valid options are [executor, challenger].
Example: weave opinit-bots init executor`,
		Args: ValidateOPinitOptionalBotNameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			minitiaHome, _ := cmd.Flags().GetString(FlagMinitiaHome)
			opInitHome, _ := cmd.Flags().GetString(FlagOPInitHome)

			userHome, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			binaryPath := filepath.Join(userHome, common.WeaveDataDirectory, opinit_bots.AppName)
			_, err = cosmosutils.GetBinaryVersion(binaryPath)
			if err != nil {
				finalModel, err := Setup(minitiaHome, opInitHome)
				if err != nil {
					return err
				}

				if _, ok := finalModel.(*opinit_bots.TerminalState); !ok {
					return nil
				}
			}
			// Initialize the context with OPInitBotsState
			ctx := weavecontext.NewAppContext(opinit_bots.NewOPInitBotsState())
			ctx = weavecontext.SetMinitiaHome(ctx, minitiaHome)
			ctx = weavecontext.SetOPInitHome(ctx, opInitHome)

			// Check if a bot name was provided as an argument
			if len(args) == 1 {
				botName := args[0]
				switch botName {
				case "executor":
					_, err := tea.NewProgram(opinit_bots.OPInitBotInitSelectExecutor(ctx)).Run()
					return err
				case "challenger":
					_, err := tea.NewProgram(opinit_bots.OPInitBotInitSelectChallenger(ctx)).Run()
					return err
				default:
					return fmt.Errorf("invalid bot name provided: %s", botName)
				}
			} else {
				// Start the bot selector program if no bot name is provided
				_, err := tea.NewProgram(opinit_bots.NewOPInitBotInitSelector(ctx)).Run()
				return err
			}
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("cannot get user home directory: %v", err))
	}

	initCmd.Flags().String(FlagMinitiaHome, filepath.Join(homeDir, common.MinitiaDirectory), "Minitia application directory to fetch artifacts from if existed")
	initCmd.Flags().String(FlagOPInitHome, filepath.Join(homeDir, common.OPinitDirectory), "OPInit bots home directory")

	return initCmd
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
			fmt.Printf("Started the OPinit %[1]s bot. You can see the logs with `weave opinit-bots log %[1]s`\n", botName)
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
			fmt.Printf("Restart the OPinit %[1]s bot process. You can see the logs with `weave opinit-bots log %[1]s`\n", botName)
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

func OPInitBotsResetCommand() *cobra.Command {
	resetCmd := &cobra.Command{
		Use:   "reset [bot-name]",
		Short: "Reset a OPinit bot's database",
		Long: `Reset a OPinit bot's database. The only argument required is the desired bot name.
Valid options are [executor, challenger] eg. weave opinit-bots reset challenger`,
		Args: ValidateOPinitBotNameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			userHome, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			binaryPath := filepath.Join(userHome, common.WeaveDataDirectory, opinit_bots.AppName)
			_, err = cosmosutils.GetBinaryVersion(binaryPath)
			if err != nil {
				panic("error getting the opinitd binary")
			}

			botName := args[0]
			execCmd := exec.Command(binaryPath, "reset-db", botName)
			if err = execCmd.Run(); err != nil {
				return fmt.Errorf("failed to reset-db: %v", err)
			}
			fmt.Println(fmt.Sprintf("Reset the OPinit %[1]s bot database successfully.", botName))
			return nil
		},
	}

	return resetCmd
}
