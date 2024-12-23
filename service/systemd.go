package service

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/initia-labs/weave/common"
)

type Systemd struct {
	commandName CommandName
}

func NewSystemd(commandName CommandName) *Systemd {
	return &Systemd{commandName: commandName}
}

func (j *Systemd) GetCommandName() string {
	return string(j.commandName)
}

func (j *Systemd) GetServiceName() (string, error) {
	slug, err := j.commandName.GetServiceSlug()
	if err != nil {
		return "", fmt.Errorf("failed to get service name: %v", err)
	}
	return slug + ".service", nil
}

func (j *Systemd) Create(binaryVersion, appHome string) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	binaryName, err := j.commandName.GetBinaryName()
	if err != nil {
		return fmt.Errorf("failed to get current binary name: %v", err)
	}
	var binaryPath string
	switch j.commandName {
	case UpgradableInitia, NonUpgradableInitia:
		binaryPath = filepath.Join(userHome, common.WeaveDataDirectory, binaryVersion)
	case Minitia:
		binaryPath = filepath.Join(userHome, common.WeaveDataDirectory, binaryVersion, strings.ReplaceAll(binaryVersion, "@", "_"))
	default:
		binaryPath = filepath.Join(userHome, common.WeaveDataDirectory)
	}

	serviceName, err := j.GetServiceName()
	if err != nil {
		return err
	}
	cmd := exec.Command("sudo", "tee", fmt.Sprintf("/etc/systemd/system/%s", serviceName))
	template := LinuxTemplateMap[j.commandName]
	cmd.Stdin = strings.NewReader(fmt.Sprintf(string(template), binaryName, currentUser.Username, binaryPath, string(j.commandName), appHome))
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}
	if err = j.daemonReload(); err != nil {
		return err
	}
	return j.enableService()
}

func (j *Systemd) daemonReload() error {
	cmd := exec.Command("sudo", "systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %v", err)
	}
	return nil
}

func (j *Systemd) enableService() error {
	serviceName, err := j.GetServiceName()
	if err != nil {
		return err
	}
	cmd := exec.Command("sudo", "systemctl", "enable", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}
	return nil
}

func (j *Systemd) Log(n int) error {
	serviceName, err := j.GetServiceName()
	if err != nil {
		return err
	}
	fmt.Printf("Streaming logs from systemd %s\n", serviceName)

	cmd := exec.Command("journalctl", "-f", "-u", serviceName, "-n", strconv.Itoa(n))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (j *Systemd) Start() error {
	serviceName, err := j.GetServiceName()
	if err != nil {
		return err
	}
	cmd := exec.Command("systemctl", "start", serviceName)
	return cmd.Run()
}

func (j *Systemd) Stop() error {
	serviceName, err := j.GetServiceName()
	if err != nil {
		return err
	}
	cmd := exec.Command("systemctl", "stop", serviceName)
	return cmd.Run()
}

func (j *Systemd) Restart() error {
	serviceName, err := j.GetServiceName()
	if err != nil {
		return err
	}
	cmd := exec.Command("systemctl", "restart", serviceName)
	return cmd.Run()
}
