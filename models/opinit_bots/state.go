package opinit_bots

import "github.com/initia-labs/weave/types"

type OPInitBotsState struct {
	BotInfos             []BotInfo
	SetupOpinitResponses map[BotName]string
	OPInitBotVersion     string
	OPInitBotEndpoint    string
	MinitiaConfig        *types.MinitiaConfig
	weave                types.WeaveState
	InitExecutorBot      bool
	InitChallengerBot    bool
	ReplaceBotConfig     bool
	Version              string
	ListenAddress        string
	L1ChainId            string
	L1RPCAddress         string
	L1GasPrice           string
	botConfig            map[string]string
	dbPath               string
	isDeleteDB           bool
}

// NewOPInitBotsState is a function to initialize OPInitBotsState with all bots in default setup state (false)
func NewOPInitBotsState() *OPInitBotsState {
	return &OPInitBotsState{
		BotInfos:             CheckIfKeysExist(BotInfos),
		SetupOpinitResponses: make(map[BotName]string),
		weave:                types.NewWeaveState(),
		MinitiaConfig:        nil,
		botConfig:            make(map[string]string),
	}
}
