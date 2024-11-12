package service

type Template string

// DarwinRunBinaryTemplate should inject the arguments as follows: [binaryName, binaryPath, appHome, userHome, weaveLogPath, serviceName]
const DarwinRunBinaryTemplate Template = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.%[1]s.daemon</string>

    <key>ProgramArguments</key>
    <array>
        <string>%[2]s/%[1]s</string>
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
        <string>%[4]s</string>
        <key>DYLD_LIBRARY_PATH</key>
        <string>%[2]s</string>
    </dict>

    <key>StandardOutPath</key>
    <string>%[5]s/%[1]s.stdout.log</string>

    <key>StandardErrorPath</key>
    <string>%[5]s/%[1]s.stderr.log</string>

    <key>HardResourceLimits</key>
    <dict>
        <key>NumberOfFiles</key>
        <integer>65535</integer>
    </dict>
</dict>
</plist>
`

// DarwinOPinitBotTemplate should inject the arguments as follows: [binaryName, binaryPath, appHome, userHome, weaveLogPath, serviceName]
const DarwinOPinitBotTemplate Template = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.%[1]s.%[6]s.daemon</string>

    <key>ProgramArguments</key>
    <array>
        <string>%[2]s/%[1]s</string>
        <string>start</string>
		<string>%[6]s</string>
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
        <string>%[4]s</string>
    </dict>

    <key>StandardOutPath</key>
    <string>%[5]s/%[1]s.%[6]s.stdout.log</string>

    <key>StandardErrorPath</key>
    <string>%[5]s/%[1]s.%[6]s.stderr.log</string>

    <key>HardResourceLimits</key>
    <dict>
        <key>NumberOfFiles</key>
        <integer>65535</integer>
    </dict>
</dict>
</plist>
`

// LinuxRunBinaryTemplate should inject the arguments as follows: [binaryName, currentUser.Username, binaryPath, serviceName, appHome]
const LinuxRunBinaryTemplate Template = `
[Unit]
Description=%[1]s
After=network.target

[Service]
Type=exec
User=%[2]s
ExecStart=%[3]s/%[1]s start --home %[5]s
KillSignal=SIGINT
Environment="LD_LIBRARY_PATH=%[3]s"

[Install]
WantedBy=multi-user.target

[Service]
LimitNOFILE=65535
`

// LinuxOPinitBotTemplate should inject the arguments as follows: [binaryName, currentUser.Username, binaryPath, serviceName, appHome]
const LinuxOPinitBotTemplate Template = `
[Unit]
Description=%[1]s %[4]s
After=network.target

[Service]
Type=exec
User=%[2]s
ExecStart=%[3]s/%[1]s start %[4]s --home %[5]s
KillSignal=SIGINT
Environment="LD_LIBRARY_PATH=%[3]s"

[Install]
WantedBy=multi-user.target

[Service]
LimitNOFILE=65535
`

var (
	LinuxTemplateMap = map[CommandName]Template{
		Initia:           LinuxRunBinaryTemplate,
		Minitia:          LinuxRunBinaryTemplate,
		OPinitExecutor:   LinuxOPinitBotTemplate,
		OPinitChallenger: LinuxOPinitBotTemplate,
	}
	DarwinTemplateMap = map[CommandName]Template{
		Initia:           DarwinRunBinaryTemplate,
		Minitia:          DarwinRunBinaryTemplate,
		OPinitExecutor:   DarwinOPinitBotTemplate,
		OPinitChallenger: DarwinOPinitBotTemplate,
	}
)
