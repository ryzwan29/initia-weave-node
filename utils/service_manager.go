package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const DarwinRunL1NodeTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.initia.daemon</string>

    <key>ProgramArguments</key>
    <array>
        <string>%[1]s/initiad</string>
        <string>start</string>
		<string>--home=%[3]s</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>

    <!-- Adding the environment variable -->
    <key>EnvironmentVariables</key>
    <dict>
		<key>HOME</key>
        <string>%[2]s</string>
        <key>DYLD_LIBRARY_PATH</key>
        <string>%[1]s</string>
    </dict>

    <key>StandardOutPath</key>
    <string>/tmp/initia.stdout.log</string>

    <key>StandardErrorPath</key>
    <string>/tmp/initia.stderr.log</string>

    <key>HardResourceLimits</key>
    <dict>
        <key>NumberOfFiles</key>
        <integer>65535</integer>
    </dict>
</dict>
</plist>
`

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
	loadCmd := exec.Command("sudo", "launchctl", "load", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName))
	if err := loadCmd.Run(); err != nil {
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
		loadCmd := exec.Command("sudo", "launchctl", "load", fmt.Sprintf("/Library/LaunchDaemons/%s.plist", serviceName))
		if err := loadCmd.Run(); err != nil {
			return fmt.Errorf("failed to load service: %v", err)
		}

		startCmd := exec.Command("sudo", "launchctl", "start", serviceName)
		if err := startCmd.Run(); err != nil {
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

func GetRunL1NodeServiceName() string {
	switch runtime.GOOS {
	case "linux":
		return RunL1NodeLinuxServiceName
	case "darwin":
		return RunL1NodeDarwinServiceName
	default:
		panic(fmt.Errorf("unsupported operating system: %s", runtime.GOOS))
	}
}

func GetDarwinRunL1NodePlist(version string) string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get user home directory: %v", err))
	}
	weaveDataPath := filepath.Join(userHome, WeaveDataDirectory)
	binaryPath := filepath.Join(weaveDataPath, "initia@"+version)
	initiaHome := filepath.Join(userHome, InitiaDirectory)
	if err = os.Setenv("DYLD_LIBRARY_PATH", binaryPath); err != nil {
		panic(fmt.Errorf("failed to set DYLD_LIBRARY_PATH: %v", err))
	}
	if err = os.Setenv("HOME", userHome); err != nil {
		panic(fmt.Errorf("failed to set HOME: %v", err))
	}

	return fmt.Sprintf(DarwinRunL1NodeTemplate, binaryPath, userHome, initiaHome)
}
