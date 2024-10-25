package service

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/initia-labs/weave/utils"
)

type Launchd struct {
	commandName CommandName
}

func NewLaunchd(commandName CommandName) *Launchd {
	return &Launchd{commandName: commandName}
}

func (j *Launchd) GetCommandName() string {
	return string(j.commandName)
}

func (j *Launchd) GetServiceName() string {
	return fmt.Sprintf("com.%s.daemon", j.commandName.MustGetServiceSlug())
}

func (j *Launchd) GetAppHome() string {
	switch j.commandName {
	case Initia:
		return utils.InitiaDirectory
	case Minitia:
		return utils.MinitiaDirectory
	case OPinitExecutor:
		return utils.OPinitDirectory
	case OPinitChallenger:
		return utils.OPinitDirectory
	}
	panic("unsupported app")
}

func (j *Launchd) Create(binaryVersion string) error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	weaveDataPath := filepath.Join(userHome, utils.WeaveDataDirectory)
	weaveLogPath := filepath.Join(userHome, utils.WeaveLogDirectory)
	binaryName := j.commandName.MustGetBinaryName()
	binaryPath := filepath.Join(weaveDataPath, binaryVersion)
	appHome := filepath.Join(userHome, j.GetAppHome())
	if err = os.Setenv("DYLD_LIBRARY_PATH", binaryPath); err != nil {
		panic(fmt.Errorf("failed to set DYLD_LIBRARY_PATH: %v", err))
	}
	if err = os.Setenv("HOME", userHome); err != nil {
		panic(fmt.Errorf("failed to set HOME: %v", err))
	}

	plistPath := filepath.Join(userHome, fmt.Sprintf("Library/LaunchAgents/%s.plist", j.GetServiceName()))
	if utils.FileOrFolderExists(plistPath) {
		err = utils.DeleteFile(plistPath)
		if err != nil {
			panic(err)
		}
	}
	cmd := exec.Command("tee", plistPath)
	template := DarwinTemplateMap[j.commandName]
	cmd.Stdin = strings.NewReader(fmt.Sprintf(string(template), binaryName, binaryPath, appHome, userHome, weaveLogPath, j.GetCommandName()))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}
	return j.reloadService()
}

func (j *Launchd) unloadService() error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	unloadCmd := exec.Command("launchctl", "unload", filepath.Join(userHome, fmt.Sprintf("Library/LaunchAgents/%s.plist", j.GetServiceName())))
	if err = unloadCmd.Run(); err != nil {
		return fmt.Errorf("failed to unload service: %v", err)
	}
	return nil
}

func (j *Launchd) reloadService() error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	unloadCmd := exec.Command("launchctl", "unload", filepath.Join(userHome, fmt.Sprintf("Library/LaunchAgents/%s.plist", j.GetServiceName())))
	_ = unloadCmd.Run()
	loadCmd := exec.Command("launchctl", "load", filepath.Join(userHome, fmt.Sprintf("Library/LaunchAgents/%s.plist", j.GetServiceName())))
	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("failed to load service: %v", err)
	}
	return nil
}

func (j *Launchd) Start() error {
	cmd := exec.Command("launchctl", "start", j.GetServiceName())
	return cmd.Run()
}

func (j *Launchd) Stop() error {
	cmd := exec.Command("launchctl", "stop", j.GetServiceName())
	return cmd.Run()
}

func (j *Launchd) Restart() error {
	err := j.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop service: %v", err)
	}
	err = j.Start()
	if err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}
	return nil
}

func (j *Launchd) Log(n int) error {
	fmt.Printf("Streaming logs from launchd %s\n", j.GetServiceName())
	return j.streamLogsFromFiles(n)
}

// streamLogsFromFiles streams logs from file-based logs
func (j *Launchd) streamLogsFromFiles(n int) error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	logFilePathOut := filepath.Join(userHome, utils.WeaveLogDirectory, fmt.Sprintf("%s.stdout.log", j.commandName.MustGetServiceSlug()))
	logFilePathErr := filepath.Join(userHome, utils.WeaveLogDirectory, fmt.Sprintf("%s.stderr.log", j.commandName.MustGetServiceSlug()))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go j.tailLogFile(logFilePathOut, os.Stdout, n)
	go j.tailLogFile(logFilePathErr, os.Stderr, n)

	<-sigChan

	fmt.Println("Stopping log streaming...")
	return nil
}

func (j *Launchd) tailLogFile(filePath string, output io.Writer, maxLogLines int) {
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
		if len(lines) > maxLogLines {
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
