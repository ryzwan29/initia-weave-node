package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/utils"
)

const (
	PreviousLogLines = 100
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
		initiaLogCommand(),
	)

	return cmd
}

func initiaInitCommand() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Init for initializing the initia CLI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if utils.IsFirstTimeSetup() {
				// Capture both the final model and the error from Run()
				finalModel, err := tea.NewProgram(models.NewExistingAppChecker()).Run()
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
			err := utils.StartService(utils.GetRunL1NodeServiceName())
			if err != nil {
				return err
			}
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
			err := utils.StopService(utils.GetRunL1NodeServiceName())
			if err != nil {
				return err
			}
			return nil
		},
	}

	return startCmd
}

func initiaLogCommand() *cobra.Command {
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Stream the logs of the initiad full node application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if the OS is Linux
			switch runtime.GOOS {
			case "linux":
				return streamLogsFromJournalctl()
			case "darwin":
				// If not Linux, fall back to file-based log streaming
				return streamLogsFromFiles()
			default:
				return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
			}
		},
	}

	return logCmd
}

// streamLogsFromJournalctl uses journalctl to stream logs from initia.service
func streamLogsFromJournalctl() error {
	// Execute the journalctl command to follow logs of initia.service
	cmd := exec.Command("journalctl", "-f", "-u", "initia.service")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command and return any errors
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stream logs using journalctl: %v", err)
	}

	return nil
}

// streamLogsFromFiles streams logs from file-based logs
func streamLogsFromFiles() error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	logFilePathOut := filepath.Join(userHome, utils.WeaveLogDirectory, "initia.stdout.log")
	logFilePathErr := filepath.Join(userHome, utils.WeaveLogDirectory, "initia.stderr.log")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go tailLogFile(logFilePathOut, os.Stdout)
	go tailLogFile(logFilePathErr, os.Stderr)

	<-sigChan

	fmt.Println("Stopping log streaming...")
	return nil
}

func tailLogFile(filePath string, output io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error opening log file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > PreviousLogLines {
			lines = lines[1:]
		}
	}

	for _, line := range lines {
		fmt.Fprintln(output, line)
	}

	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		fmt.Printf("error seeking to end of log file %s: %v\n", filePath, err)
		return
	}

	for {
		var line = make([]byte, 4096)
		n, err := file.Read(line)
		if err != nil && err != io.EOF {
			fmt.Printf("error reading log file %s: %v\n", filePath, err)
			return
		}

		if n > 0 {
			output.Write(line[:n])
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}
