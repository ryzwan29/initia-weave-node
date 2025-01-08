package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/analytics"
	"github.com/initia-labs/weave/common"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/cosmosutils"
	"github.com/initia-labs/weave/io"
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
		Use:                        "opinit",
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

type HomeConfig struct {
	MinitiaHome string
	OPInitHome  string
	UserHome    string
}

func RunOPInit(nextModelFunc func(ctx context.Context) (tea.Model, error), homeConfig HomeConfig) (tea.Model, error) {
	// Initialize the context with OPInitBotsState
	ctx := weavecontext.NewAppContext(opinit_bots.NewOPInitBotsState())
	ctx = weavecontext.SetMinitiaHome(ctx, homeConfig.MinitiaHome)
	ctx = weavecontext.SetOPInitHome(ctx, homeConfig.OPInitHome)

	// Start the program
	if finalModel, err := tea.NewProgram(
		opinit_bots.NewEnsureOPInitBotsBinaryLoadingModel(
			ctx,
			func(nextCtx context.Context) (tea.Model, error) {
				return opinit_bots.ProcessMinitiaConfig(nextCtx, nextModelFunc)
			},
		), tea.WithAltScreen(),
	).Run(); err != nil {
		fmt.Println("Error running program:", err)
		return finalModel, err
	} else {
		fmt.Println(finalModel.View())
		return finalModel, nil
	}
}

func OPInitBotsKeysSetupCommand() *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup-keys",
		Short: "Setup keys for OPInit bots",
		RunE: func(cmd *cobra.Command, args []string) error {
			analytics.TrackRunEvent(cmd, args, analytics.OPinitComponent)
			minitiaHome, _ := cmd.Flags().GetString(FlagMinitiaHome)
			opInitHome, _ := cmd.Flags().GetString(FlagOPInitHome)

			userHome, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("error getting user home directory: %v", err)
			}

			_, err = RunOPInit(opinit_bots.NewSetupBotCheckbox, HomeConfig{
				MinitiaHome: minitiaHome,
				OPInitHome:  opInitHome,
				UserHome:    userHome,
			})
			analytics.TrackCompletedEvent(cmd, analytics.OPinitComponent)

			return err
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("cannot get user home directory: %v", err))
	}

	setupCmd.Flags().String(FlagMinitiaHome, filepath.Join(homeDir, common.MinitiaDirectory), "Rollup application directory to fetch artifacts from if existed")
	setupCmd.Flags().String(FlagOPInitHome, filepath.Join(homeDir, common.OPinitDirectory), "OPInit bots home directory")

	return setupCmd
}

func generateKeyFile(userHome string, keyPath string) (opinit_bots.KeyFile, error) {
	keyFile, err := opinit_bots.GenerateMnemonicKeyfile()
	if err != nil {
		return keyFile, err
	}

	// Marshal KeyFile to JSON
	data, err := json.MarshalIndent(keyFile, "", "  ")
	if err != nil {
		return keyFile, fmt.Errorf("error marshaling KeyFile to JSON: %w", err)
	}

	// Write JSON data to a file
	err = os.WriteFile(keyPath, data, 0644)
	if err != nil {
		return keyFile, fmt.Errorf("error writing to file: %w", err)
	}

	return keyFile, nil
}

func validateConfigFlags(configPath, keyFilePath string, isGenerateKeyFile bool) error {
	if configPath != "" {
		if keyFilePath != "" && isGenerateKeyFile {
			return fmt.Errorf("invalid configuration: both keyFilePath and isGenerateKeyFile cannot be set at the same time")
		}
		if keyFilePath == "" && !isGenerateKeyFile {
			return fmt.Errorf("invalid configuration: if configPath is set, either keyFilePath or isGenerateKeyFile must be provided")
		}
		if !io.FileOrFolderExists(configPath) {
			return fmt.Errorf("the provided configPath does not exist: %s", configPath)
		}
	} else {
		// If configPath is empty, neither keyFilePath nor isGenerateKeyFile should be set
		if keyFilePath != "" || isGenerateKeyFile {
			return fmt.Errorf("invalid configuration: if configPath is not set, neither keyFilePath nor isGenerateKeyFile should be provided")
		}
	}

	return nil
}

func handleWithConfig(userHome, opInitHome, configPath, keyFilePath string, args []string, force, isGenerateKeyFile bool) error {
	botName := args[0]
	if botName != "executor" && botName != "challenger" {
		return fmt.Errorf("bot name '%s' is not recognized. Allowed values are 'executor' or 'challenger'", botName)
	}

	fileData, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var keyFile opinit_bots.KeyFile
	if isGenerateKeyFile {
		keyPath := filepath.Join(userHome, common.WeaveDataDirectory, common.OpinitGeneratedKeyFilename)
		keyFile, err = generateKeyFile(userHome, keyPath)
		if err != nil {
			return err
		}
		fmt.Printf("Key file successfully generated. You can find it at: %s\n", keyPath)
	} else {
		if !io.FileOrFolderExists(keyFilePath) {
			return fmt.Errorf("key file is missing at path: %s", keyFilePath)
		}

		// Read and unmarshal key file data
		keyFile, err = readAndUnmarshalKeyFile(keyFilePath)
		if err != nil {
			return err
		}
	}
	// Handle existing opInitHome directory
	if err := handleExistingOpInitHome(opInitHome, force); err != nil {
		return err
	}

	// Process bot config based on arguments
	if len(args) != 1 {
		return fmt.Errorf("please specify bot name")
	}

	return initializeBotWithConfig(fileData, keyFile, opInitHome, userHome, botName)
}

// readAndUnmarshalKeyFile read and unmarshal the key file into the KeyFile struct
func readAndUnmarshalKeyFile(keyFilePath string) (opinit_bots.KeyFile, error) {
	fileData, err := os.ReadFile(keyFilePath)
	if err != nil {
		return opinit_bots.KeyFile{}, err
	}

	var keyFile opinit_bots.KeyFile
	err = json.Unmarshal(fileData, &keyFile)
	return keyFile, err
}

// handleExistingOpInitHome handle the case where the opInitHome directory exists
func handleExistingOpInitHome(opInitHome string, force bool) error {
	if io.FileOrFolderExists(opInitHome) {
		if force {
			if err := io.DeleteDirectory(opInitHome); err != nil {
				return fmt.Errorf("failed to delete %s: %v", opInitHome, err)
			}
		} else {
			return fmt.Errorf("existing %s folder detected. Use --force or -f to override", opInitHome)
		}
	}
	return nil
}

// initializeBotWithConfig initialize a bot based on the provided config
func initializeBotWithConfig(fileData []byte, keyFile opinit_bots.KeyFile, opInitHome, userHome, botName string) error {
	var err error

	switch botName {
	case "executor":
		var config opinit_bots.ExecutorConfig
		err = json.Unmarshal(fileData, &config)
		if err != nil {
			return err
		}
		err = opinit_bots.InitializeExecutorWithConfig(config, &keyFile, opInitHome, userHome)
	case "challenger":
		var config opinit_bots.ChallengerConfig
		err = json.Unmarshal(fileData, &config)
		if err != nil {
			return err
		}
		err = opinit_bots.InitializeChallengerWithConfig(config, &keyFile, opInitHome, userHome)
	}
	if err != nil {
		return err
	}

	fmt.Printf("OPInit bot setup successfully. Config file is saved at %s. Feel free to modify it as needed.\n", filepath.Join(opInitHome, fmt.Sprintf("%s.json", botName)))
	return nil
}

func OPInitBotsInitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [bot-name]",
		Short: "Initialize OPinit bots",
		Long: `Initialize the OPinit bot. The argument is optional, as you will be prompted to select a bot if no bot name is provided.
Alternatively, you can specify a bot name as an argument to skip the selection. Valid options are [executor, challenger].
Example: weave opinit init executor`,
		Args: ValidateOPinitOptionalBotNameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			analytics.TrackRunEvent(cmd, args, analytics.OPinitComponent)
			minitiaHome, _ := cmd.Flags().GetString(FlagMinitiaHome)
			opInitHome, _ := cmd.Flags().GetString(FlagOPInitHome)
			force, _ := cmd.Flags().GetBool(FlagForce)
			configPath, _ := cmd.Flags().GetString(FlagWithConfig)
			keyFilePath, _ := cmd.Flags().GetString(FlagKeyFile)
			isGenerateKeyFile, _ := cmd.Flags().GetBool(FlagGenerateKeyFile)

			userHome, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("error getting user home directory: %v", err)
			}

			err = validateConfigFlags(configPath, keyFilePath, isGenerateKeyFile)
			if err != nil {
				return err
			}
			if configPath != "" {
				return handleWithConfig(userHome, opInitHome, configPath, keyFilePath, args, force, isGenerateKeyFile)
			}

			var rootProgram func(ctx context.Context) (tea.Model, error)
			// Check if a bot name was provided as an argument
			if len(args) == 1 {
				botName := args[0]
				switch botName {
				case "executor":
					rootProgram = opinit_bots.OPInitBotInitSelectExecutor
				case "challenger":
					rootProgram = opinit_bots.OPInitBotInitSelectChallenger
				default:
					return fmt.Errorf("invalid bot name provided: %s", botName)
				}
			} else {
				// Start the bot selector program if no bot name is provided
				rootProgram = opinit_bots.NewOPInitBotInitSelector
			}

			_, err = RunOPInit(rootProgram, HomeConfig{
				MinitiaHome: minitiaHome,
				OPInitHome:  opInitHome,
				UserHome:    userHome,
			})

			analytics.TrackCompletedEvent(cmd, analytics.OPinitComponent)
			return err
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("cannot get user home directory: %v", err))
	}

	initCmd.Flags().String(FlagMinitiaHome, filepath.Join(homeDir, common.MinitiaDirectory), "Rollup application directory to fetch artifacts from if existed")
	initCmd.Flags().String(FlagOPInitHome, filepath.Join(homeDir, common.OPinitDirectory), "OPInit bots home directory")
	initCmd.Flags().String(FlagWithConfig, "", "Bypass the interactive setup and initialize the bot by providing a path to a config file. Either --key-file or --generate-key-file has to be specified")
	initCmd.Flags().String(FlagKeyFile, "", "Path to key-file.json. Cannot be specified together with --generate-key-file")
	initCmd.Flags().BoolP(FlagForce, "f", false, "Force the setup by deleting the existing .opinit directory if it exists")
	initCmd.Flags().BoolP(FlagGenerateKeyFile, "", false, "Use this flag to generate the bot keys. Cannot be specified together with --key-file")

	return initCmd
}

func OPInitBotsStartCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start [bot-name]",
		Short: "Start the OPinit bot process.",
		Long: `Use this command to start the OPinit bot, where the only argument required is the desired bot name. 
Valid options are [executor, challenger] eg. weave opinit start executor
 `,
		Args: ValidateOPinitBotNameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			detach, err := cmd.Flags().GetBool(FlagDetach)
			if err != nil {
				return err
			}

			botName := args[0]
			bot := service.CommandName(botName)
			s, err := service.NewService(bot)
			if err != nil {
				return err
			}

			if detach {
				err = s.Start()
				if err != nil {
					return err
				}
				fmt.Printf("Started the OPinit %[1]s bot. You can see the logs with `weave opinit log %[1]s`\n", botName)
				return nil
			}

			return service.NonDetachStart(s)
		},
	}

	startCmd.Flags().BoolP(FlagDetach, "d", false, "Run the OPinit bot in detached mode")

	return startCmd
}

func OPInitBotsStopCommand() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "stop [bot-name]",
		Short: "Stop the running OPinit bot process.",
		Long: `Use this command to stop the running OPinit bot, where the only argument required is the desired bot name.
Valid options are [executor, challenger] eg. weave opinit stop challenger`,
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
			fmt.Printf("Stopped the OPinit %s bot process.\n", botName)
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
Valid options are [executor, challenger] eg. weave opinit restart executor`,
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
			fmt.Printf("Restart the OPinit %[1]s bot process. You can see the logs with `weave opinit log %[1]s`\n", botName)
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
Valid options are [executor, challenger] eg. weave opinit log executor`,
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
Valid options are [executor, challenger] eg. weave opinit reset challenger`,
		Args: ValidateOPinitBotNameArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			analytics.TrackRunEvent(cmd, args, analytics.OPinitComponent)
			userHome, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("error getting user home directory: %v", err)
			}

			binaryPath := filepath.Join(userHome, common.WeaveDataDirectory, opinit_bots.AppName)
			_, err = cosmosutils.GetBinaryVersion(binaryPath)
			if err != nil {
				return fmt.Errorf("error getting the opinitd binary: %v", err)
			}

			botName := args[0]
			execCmd := exec.Command(binaryPath, "reset-db", botName)
			if err = execCmd.Run(); err != nil {
				return fmt.Errorf("failed to reset-db: %v", err)
			}
			fmt.Printf("Reset the OPinit %s bot database successfully.\n", botName)
			return nil
		},
	}

	return resetCmd
}
