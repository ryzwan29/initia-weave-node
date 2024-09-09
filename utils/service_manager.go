package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func CreateService(serviceName, serviceContent string) error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("sudo", "tee", fmt.Sprintf("/etc/systemd/system/%s.service", serviceName))
		cmd.Stdin = strings.NewReader(serviceContent)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create service: %v", err)
		}
		if err := DaemonReload(); err != nil {
			return err
		}
		return EnableService(serviceName)
	case "darwin":
		cmd := exec.Command("sudo", "tee", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName))
		cmd.Stdin = strings.NewReader(serviceContent)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create service: %v", err)
		}
		return LoadService(serviceName)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func DaemonReload() error {
	cmd := exec.Command("sudo", "systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd daemon: %v", err)
	}
	return nil
}

func EnableService(serviceName string) error {
	cmd := exec.Command("sudo", "systemctl", "enable", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}
	return nil
}

func LoadService(serviceName string) error {
	cmd := exec.Command("sudo", "launchctl", "load", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to load service: %v", err)
	}
	return nil
}

func StartService(serviceName string) error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("sudo", "systemctl", "start", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start service: %v", err)
		}
		return nil
	case "darwin":
		cmd := exec.Command("sudo", "launchctl", "load", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start service: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func StopService(serviceName string) error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("sudo", "systemctl", "stop", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop service: %v", err)
		}
		return nil
	case "darwin":
		cmd := exec.Command("sudo", "launchctl", "unload", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop service: %v", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}
