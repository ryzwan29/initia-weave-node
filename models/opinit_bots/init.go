package opinit_bots

import (
	"encoding/json"
	"fmt"
	"github.com/initia-labs/weave/service"
	"os"
	"path/filepath"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
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
	{Name: "version", Type: NumberField, Question: "Please specify the version", Placeholder: `Press tab to use "1"`, DefaultValue: "1", ValidateFn: utils.IsValidInteger},

	// Listen Address
	{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address", Placeholder: `Add listen address ex. localhost:3000`, ValidateFn: utils.ValidateEmptyString},

	// L1 Node Configuration
	{Name: "l1_node.chain_id", Type: StringField, Question: "Please specify the L1 chain_id", Placeholder: "Add alphanumeric", ValidateFn: utils.ValidateEmptyString},
	{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address", Placeholder: "Add RPC address ex. tcp://localhost:26657", ValidateFn: utils.ValidateURL},
	{Name: "l1_node.gas_price", Type: StringField, Question: "Please specify the L1 gas_price", Placeholder: `Press tab to use "0.15uinit"`, DefaultValue: "0.15uinit", ValidateFn: utils.ValidateDecCoin},

	// L2 Node Configuration
	{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: utils.ValidateEmptyString},
	{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. tcp://localhost:26657", ValidateFn: utils.ValidateURL},
	{Name: "l2_node.gas_price", Type: StringField, Question: "Please specify the L2 gas_price", Placeholder: `Press tab to use "0.15uinit"`, DefaultValue: "0.15uinit", ValidateFn: utils.ValidateDecCoin},
}

var defaultDALayerFields = []Field{
	// Version
	{Name: "l2_start_height", Type: NumberField, Question: "Please specify the l2_start_height", ValidateFn: utils.IsValidInteger},

	// Listen Address
	{Name: "batch_start_height", Type: NumberField, Question: "Please specify the batch_start_height", ValidateFn: utils.IsValidInteger},
}

var defaultChallengerFields = []Field{
	// Version
	{Name: "version", Type: NumberField, Question: "Please specify the version", Placeholder: `Press tab to use "1"`, DefaultValue: "1", ValidateFn: utils.IsValidInteger},

	// Listen Address
	{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address", Placeholder: `Add listen address ex. localhost:3000`, ValidateFn: utils.ValidateEmptyString},

	// L1 Node Configuration
	{Name: "l1_node.chain_id", Type: StringField, Question: "Please specify the L1 chain_id", Placeholder: "Add alphanumeric", ValidateFn: utils.ValidateEmptyString},
	{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address", Placeholder: "Add RPC address ex. tcp://localhost:26657", ValidateFn: utils.ValidateURL},

	// L2 Node Configuration
	{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: utils.ValidateEmptyString},
	{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. tcp://localhost:26657", ValidateFn: utils.ValidateURL},

	// Version
	{Name: "l2_start_height", Type: NumberField, Question: "Please specify the l2_start_height", ValidateFn: utils.IsValidInteger},
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

			minitiaConfigPath := filepath.Join(homeDir, utils.MinitiaArtifactsDirectory, "config.json")

			// Check if the config file exists
			if utils.FileOrFolderExists(minitiaConfigPath) {
				// Load the config if found
				configData, err := os.ReadFile(minitiaConfigPath)
				if err != nil {
					panic(err)
				}

				var minitiaConfig types.MinitiaConfig
				err = json.Unmarshal(configData, &minitiaConfig)
				if err != nil {
					panic(err)
				}

				m.state.MinitiaConfig = &minitiaConfig // assuming m.state has a field for storing the config
			}

			m.state.dbPath = filepath.Join(homeDir, utils.OPinitDirectory, "executor.db")
			if utils.FileOrFolderExists(m.state.dbPath) {
				return NewDeleteDBSelector(m.state, "executor"), cmd
			}

			executorJsonPath := filepath.Join(homeDir, utils.OPinitDirectory, "executor.json")
			if utils.FileOrFolderExists(executorJsonPath) {
				return NewUseCurrentCofigSelector(m.state, "executor"), cmd
			}

			if m.state.MinitiaConfig != nil {
				return NewPrefillMinitiaConfig(m.state), cmd
			}

			return NewL1PrefillSelector(m.state), cmd
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
			return m, tea.Quit
		case "replace":
			m.state.ReplaceBotConfig = true
			if m.state.MinitiaConfig != nil {
				return NewPrefillMinitiaConfig(m.state), cmd
			}
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

type PrefillMinitiaConfigOption string

const (
	PrefillMinitiaConfig_Yes = "Yes, prefill"
	PrefillMinitiaConfig_No  = "No, skip"
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
				PrefillMinitiaConfig_Yes,
				PrefillMinitiaConfig_No,
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
		case PrefillMinitiaConfig_Yes:
			minitiaConfig := m.state.MinitiaConfig
			defaultExecutorFields[2].PrefillValue = minitiaConfig.L1Config.ChainID
			defaultExecutorFields[3].PrefillValue = minitiaConfig.L1Config.RpcUrl
			defaultExecutorFields[4].PrefillValue = minitiaConfig.L1Config.GasPrices
			defaultExecutorFields[5].PrefillValue = minitiaConfig.L2Config.ChainID

			defaultChallengerFields[2].PrefillValue = minitiaConfig.L1Config.ChainID
			defaultChallengerFields[3].PrefillValue = minitiaConfig.L1Config.RpcUrl
			defaultChallengerFields[4].PrefillValue = minitiaConfig.L2Config.ChainID
			if m.state.InitExecutorBot {
				return NewFieldInputModel(m.state, defaultExecutorFields, NewSetDALayer), cmd
			} else if m.state.InitChallengerBot {
				return NewFieldInputModel(m.state, defaultChallengerFields, NewStartingInitBot), cmd
			}

		case PrefillMinitiaConfig_No:
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
	L1PrefillOption_Mainnet L1PrefillOption = ""
	L1PrefillOption_Testnet L1PrefillOption = ""
	L1PrefillOption_Custom  L1PrefillOption = "Custom"
)

type L1PrefillSelector struct {
	utils.Selector[L1PrefillOption]
	state    *OPInitBotsState
	question string
}

func NewL1PrefillSelector(state *OPInitBotsState) *L1PrefillSelector {
	L1PrefillOption_Mainnet = L1PrefillOption(fmt.Sprintf("Mainnet (%s)", utils.GetConfig("constants.chain_id.mainnet")))
	L1PrefillOption_Testnet = L1PrefillOption(fmt.Sprintf("Testnet (%s)", utils.GetConfig("constants.chain_id.testnet")))
	return &L1PrefillSelector{
		Selector: utils.Selector[L1PrefillOption]{
			Options: []L1PrefillOption{
				L1PrefillOption_Mainnet,
				L1PrefillOption_Testnet,
				L1PrefillOption_Custom,
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

		case L1PrefillOption_Mainnet:
			chainId = fmt.Sprintf("%s", utils.GetConfig("constants.chain_id.mainnet"))
			rpc = fmt.Sprintf("%s", utils.GetConfig("constants.endpoints.mainnet.rpc"))

		case L1PrefillOption_Testnet:
			chainId = fmt.Sprintf("%s", utils.GetConfig("constants.chain_id.testnet"))
			rpc = fmt.Sprintf("%s", utils.GetConfig("constants.endpoints.testnet.rpc"))
		}
		defaultExecutorFields[2].PrefillValue = chainId
		defaultExecutorFields[3].PrefillValue = rpc

		defaultChallengerFields[2].PrefillValue = chainId
		defaultChallengerFields[3].PrefillValue = rpc

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
		// m.state.botConfig["da_layer_network"] = string(*selected)
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
	var bot string
	if state.InitExecutorBot {
		bot = "executor"
	} else {
		bot = "challenger"
	}
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
				DANode: NodeSettings{
					ChainID:       configMap["da.chain_id"],
					RPCAddress:    configMap["da.rpc_address"],
					Bech32Prefix:  configMap["da.bech32_prefix"],
					GasPrice:      configMap["da.gas_price"],
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

			srv, err := service.NewService(service.OPinitExecutor)
			if err != nil {
				panic(fmt.Sprintf("failed to initialize service: %v", err))
			}

			if err = srv.Create(""); err != nil {
				panic(fmt.Sprintf("failed to create service: %v", err))
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

			srv, err := service.NewService(service.OPinitChallenger)
			if err != nil {
				panic(fmt.Sprintf("failed to initialize service: %v", err))
			}

			if err = srv.Create(""); err != nil {
				panic(fmt.Sprintf("failed to create service: %v", err))
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
