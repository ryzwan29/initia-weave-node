package opinit_bots

import (
	"os/exec"
	"strings"
)

// Define a custom type for the bot names
type BotName string

// Create constants for the BotNames
const (
	BridgeExecutor  BotName = "bridge executor"
	OutputSubmitter BotName = "output submitter"
	BatchSubmitter  BotName = "batch submitter"
	Challenger      BotName = "challenger"
)

// Define a custom type for bot key names
type BotKeyName string

// Create constants for the bot key names
const (
	BridgeExecutorKeyName  = "weave-bridge-executor"
	OutputSubmitterKeyName = "weave-output-submitter"
	BatchSubmitterKeyName  = "weave-batch-submitter"
	ChallengerKeyName      = "weave-challenger"
)

// Create a slice of BotName to hold all bot names
var BotNames = []BotName{
	BridgeExecutor,
	OutputSubmitter,
	BatchSubmitter,
	Challenger,
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

// Create a slice of BotInfo for all bots with key names filled in
var BotInfos = []BotInfo{
	{
		BotName:    BridgeExecutor,
		IsSetup:    false, // Default not set up
		KeyName:    BridgeExecutorKeyName,
		Mnemonic:   "", // Add mnemonic if needed
		IsNotExist: false,
	},
	{
		BotName:    OutputSubmitter,
		IsSetup:    false, // Default not set up
		KeyName:    OutputSubmitterKeyName,
		Mnemonic:   "", // Add mnemonic if needed
		IsNotExist: false,
	},
	{
		BotName:    BatchSubmitter,
		IsSetup:    false, // Default not set up
		KeyName:    BatchSubmitterKeyName,
		Mnemonic:   "", // Add mnemonic if needed
		IsNotExist: false,
	},
	{
		BotName:    Challenger,
		IsSetup:    false, // Default not set up
		KeyName:    ChallengerKeyName,
		Mnemonic:   "", // Add mnemonic if needed
		IsNotExist: false,
	},
}

// CheckIfKeysExist checks the output of `initiad keys list` and sets IsNotExist for missing keys
func CheckIfKeysExist(botInfos []BotInfo) []BotInfo {
	// Run the `initiad keys list --keyring-backend test` command
	cmd := exec.Command("initiad", "keys", "list", "--keyring-backend", "test")
	outputBytes, err := cmd.Output()
	if err != nil {
		panic("Error running `initiad keys list --keyring-backend test`: " + err.Error())
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
