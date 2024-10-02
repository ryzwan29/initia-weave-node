package opinit_bots

type OPInitBotsState struct {
	BotInfos             []BotInfo
	SetupOpinitResponses []string
}

// Function to initialize OPInitBotsState with all bots in default setup state (false)
func NewOPInitBotsState() *OPInitBotsState {
	return &OPInitBotsState{
		BotInfos:             CheckIfKeysExist(BotInfos),
		SetupOpinitResponses: make([]string, 0),
	}
}
