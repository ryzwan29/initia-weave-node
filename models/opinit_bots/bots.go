package opinit_bots

import (
	"os/exec"
	"strings"
)

const (
	OpinitBotBinaryVersion = "v0.1.10"
)

// BotName defines a custom type for the bot names
type BotName string

// Create constants for the BotNames
const (
	BridgeExecutor       BotName = "Bridge Executor"
	OutputSubmitter      BotName = "Output Submitter"
	BatchSubmitter       BotName = "Batch Submitter"
	Challenger           BotName = "Challenger"
	OracleBridgeExecutor BotName = "Oracle Bridge Executor"
)

// BotKeyName defines a custom type for bot key names
type BotKeyName string

// Create constants for the bot key names
const (
	BridgeExecutorKeyName       = "weave_bridge_executor"
	OutputSubmitterKeyName      = "weave_output_submitter"
	BatchSubmitterKeyName       = "weave_batch_submitter"
	ChallengerKeyName           = "weave_challenger"
	OracleBridgeExecutorKeyName = "weave_oracle_bridge_executor"
)

// BotNames to hold all bot names
var BotNames = []BotName{
	BridgeExecutor,
	OutputSubmitter,
	BatchSubmitter,
	Challenger,
	OracleBridgeExecutor,
}

// BotInfo struct to hold all relevant bot information
type BotInfo struct {
	BotName       BotName
	IsSetup       bool
	KeyName       string
	Mnemonic      string
	IsNotExist    bool // Indicates if the key doesn't exist in the `initiad keys list` output
	IsGenerateKey bool
	DALayer       string
}

// BotInfos for all bots with key names filled in
var BotInfos = []BotInfo{
	{
		BotName:    BridgeExecutor,
		IsSetup:    false, // Default isn't set up
		KeyName:    BridgeExecutorKeyName,
		Mnemonic:   "", // Add mnemonic if needed
		IsNotExist: false,
	},
	{
		BotName:    OutputSubmitter,
		IsSetup:    false, // Default isn't set up
		KeyName:    OutputSubmitterKeyName,
		Mnemonic:   "", // Add mnemonic if needed
		IsNotExist: false,
	},
	{
		BotName:    BatchSubmitter,
		IsSetup:    false, // Default isn't set up
		KeyName:    BatchSubmitterKeyName,
		Mnemonic:   "", // Add mnemonic if needed
		IsNotExist: false,
	},
	{
		BotName:    Challenger,
		IsSetup:    false, // Default isn't set up
		KeyName:    ChallengerKeyName,
		Mnemonic:   "", // Add mnemonic if needed
		IsNotExist: false,
	},
	{
		BotName:    OracleBridgeExecutor,
		IsSetup:    false, // Default isn't set up
		KeyName:    OracleBridgeExecutorKeyName,
		Mnemonic:   "", // Add mnemonic if needed
		IsNotExist: false,
	},
}

// CheckIfKeysExist checks the output of `initiad keys list` and sets IsNotExist for missing keys
func CheckIfKeysExist(botInfos []BotInfo) []BotInfo {
	cmd := exec.Command(AppName, "keys", "list", "weave-dummy")
	outputBytes, err := cmd.Output()
	if err != nil {
		return botInfos
	}
	output := string(outputBytes)

	// Split the output by line and check if the KeyName exists
	for i := range botInfos {
		if !strings.Contains(output, botInfos[i].KeyName) {
			botInfos[i].IsNotExist = true
		}
	}
	return botInfos
}
