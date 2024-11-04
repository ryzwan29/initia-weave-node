package opinit_bots

import (
	"github.com/initia-labs/weave/types"
)

// OPInitBotsState is the structure holding the bot state and related configurations
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
	daIsCelestia         bool
	dbPath               string
	isDeleteDB           bool
}

// NewOPInitBotsState initializes OPInitBotsState with default values
func NewOPInitBotsState() OPInitBotsState {
	return OPInitBotsState{
		BotInfos:             CheckIfKeysExist(BotInfos),
		SetupOpinitResponses: make(map[BotName]string),
		weave:                types.NewWeaveState(),
		MinitiaConfig:        nil,
		botConfig:            make(map[string]string),
	}
}

// Clone creates a deep copy of the OPInitBotsState to ensure state independence
func (state OPInitBotsState) Clone() OPInitBotsState {
	clone := OPInitBotsState{
		BotInfos:             make([]BotInfo, len(state.BotInfos)),
		SetupOpinitResponses: make(map[BotName]string),
		OPInitBotVersion:     state.OPInitBotVersion,
		OPInitBotEndpoint:    state.OPInitBotEndpoint,
		MinitiaConfig:        state.MinitiaConfig, // Assuming this can be reused or cloned if necessary
		weave:                state.weave,         // Assuming weave can be reused or cloned if necessary
		InitExecutorBot:      state.InitExecutorBot,
		InitChallengerBot:    state.InitChallengerBot,
		ReplaceBotConfig:     state.ReplaceBotConfig,
		Version:              state.Version,
		ListenAddress:        state.ListenAddress,
		L1ChainId:            state.L1ChainId,
		L1RPCAddress:         state.L1RPCAddress,
		L1GasPrice:           state.L1GasPrice,
		botConfig:            make(map[string]string),
		dbPath:               state.dbPath,
		isDeleteDB:           state.isDeleteDB,
	}

	// Copy slice data
	copy(clone.BotInfos, state.BotInfos)

	// Copy map data
	for k, v := range state.SetupOpinitResponses {
		clone.SetupOpinitResponses[k] = v
	}
	for k, v := range state.botConfig {
		clone.botConfig[k] = v
	}

	return clone
}
