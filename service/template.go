package service

type Template string

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

const LinuxRunBinaryTemplate Template = `
[Unit]
Description=%[1]s
After=network.target

[Service]
Type=exec
User=%[2]s
ExecStart=%[3]s/%[1]s start
KillSignal=SIGINT
Environment="LD_LIBRARY_PATH=%[3]s"

[Install]
WantedBy=multi-user.target

[Service]
LimitNOFILE=65535
`

var (
	LinuxTemplateMap = map[CommandName]Template{
		Initia:  LinuxRunBinaryTemplate,
		Minitia: LinuxRunBinaryTemplate,
	}
	DarwinTemplateMap = map[CommandName]Template{
		Initia:  DarwinRunBinaryTemplate,
		Minitia: DarwinRunBinaryTemplate,
	}
)
