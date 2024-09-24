package utils

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
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
    <false/>

    <key>KeepAlive</key>
    <false/>

    <!-- Adding the environment variable -->
    <key>EnvironmentVariables</key>
    <dict>
		<key>HOME</key>
        <string>%[2]s</string>
        <key>DYLD_LIBRARY_PATH</key>
        <string>%[1]s</string>
    </dict>

    <key>StandardOutPath</key>
    <string>%[4]s/initia.stdout.log</string>

    <key>StandardErrorPath</key>
    <string>%[4]s/initia.stderr.log</string>

    <key>HardResourceLimits</key>
    <dict>
        <key>NumberOfFiles</key>
        <integer>65535</integer>
    </dict>
</dict>
</plist>
`

const LinuxRunL1NodeTemplate = `
[Unit]
Description=Initia Daemon
After=network.target

[Service]
Type=exec
User=%[1]s
ExecStart=%[2]s/initiad start
KillSignal=SIGINT
Environment="LD_LIBRARY_PATH=%[2]s"

[Install]
WantedBy=multi-user.target

[Service]
LimitNOFILE=65535
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
		userHome, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %v", err)
		}
		cmd := exec.Command("tee", filepath.Join(userHome, fmt.Sprintf("Library/LaunchAgents/%s.plist", serviceName)))
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
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}
	loadCmd := exec.Command("launchctl", "load", filepath.Join(userHome, fmt.Sprintf("Library/LaunchAgents/%s.plist", serviceName)))
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
		startCmd := exec.Command("launchctl", "start", serviceName)
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
		cmd := exec.Command("launchctl", "stop", serviceName)
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

func GetRunL1NodeServiceContent(version string) string {
	switch runtime.GOOS {
	case "linux":
		return GetLinuxRunL1NodeServiceContent(version)
	case "darwin":
		return GetDarwinRunL1NodePlist(version)
	default:
		panic(fmt.Errorf("unsupported operating system: %s", runtime.GOOS))
	}
}

func GetLinuxRunL1NodeServiceContent(version string) string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get user home directory: %v", err))
	}

	currentUser, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("failed to get current user: %v", err))
	}

	weaveDataPath := filepath.Join(userHome, WeaveDataDirectory)
	binaryPath := filepath.Join(weaveDataPath, "initia@"+version, "initia_"+version)

	return fmt.Sprintf(LinuxRunL1NodeTemplate, currentUser.Username, binaryPath)
}

func GetDarwinRunL1NodePlist(version string) string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get user home directory: %v", err))
	}
	weaveDataPath := filepath.Join(userHome, WeaveDataDirectory)
	weaveLogPath := filepath.Join(userHome, WeaveLogDirectory)
	binaryPath := filepath.Join(weaveDataPath, "initia@"+version)
	initiaHome := filepath.Join(userHome, InitiaDirectory)
	if err = os.Setenv("DYLD_LIBRARY_PATH", binaryPath); err != nil {
		panic(fmt.Errorf("failed to set DYLD_LIBRARY_PATH: %v", err))
	}
	if err = os.Setenv("HOME", userHome); err != nil {
		panic(fmt.Errorf("failed to set HOME: %v", err))
	}

	return fmt.Sprintf(DarwinRunL1NodeTemplate, binaryPath, userHome, initiaHome, weaveLogPath)
}
