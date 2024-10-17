package opinit_bots

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

type OPInitBotInitOption string

const (
	Executor_OPInitBotInitOption   OPInitBotInitOption = "Executor"
	Challenger_OPInitBotInitOption OPInitBotInitOption = "Challenger"
)

type OPInitBotInitSelector struct {
	utils.Selector[OPInitBotInitOption]
	state    *OPInitBotsState
	question string
}

var defaultExecutorFields = []Field{
	// Version
	{Name: "version", Type: NumberField, Question: "Please specify the version", Placeholder: `Press tab to use "1"`, DefaultValue: "1"},

	// Listen Address
	{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address", Placeholder: `Add listen address ex. localhost:3000`},

	// L1 Node Configuration
	{Name: "l1_node.chain_id", Type: StringField, Question: "Please specify the L1 chain_id", Placeholder: "Add alphanumeric"},
	{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address", Placeholder: "Add RPC address ex. tcp://localhost:26657"},
	{Name: "l1_node.gas_price", Type: StringField, Question: "Please specify the L1 gas_price", Placeholder: `Press tab to use "0.15uinit"`, DefaultValue: "0.15uinit"},

	// L2 Node Configuration
	{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric"},
	{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. tcp://localhost:26657"},
	{Name: "l2_node.gas_price", Type: StringField, Question: "Please specify the L2 gas_price", Placeholder: `Press tab to use "0.15uinit"`, DefaultValue: "0.15uinit"},
}

var defaultDALayerFields = []Field{
	// Version
	{Name: "l2_start_height", Type: NumberField, Question: "Please specify the l2_start_height"},

	// Listen Address
	{Name: "batch_start_height", Type: NumberField, Question: "Please specify the batch_start_height"},
}

var defaultChallengerFields = []Field{
	// Version
	{Name: "version", Type: NumberField, Question: "Please specify the version", Placeholder: `Press tab to use "1"`, DefaultValue: "1"},

	// Listen Address
	{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address", Placeholder: `Add listen address ex. localhost:3000`},

	// L1 Node Configuration
	{Name: "l1_node.chain_id", Type: StringField, Question: "Please specify the L1 chain_id", Placeholder: "Add alphanumeric"},
	{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address", Placeholder: "Add RPC address ex. tcp://localhost:26657"},

	// L2 Node Configuration
	{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric"},
	{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. tcp://localhost:26657"},

	// Version
	{Name: "l2_start_height", Type: NumberField, Question: "Please specify the l2_start_height"},
}

func NewOPInitBotInitSelector(state *OPInitBotsState) tea.Model {
	return &OPInitBotInitSelector{
		Selector: utils.Selector[OPInitBotInitOption]{
			Options: []OPInitBotInitOption{Executor_OPInitBotInitOption, Challenger_OPInitBotInitOption},
		},
		state:    state,
		question: "Which bot would you like to run?",
	}
}

func (m *OPInitBotInitSelector) GetQuestion() string {
	return m.question
}

func (m *OPInitBotInitSelector) Init() tea.Cmd {
	return nil
}

func (m *OPInitBotInitSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bot"}, string(*selected)))
		homeDir, _ := os.UserHomeDir()

		switch *selected {
		case Executor_OPInitBotInitOption:
			m.state.InitExecutorBot = true

			m.state.dbPath = filepath.Join(homeDir, utils.OPinitDirectory, "executor.db")
			if utils.FileOrFolderExists(m.state.dbPath) {
				return NewDeleteDBSelector(m.state, "executor"), cmd
			}

			return NewUseCurrentCofigSelector(m.state, "executor"), cmd
		case Challenger_OPInitBotInitOption:
			m.state.InitChallengerBot = true

			m.state.dbPath = filepath.Join(homeDir, utils.OPinitDirectory, "challenger.db")
			if utils.FileOrFolderExists(m.state.dbPath) {
				return NewDeleteDBSelector(m.state, "challenger"), cmd
			}
			return NewUseCurrentCofigSelector(m.state, "challenger"), cmd

		}
	}

	return m, cmd
}

func (m *OPInitBotInitSelector) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{"bot"}, styles.Question) + m.Selector.View()
}

type DeleteDBOption string

const (
	DeleteDBOption_No  = "No"
	DeleteDBOption_Yes = "Yes, reset"
)

type DeleteDBSelector struct {
	utils.Selector[DeleteDBOption]
	state    *OPInitBotsState
	question string
	bot      string
}

func NewDeleteDBSelector(state *OPInitBotsState, bot string) *DeleteDBSelector {
	return &DeleteDBSelector{
		Selector: utils.Selector[DeleteDBOption]{
			Options: []DeleteDBOption{
				DeleteDBOption_No,
				DeleteDBOption_Yes,
			},
		},
		state:    state,
		question: "Would you like to reset the database?",
		bot:      bot,
	}
}

func (m *DeleteDBSelector) GetQuestion() string {
	return m.question
}

func (m *DeleteDBSelector) Init() tea.Cmd {
	return nil
}

func (m *DeleteDBSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, string(*selected)))

		switch *selected {
		case DeleteDBOption_No:
			m.state.isDeleteDB = false
		case DeleteDBOption_Yes:
			m.state.isDeleteDB = true
		}
		return NewUseCurrentCofigSelector(m.state, m.bot), cmd

	}

	return m, cmd
}

func (m *DeleteDBSelector) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + m.Selector.View()
}

type UseCurrentCofigSelector struct {
	utils.Selector[string]
	state      *OPInitBotsState
	question   string
	configPath string
}

func NewUseCurrentCofigSelector(state *OPInitBotsState, bot string) *UseCurrentCofigSelector {
	configPath := fmt.Sprintf(".opinit/%s.json", bot)
	return &UseCurrentCofigSelector{
		Selector: utils.Selector[string]{
			Options: []string{
				"use current file",
				"replace",
			},
		},
		state:      state,
		question:   fmt.Sprintf("Existing %s detected. Would you like to use the current one or replace it?", configPath),
		configPath: configPath,
	}
}

func (m *UseCurrentCofigSelector) GetQuestion() string {
	return m.question
}

func (m *UseCurrentCofigSelector) Init() tea.Cmd {
	return nil
}

func (m *UseCurrentCofigSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{m.configPath}, string(*selected)))

		switch *selected {
		case "use current file":
			m.state.ReplaceBotConfig = false
		case "replace":
			m.state.ReplaceBotConfig = true
			// TODO: load config
			if m.state.InitExecutorBot {
				return NewFieldInputModel(m.state, defaultExecutorFields, NewSetDALayer), cmd
			} else if m.state.InitChallengerBot {
				return NewFieldInputModel(m.state, defaultChallengerFields, NewStartingInitBot), cmd
			}

			return m, cmd

		}
	}

	return m, cmd
}

func (m *UseCurrentCofigSelector) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{m.configPath}, styles.Question) + m.Selector.View()
}

type DALayerNetwork string

const (
	Initia          DALayerNetwork = "Initia"
	CelestiaMainnet DALayerNetwork = "Celestia Mainnet"
	CelestiaTestnet DALayerNetwork = "Celestia Testnet"
	// Add other types as needed
)

type SetDALayer struct {
	utils.Selector[DALayerNetwork]
	state    *OPInitBotsState
	question string
}

func NewSetDALayer(state *OPInitBotsState) tea.Model {
	return &SetDALayer{
		Selector: utils.Selector[DALayerNetwork]{
			Options: []DALayerNetwork{
				Initia,
				CelestiaMainnet,
				CelestiaTestnet,
			},
		},
		state:    state,
		question: "Which DA Layer would you like to use?",
	}
}

func (m *SetDALayer) GetQuestion() string {
	return m.question
}

func (m *SetDALayer) Init() tea.Cmd {
	return nil
}

func (m *SetDALayer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"DA Layer"}, string(*selected)))
		m.state.botConfig["da_layer_network"] = string(*selected)
		return NewFieldInputModel(m.state, defaultDALayerFields, NewStartingInitBot), cmd
	}

	return m, cmd
}

func (m *SetDALayer) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"DA Layer"}, styles.Question) + m.Selector.View()
}

type StartingInitBot struct {
	state   *OPInitBotsState
	loading utils.Loading
}

func NewStartingInitBot(state *OPInitBotsState) tea.Model {
	return &StartingInitBot{
		state:   state,
		loading: utils.NewLoading("Setting up OPinit bot (TODO)...", WaitStartingInitBot(state)),
	}
}

func WaitStartingInitBot(state *OPInitBotsState) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1500 * time.Millisecond)
		configMap := state.botConfig
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}

		if state.isDeleteDB {
			err := utils.DeleteFile(state.dbPath)
			if err != nil {
				panic(err)
			}
		}

		if state.InitExecutorBot {
			version, _ := strconv.Atoi(configMap["version"])
			l2StartHeight, _ := strconv.Atoi(configMap["l2_start_height"])
			batchStartHeight, _ := strconv.Atoi(configMap["batch_start_height"])

			config := ExecutorConfig{
				Version:       version,
				ListenAddress: configMap["listen_address"],
				L1Node: NodeSettings{
					ChainID:       configMap["l1_node.chain_id"],
					RPCAddress:    configMap["l1_node.rpc_address"],
					Bech32Prefix:  "init",
					GasPrice:      configMap["l1_node.gas_price"],
					GasAdjustment: 1.5,
					TxTimeout:     60,
				},
				L2Node: NodeSettings{
					ChainID:       configMap["l2_node.chain_id"],
					RPCAddress:    configMap["l2_node.rpc_address"],
					Bech32Prefix:  "init",
					GasPrice:      configMap["l2_node.gas_price"],
					GasAdjustment: 1.5,
					TxTimeout:     60,
				},
				// TODO: revisit da node
				DANode: NodeSettings{
					ChainID:       configMap["l2_node.chain_id"],
					RPCAddress:    configMap["l2_node.rpc_address"],
					Bech32Prefix:  "init",
					GasPrice:      configMap["l2_node.gas_price"],
					GasAdjustment: 1.5,
					TxTimeout:     60,
				},
				OutputSubmitter:       "",
				BridgeExecutor:        "",
				BatchSubmitterEnabled: true,
				MaxChunks:             5000,
				MaxChunkSize:          300000,
				MaxSubmissionTime:     3600,
				L2StartHeight:         l2StartHeight,
				BatchStartHeight:      batchStartHeight,
			}
			configBz, err := json.MarshalIndent(config, "", " ")
			if err != nil {
				panic(fmt.Errorf("failed to marshal config: %v", err))
			}

			configFilePath := filepath.Join(userHome, utils.OPinitDirectory, "executor.json")
			if err = os.WriteFile(configFilePath, configBz, 0600); err != nil {
				panic(fmt.Errorf("failed to write config file: %v", err))
			}
		} else if state.InitChallengerBot {
			version, _ := strconv.Atoi(configMap["version"])
			l2StartHeight, _ := strconv.Atoi(configMap["l2_start_height"])
			config := ChallengerConfig{
				Version:       version,
				ListenAddress: configMap["listen_address"],
				L1Node: NodeConfig{
					ChainID:      configMap["l1_node.chain_id"],
					RPCAddress:   configMap["l1_node.rpc_address"],
					Bech32Prefix: "init",
				},
				L2Node: NodeConfig{
					ChainID:      configMap["l2_node.chain_id"],
					RPCAddress:   configMap["l2_node.rpc_address"],
					Bech32Prefix: "init",
				},
				L2StartHeight: l2StartHeight,
			}
			configBz, err := json.MarshalIndent(config, "", " ")
			if err != nil {
				panic(fmt.Errorf("failed to marshal config: %v", err))
			}

			configFilePath := filepath.Join(userHome, utils.OPinitDirectory, "challenger.json")
			if err = os.WriteFile(configFilePath, configBz, 0600); err != nil {
				panic(fmt.Errorf("failed to write config file: %v", err))
			}
		}
		return utils.EndLoading{}
	}
}

func (m *StartingInitBot) Init() tea.Cmd {
	return m.loading.Init()
}

func (m *StartingInitBot) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		return NewOPinitBotSuccessful(m.state), nil
	}
	return m, cmd
}

func (m *StartingInitBot) View() string {
	return m.state.weave.Render() + m.loading.View()
}

type OPinitBotSuccessful struct {
	state *OPInitBotsState
}

func NewOPinitBotSuccessful(state *OPInitBotsState) *OPinitBotSuccessful {
	return &OPinitBotSuccessful{
		state: state,
	}
}

func (m *OPinitBotSuccessful) Init() tea.Cmd {
	return nil
}

func (m *OPinitBotSuccessful) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m *OPinitBotSuccessful) View() string {
	return m.state.weave.Render() + styles.RenderPrompt("OPInit bot setup successful.", []string{}, styles.Completed) + "\n"
}
