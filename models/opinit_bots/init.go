package opinit_bots

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/service"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/utils"
)

type OPInitBotInitOption string

const (
	ExecutorOPInitBotInitOption   OPInitBotInitOption = "Executor"
	ChallengerOPInitBotInitOption OPInitBotInitOption = "Challenger"
)

type OPInitBotInitSelector struct {
	utils.Selector[OPInitBotInitOption]
	state    *OPInitBotsState
	question string
}

var defaultExecutorFields = []*Field{
	// Listen Address
	{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address", Placeholder: `Press tab to use "localhost:3000"`, DefaultValue: "localhost:3000", ValidateFn: utils.ValidateEmptyString},

	// L1 Node Configuration
	{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", ValidateFn: utils.ValidateURL},

	// L2 Node Configuration
	{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: utils.ValidateEmptyString},
	{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: `Press tab to use "http://localhost:26657"`, DefaultValue: "http://localhost:26657", ValidateFn: utils.ValidateURL},
	{Name: "l2_node.gas_price", Type: StringField, Question: "Please specify the L2 gas_price", Placeholder: `Press tab to use "0.015umin"`, DefaultValue: "0.015umin", ValidateFn: utils.ValidateDecCoin},
}

var defaultChallengerFields = []*Field{
	// Listen Address
	{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address", Placeholder: `Add listen address ex. localhost:3000`, ValidateFn: utils.ValidateEmptyString},

	// L1 Node Configuration
	{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", ValidateFn: utils.ValidateURL},

	// L2 Node Configuration
	{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: utils.ValidateEmptyString},
	{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", ValidateFn: utils.ValidateURL},
}

func GetField(fields []*Field, name string) *Field {
	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}
	panic(fmt.Sprintf("field %s not found", name))
}

func NewOPInitBotInitSelector(state *OPInitBotsState) tea.Model {
	return &OPInitBotInitSelector{
		Selector: utils.Selector[OPInitBotInitOption]{
			Options: []OPInitBotInitOption{ExecutorOPInitBotInitOption, ChallengerOPInitBotInitOption},
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

type BotConfigChainId struct {
	L1Node struct {
		ChainID string `json:"chain_id"`
	} `json:"l1_node"`
	L2Node struct {
		ChainID string `json:"chain_id"`
	} `json:"l2_node"`
}

func OPInitBotInitSelectExecutor(state *OPInitBotsState) tea.Model {
	homeDir, _ := os.UserHomeDir()
	state.InitExecutorBot = true
	minitiaConfigPath := filepath.Join(homeDir, utils.MinitiaArtifactsDirectory, "config.json")

	if utils.FileOrFolderExists(minitiaConfigPath) {
		configData, err := os.ReadFile(minitiaConfigPath)
		if err != nil {
			panic(err)
		}

		var minitiaConfig types.MinitiaConfig
		err = json.Unmarshal(configData, &minitiaConfig)
		if err != nil {
			panic(err)
		}

		state.MinitiaConfig = &minitiaConfig
	}

	state.dbPath = filepath.Join(homeDir, utils.OPinitDirectory, "executor.db")
	if utils.FileOrFolderExists(state.dbPath) {
		return NewDeleteDBSelector(state, "executor")
	}

	executorJsonPath := filepath.Join(homeDir, utils.OPinitDirectory, "executor.json")
	if utils.FileOrFolderExists(executorJsonPath) {
		file, err := os.ReadFile(executorJsonPath)
		if err != nil {
			panic(err)
		}

		var botConfigChainId BotConfigChainId

		err = json.Unmarshal(file, &botConfigChainId)
		if err != nil {
			panic(err)
		}
		state.botConfig["l1_node.chain_id"] = botConfigChainId.L1Node.ChainID
		state.botConfig["l2_node.chain_id"] = botConfigChainId.L2Node.ChainID
		return NewUseCurrentConfigSelector(state, "executor")
	}

	state.ReplaceBotConfig = true

	if state.MinitiaConfig != nil {
		return NewPrefillMinitiaConfig(state)
	}

	return NewL1PrefillSelector(state)
}

func OPInitBotInitSelectChallenger(state *OPInitBotsState) tea.Model {
	homeDir, _ := os.UserHomeDir()
	state.InitChallengerBot = true

	state.dbPath = filepath.Join(homeDir, utils.OPinitDirectory, "challenger.db")
	if utils.FileOrFolderExists(state.dbPath) {
		return NewDeleteDBSelector(state, "challenger")
	}

	challengerJsonPath := filepath.Join(homeDir, utils.OPinitDirectory, "challenger.json")
	if utils.FileOrFolderExists(challengerJsonPath) {
		file, err := os.ReadFile(challengerJsonPath)
		if err != nil {
			panic(err)
		}

		var botConfigChainId BotConfigChainId

		err = json.Unmarshal(file, &botConfigChainId)
		if err != nil {
			panic(err)
		}
		state.botConfig["l1_node.chain_id"] = botConfigChainId.L1Node.ChainID
		state.botConfig["l2_node.chain_id"] = botConfigChainId.L2Node.ChainID
		return NewUseCurrentConfigSelector(state, "executor")
	}

	state.ReplaceBotConfig = true

	if state.MinitiaConfig != nil {
		return NewPrefillMinitiaConfig(state)
	}
	return NewL1PrefillSelector(state)
}

func (m *OPInitBotInitSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bot"}, string(*selected)))

		var nextModel tea.Model
		switch *selected {
		case ExecutorOPInitBotInitOption:
			nextModel = OPInitBotInitSelectExecutor(m.state)
		case ChallengerOPInitBotInitOption:
			nextModel = OPInitBotInitSelectChallenger(m.state)
		}

		return nextModel, cmd
	}

	return m, cmd
}

func (m *OPInitBotInitSelector) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{"bot"}, styles.Question) + m.Selector.View()
}

type DeleteDBOption string

const (
	DeleteDBOptionNo  = "No"
	DeleteDBOptionYes = "Yes, reset"
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
				DeleteDBOptionNo,
				DeleteDBOptionYes,
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
		case DeleteDBOptionNo:
			m.state.isDeleteDB = false
		case DeleteDBOptionYes:
			m.state.isDeleteDB = true
		}
		return NewUseCurrentConfigSelector(m.state, m.bot), cmd

	}

	return m, cmd
}

func (m *DeleteDBSelector) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + m.Selector.View()
}

type UseCurrentConfigSelector struct {
	utils.Selector[string]
	state      *OPInitBotsState
	question   string
	configPath string
}

func NewUseCurrentConfigSelector(state *OPInitBotsState, bot string) *UseCurrentConfigSelector {
	configPath := fmt.Sprintf(".opinit/%s.json", bot)
	return &UseCurrentConfigSelector{
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

func (m *UseCurrentConfigSelector) GetQuestion() string {
	return m.question
}

func (m *UseCurrentConfigSelector) Init() tea.Cmd {
	return nil
}

func (m *UseCurrentConfigSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{m.configPath}, *selected))

		switch *selected {
		case "use current file":
			m.state.ReplaceBotConfig = false

			model := NewStartingInitBot(m.state)
			return model, model.Init()
		case "replace":
			m.state.ReplaceBotConfig = true
			if m.state.MinitiaConfig != nil {
				return NewPrefillMinitiaConfig(m.state), cmd
			}
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

func (m *UseCurrentConfigSelector) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{m.configPath}, styles.Question) + m.Selector.View()
}

type PrefillMinitiaConfigOption string

const (
	PrefillMinitiaConfigYes = "Yes, prefill"
	PrefillMinitiaConfigNo  = "No, skip"
)

type PrefillMinitiaConfig struct {
	utils.Selector[PrefillMinitiaConfigOption]
	state    *OPInitBotsState
	question string
}

func NewPrefillMinitiaConfig(state *OPInitBotsState) *PrefillMinitiaConfig {
	return &PrefillMinitiaConfig{
		Selector: utils.Selector[PrefillMinitiaConfigOption]{
			Options: []PrefillMinitiaConfigOption{
				PrefillMinitiaConfigYes,
				PrefillMinitiaConfigNo,
			},
		},
		state:    state,
		question: "Existing .minitia/artifacts/config.json detected. Would you like to use the data in this file to pre-fill some fields?",
	}
}

func (m *PrefillMinitiaConfig) GetQuestion() string {
	return m.question
}

func (m *PrefillMinitiaConfig) Init() tea.Cmd {
	return nil
}

func (m *PrefillMinitiaConfig) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{".minitia/artifacts/config.json"}, string(*selected)))

		switch *selected {
		case PrefillMinitiaConfigYes:
			minitiaConfig := m.state.MinitiaConfig
			m.state.botConfig["l1_node.chain_id"] = minitiaConfig.L1Config.ChainID
			GetField(defaultExecutorFields, "l1_node.rpc_address").PrefillValue = minitiaConfig.L1Config.RpcUrl
			GetField(defaultExecutorFields, "l2_node.chain_id").PrefillValue = minitiaConfig.L2Config.ChainID
			GetField(defaultExecutorFields, "l2_node.gas_price").PrefillValue = "0.015" + minitiaConfig.L2Config.Denom
			GetField(defaultExecutorFields, "l2_node.gas_price").Placeholder = "Press tab to use " + "\"0.015" + minitiaConfig.L2Config.Denom + "\""

			GetField(defaultChallengerFields, "l1_node.rpc_address").PrefillValue = minitiaConfig.L1Config.RpcUrl
			GetField(defaultChallengerFields, "l2_node.chain_id").PrefillValue = minitiaConfig.L2Config.ChainID
			if m.state.InitExecutorBot {
				return NewFieldInputModel(m.state, defaultExecutorFields, NewSetDALayer), cmd
			} else if m.state.InitChallengerBot {
				return NewFieldInputModel(m.state, defaultChallengerFields, NewStartingInitBot), cmd
			}

		case PrefillMinitiaConfigNo:
			return NewL1PrefillSelector(m.state), cmd
		}

	}

	return m, cmd
}

func (m *PrefillMinitiaConfig) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{".minitia/artifacts/config.json"}, styles.Question) + m.Selector.View()
}

type L1PrefillOption string

var (
	L1PrefillOptionMainnet L1PrefillOption = ""
	L1PrefillOptionTestnet L1PrefillOption = ""
	L1PrefillOptionCustom  L1PrefillOption = "Custom"
)

type L1PrefillSelector struct {
	utils.Selector[L1PrefillOption]
	state    *OPInitBotsState
	question string
}

func NewL1PrefillSelector(state *OPInitBotsState) *L1PrefillSelector {
	L1PrefillOptionMainnet = L1PrefillOption(fmt.Sprintf("Mainnet (%s)", utils.GetConfig("constants.chain_id.mainnet")))
	L1PrefillOptionTestnet = L1PrefillOption(fmt.Sprintf("Testnet (%s)", utils.GetConfig("constants.chain_id.testnet")))
	return &L1PrefillSelector{
		Selector: utils.Selector[L1PrefillOption]{
			Options: []L1PrefillOption{
				L1PrefillOptionMainnet,
				L1PrefillOptionTestnet,
				L1PrefillOptionCustom,
			},
		},
		state:    state,
		question: "Which L1 would you like your Minitia to connect to?",
	}
}

func (m *L1PrefillSelector) GetQuestion() string {
	return m.question
}

func (m *L1PrefillSelector) Init() tea.Cmd {
	return nil
}

func (m *L1PrefillSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"L1"}, string(*selected)))

		var chainId, rpc string
		switch *selected {

		case L1PrefillOptionMainnet:
			chainId = fmt.Sprintf("%s", utils.GetConfig("constants.chain_id.mainnet"))
			rpc = fmt.Sprintf("%s", utils.GetConfig("constants.endpoints.mainnet.rpc"))

		case L1PrefillOptionTestnet:
			chainId = fmt.Sprintf("%s", utils.GetConfig("constants.chain_id.testnet"))
			rpc = fmt.Sprintf("%s", utils.GetConfig("constants.endpoints.testnet.rpc"))
		}
		m.state.botConfig["l1_node.chain_id"] = chainId

		// To be replaced with information from registry
		m.state.botConfig["l1_node.gas_price"] = "0.015uinit"

		GetField(defaultExecutorFields, "l1_node.rpc_address").PrefillValue = rpc
		GetField(defaultChallengerFields, "l1_node.rpc_address").PrefillValue = rpc

		if m.state.InitExecutorBot {
			return NewFieldInputModel(m.state, defaultExecutorFields, NewSetDALayer), cmd
		} else if m.state.InitChallengerBot {
			return NewFieldInputModel(m.state, defaultChallengerFields, NewStartingInitBot), cmd
		}

	}

	return m, cmd
}

func (m *L1PrefillSelector) View() string {
	return m.state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L1"}, styles.Question) + m.Selector.View()
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
		switch *selected {
		case Initia:
			m.state.botConfig["da.chain_id"] = m.state.botConfig["l1_node.chain_id"]
			m.state.botConfig["da.rpc_address"] = m.state.botConfig["l1_node.rpc_address"]
			m.state.botConfig["da.bech32_prefix"] = "init"
			m.state.botConfig["da.gas_price"] = m.state.botConfig["l1_node.gas_price"]
		case CelestiaMainnet:
			m.state.botConfig["da.chain_id"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_mainnet.chain_id"))
			m.state.botConfig["da.rpc_address"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_mainnet.rpc"))
			m.state.botConfig["da.bech32_prefix"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_mainnet.bech32_prefix"))
			m.state.botConfig["da.gas_price"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_mainnet.gas_price"))
		case CelestiaTestnet:
			m.state.botConfig["da.chain_id"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_testnet.chain_id"))
			m.state.botConfig["da.rpc_address"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_testnet.rpc"))
			m.state.botConfig["da.bech32_prefix"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_testnet.bech32_prefix"))
			m.state.botConfig["da.gas_price"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_testnet.gas_price"))
		}
		model := NewStartingInitBot(m.state)
		return model, model.Init()
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
	var bot string
	if state.InitExecutorBot {
		bot = "executor"
	} else {
		bot = "challenger"
	}

	// default config
	state.botConfig["version"] = "1"

	return &StartingInitBot{
		state:   state,
		loading: utils.NewLoading(fmt.Sprintf("Setting up OPinit bot %s...", bot), WaitStartingInitBot(state)),
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
			err := utils.DeleteDirectory(state.dbPath)
			if err != nil {
				panic(err)
			}
		}

		weaveDummyKeyPath := filepath.Join(userHome, utils.OPinitDirectory, "weave-dummy")
		l1KeyPath := filepath.Join(userHome, utils.OPinitDirectory, configMap["l1_node.chain_id"])
		l2KeyPath := filepath.Join(userHome, utils.OPinitDirectory, configMap["l2_node.chain_id"])

		err = utils.CopyDirectory(weaveDummyKeyPath, l1KeyPath)
		if err != nil {
			panic(err)
		}
		err = utils.CopyDirectory(weaveDummyKeyPath, l2KeyPath)
		if err != nil {
			panic(err)
		}

		if state.InitExecutorBot {
			srv, err := service.NewService(service.OPinitExecutor)
			if err != nil {
				panic(fmt.Sprintf("failed to initialize service: %v", err))
			}

			if err = srv.Create(""); err != nil {
				panic(fmt.Sprintf("failed to create service: %v", err))
			}

			if !state.ReplaceBotConfig {
				return utils.EndLoading{}
			}

			version, _ := strconv.Atoi(configMap["version"])

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
				DANode: NodeSettings{
					ChainID:       configMap["da.chain_id"],
					RPCAddress:    configMap["da.rpc_address"],
					Bech32Prefix:  configMap["da.bech32_prefix"],
					GasPrice:      configMap["da.gas_price"],
					GasAdjustment: 1.5,
					TxTimeout:     60,
				},
				OutputSubmitter:       OutputSubmitterKeyName,
				BridgeExecutor:        BridgeExecutorKeyName,
				BatchSubmitterEnabled: true,
				MaxChunks:             5000,
				MaxChunkSize:          300000,
				MaxSubmissionTime:     3600,
				L2StartHeight:         0,
				BatchStartHeight:      0,
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
			srv, err := service.NewService(service.OPinitChallenger)
			if err != nil {
				panic(fmt.Sprintf("failed to initialize service: %v", err))
			}

			if err = srv.Create(""); err != nil {
				panic(fmt.Sprintf("failed to create service: %v", err))
			}

			if !state.ReplaceBotConfig {
				return utils.EndLoading{}
			}

			version, _ := strconv.Atoi(configMap["version"])
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
				L2StartHeight: 0,
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

func (m *OPinitBotSuccessful) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m *OPinitBotSuccessful) View() string {
	return m.state.weave.Render() + styles.RenderPrompt("OPInit bot setup successful.", []string{}, styles.Completed) + "\n"
}
