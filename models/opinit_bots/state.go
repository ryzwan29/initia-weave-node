package opinit_bots

import "github.com/initia-labs/weave/types"

type OPInitBotsState struct {
	BotInfos             []BotInfo
	SetupOpinitResponses []string
	OPInitBotVersion     string
	OPInitBotEndpoint    string
	weave                types.WeaveState
}

// Function to initialize OPInitBotsState with all bots in default setup state (false)
func NewOPInitBotsState() *OPInitBotsState {
	return &OPInitBotsState{
		BotInfos:             CheckIfKeysExist(BotInfos),
		SetupOpinitResponses: make([]string, 0),
		weave:                types.NewWeaveState(),
	}
}
