package opinit_bots

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/types"
)

// PageStatePair represents a pair of the page (model) and its associated state (OPInitBotsState)
type PageStatePair struct {
	Page  tea.Model
	State *OPInitBotsState
}

// AppState wraps around OPInitBotsState and manages page-state navigation
type AppState struct {
	currentModel tea.Model        // Current active model (page)
	currentState *OPInitBotsState // Current active state (OPInitBotsState)
	pageStack    []PageStatePair  // Stack to store page and state pairs
}

// NewAppState initializes AppState with a new OPInitBotsState and an empty page stack
func NewAppState() *AppState {
	return &AppState{
		currentState: NewOPInitBotsState(),
	}
}

func (state *AppState) SetCurrentModel(currentModel tea.Model) {
	state.currentModel = currentModel
}

func (state *AppState) GetCurrentModel() tea.Model {
	return state.currentModel
}

// PushPageState pushes the current page and a cloned version of its state onto the stack
func (state *AppState) PushPageState(page tea.Model, opState *OPInitBotsState) {
	clonedState := opState.Clone() // Clone the state before pushing it onto the stack
	state.pageStack = append(state.pageStack, PageStatePair{Page: page, State: clonedState})
}

// PopPageState pops the last page-state pair from the stack and returns it
func (state *AppState) PopPageState() *PageStatePair {
	if len(state.pageStack) == 0 {
		return nil
	}
	lastPair := state.pageStack[len(state.pageStack)-1]
	state.pageStack = state.pageStack[:len(state.pageStack)-1]
	return &lastPair
}

// GoBack restores the last page and state pair from the stack
func (state *AppState) GoBack() tea.Model {
	prevPair := state.PopPageState()
	if prevPair != nil {
		state.currentModel = prevPair.Page
		state.currentState = prevPair.State // Restore the cloned state
	}
	return state.currentModel
}

// Clone creates a deep copy of the OPInitBotsState to ensure state independence
func (state *OPInitBotsState) Clone() *OPInitBotsState {
	// Create a new instance
	clone := &OPInitBotsState{
		BotInfos:             make([]BotInfo, len(state.BotInfos)),
		SetupOpinitResponses: make(map[BotName]string),
		OPInitBotVersion:     state.OPInitBotVersion,
		OPInitBotEndpoint:    state.OPInitBotEndpoint,
		MinitiaConfig:        state.MinitiaConfig, // Assuming MinitiaConfig can be reused or cloned if needed
		weave:                state.weave,         // Assuming weave can be reused or cloned if needed
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

func HandleCmdZ(state *AppState, msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Detect Cmd+Z as Alt+Z (this may be handled as Alt/Option key in the terminal)
		if keyMsg.String() == "ctrl+z" {
			// Undo: Go back to the previous page and state
			return state.GoBack(), nil, true
		}
	}
	return nil, nil, false
}

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
