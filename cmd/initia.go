package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/initia-labs/weave/models"
	"github.com/initia-labs/weave/models/weaveinit"
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
				_, err := tea.NewProgram(models.NewExistingAppChecker()).Run()
				if err != nil {
					return err
				}
			}

			_, err := tea.NewProgram(weaveinit.NewRunL1NodeNetworkSelect(weaveinit.NewRunL1NodeState())).Run()
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
		},
	}

	return logCmd
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
