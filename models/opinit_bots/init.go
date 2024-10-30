package opinit_bots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/initia-labs/weave/service"

	tea "github.com/charmbracelet/bubbletea"

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
	utils.BaseModel
	utils.Selector[OPInitBotInitOption]
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

func NewOPInitBotInitSelector(ctx context.Context) tea.Model {
	return &OPInitBotInitSelector{
		Selector: utils.Selector[OPInitBotInitOption]{
			Options:    []OPInitBotInitOption{ExecutorOPInitBotInitOption, ChallengerOPInitBotInitOption},
			CannotBack: true,
		},
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Which bot would you like to run?",
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

func OPInitBotInitSelectExecutor(ctx context.Context) tea.Model {
	homeDir, _ := os.UserHomeDir()

	state := utils.GetCurrentState[OPInitBotsState](ctx)
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
		ctx = utils.SetCurrentState(ctx, state)
		return NewDeleteDBSelector(ctx, "executor")
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
		ctx = utils.SetCurrentState(ctx, state)
		return NewUseCurrentConfigSelector(ctx, "executor")
	}

	if state.MinitiaConfig != nil {
		ctx = utils.SetCurrentState(ctx, state)
		return NewPrefillMinitiaConfig(ctx)
	}

	ctx = utils.SetCurrentState(ctx, state)
	return NewL1PrefillSelector(ctx)
}

func OPInitBotInitSelectChallenger(ctx context.Context) tea.Model {
	homeDir, _ := os.UserHomeDir()

	state := utils.GetCurrentState[OPInitBotsState](ctx)
	state.InitChallengerBot = true

	state.dbPath = filepath.Join(homeDir, utils.OPinitDirectory, "challenger.db")
	if utils.FileOrFolderExists(state.dbPath) {
		ctx = utils.SetCurrentState(ctx, state)
		return NewDeleteDBSelector(ctx, "challenger")
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
		ctx = utils.SetCurrentState(ctx, state)
		return NewUseCurrentConfigSelector(ctx, "executor")
	}

	if state.MinitiaConfig != nil {
		ctx = utils.SetCurrentState(ctx, state)
		return NewPrefillMinitiaConfig(ctx)
	}
	ctx = utils.SetCurrentState(ctx, state)
	return NewL1PrefillSelector(ctx)
}

func (m *OPInitBotInitSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bot"}, string(*selected)))
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		var nextModel tea.Model
		switch *selected {
		case ExecutorOPInitBotInitOption:
			nextModel = OPInitBotInitSelectExecutor(m.Ctx)
		case ChallengerOPInitBotInitOption:
			nextModel = OPInitBotInitSelectChallenger(m.Ctx)
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
	utils.BaseModel
	question string
	bot      string
}

func NewDeleteDBSelector(ctx context.Context, bot string) *DeleteDBSelector {
	return &DeleteDBSelector{
		Selector: utils.Selector[DeleteDBOption]{
			Options: []DeleteDBOption{
				DeleteDBOptionNo,
				DeleteDBOptionYes,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Would you like to reset the database?",
		bot:       bot,
	}
}

func (m *DeleteDBSelector) GetQuestion() string {
	return m.question
}

func (m *DeleteDBSelector) Init() tea.Cmd {
	return nil
}

func (m *DeleteDBSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, string(*selected)))
		switch *selected {
		case DeleteDBOptionNo:
			state.isDeleteDB = false
		case DeleteDBOptionYes:
			state.isDeleteDB = true
		}
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NewUseCurrentConfigSelector(m.Ctx, m.bot), cmd

	}

	return m, cmd
}

func (m *DeleteDBSelector) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + m.Selector.View()
}

type UseCurrentConfigSelector struct {
	utils.Selector[string]
	utils.BaseModel
	question   string
	configPath string
}

func NewUseCurrentConfigSelector(ctx context.Context, bot string) *UseCurrentConfigSelector {
	configPath := fmt.Sprintf(".opinit/%s.json", bot)
	return &UseCurrentConfigSelector{
		Selector: utils.Selector[string]{
			Options: []string{
				"use current file",
				"replace",
			},
		},
		BaseModel:  utils.BaseModel{Ctx: ctx},
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
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{m.configPath}, *selected))
		switch *selected {
		case "use current file":
			state.ReplaceBotConfig = false
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			model := NewStartingInitBot(m.Ctx)
			return model, model.Init()
		case "replace":
			state.ReplaceBotConfig = true
			if state.MinitiaConfig != nil {
				m.Ctx = utils.SetCurrentState(m.Ctx, state)
				return NewPrefillMinitiaConfig(m.Ctx), cmd
			}
			if state.InitExecutorBot {
				m.Ctx = utils.SetCurrentState(m.Ctx, state)
				return NewFieldInputModel(m.Ctx, defaultExecutorFields, NewSetDALayer), cmd
			} else if state.InitChallengerBot {
				m.Ctx = utils.SetCurrentState(m.Ctx, state)
				return NewFieldInputModel(m.Ctx, defaultChallengerFields, NewStartingInitBot), cmd
			}
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return m, cmd
		}
	}

	return m, cmd
}

func (m *UseCurrentConfigSelector) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{m.configPath}, styles.Question) + m.Selector.View()
}

type PrefillMinitiaConfigOption string

const (
	PrefillMinitiaConfigYes = "Yes, prefill"
	PrefillMinitiaConfigNo  = "No, skip"
)

type PrefillMinitiaConfig struct {
	utils.Selector[PrefillMinitiaConfigOption]
	utils.BaseModel
	question string
}

func NewPrefillMinitiaConfig(ctx context.Context) *PrefillMinitiaConfig {
	return &PrefillMinitiaConfig{
		Selector: utils.Selector[PrefillMinitiaConfigOption]{
			Options: []PrefillMinitiaConfigOption{
				PrefillMinitiaConfigYes,
				PrefillMinitiaConfigNo,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Existing .minitia/artifacts/config.json detected. Would you like to use the data in this file to pre-fill some fields?",
	}
}

func (m *PrefillMinitiaConfig) GetQuestion() string {
	return m.question
}

func (m *PrefillMinitiaConfig) Init() tea.Cmd {
	return nil
}

func (m *PrefillMinitiaConfig) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{".minitia/artifacts/config.json"}, string(*selected)))

		switch *selected {
		case PrefillMinitiaConfigYes:
			minitiaConfig := state.MinitiaConfig
			defaultExecutorFields[2].PrefillValue = minitiaConfig.L1Config.ChainID
			defaultExecutorFields[3].PrefillValue = minitiaConfig.L1Config.RpcUrl
			defaultExecutorFields[4].PrefillValue = minitiaConfig.L1Config.GasPrices
			defaultExecutorFields[5].PrefillValue = minitiaConfig.L2Config.ChainID

			defaultChallengerFields[2].PrefillValue = minitiaConfig.L1Config.ChainID
			defaultChallengerFields[3].PrefillValue = minitiaConfig.L1Config.RpcUrl
			defaultChallengerFields[4].PrefillValue = minitiaConfig.L2Config.ChainID
			m.Ctx = utils.SetCurrentState(m.Ctx, state)

			if state.InitExecutorBot {
				return NewFieldInputModel(m.Ctx, defaultExecutorFields, NewSetDALayer), cmd
			} else if state.InitChallengerBot {
				return NewFieldInputModel(m.Ctx, defaultChallengerFields, NewStartingInitBot), cmd
			}

		case PrefillMinitiaConfigNo:
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewL1PrefillSelector(m.Ctx), cmd
		}

	}

	return m, cmd
}

func (m *PrefillMinitiaConfig) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{".minitia/artifacts/config.json"}, styles.Question) + m.Selector.View()
}

type L1PrefillOption string

var (
	L1PrefillOptionMainnet L1PrefillOption = ""
	L1PrefillOptionTestnet L1PrefillOption = ""
	L1PrefillOptionCustom  L1PrefillOption = "Custom"
)

type L1PrefillSelector struct {
	utils.Selector[L1PrefillOption]
	utils.BaseModel
	question string
}

func NewL1PrefillSelector(ctx context.Context) *L1PrefillSelector {
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
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Which L1 would you like your Minitia to connect to?",
	}
}

func (m *L1PrefillSelector) GetQuestion() string {
	return m.question
}

func (m *L1PrefillSelector) Init() tea.Cmd {
	return nil
}

func (m *L1PrefillSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"L1"}, string(*selected)))

		var chainId, rpc string
		switch *selected {

		case L1PrefillOptionMainnet:
			chainId = fmt.Sprintf("%s", utils.GetConfig("constants.chain_id.mainnet"))
			rpc = fmt.Sprintf("%s", utils.GetConfig("constants.endpoints.mainnet.rpc"))

		case L1PrefillOptionTestnet:
			chainId = fmt.Sprintf("%s", utils.GetConfig("constants.chain_id.testnet"))
			rpc = fmt.Sprintf("%s", utils.GetConfig("constants.endpoints.testnet.rpc"))
		}
		defaultExecutorFields[2].PrefillValue = chainId
		defaultExecutorFields[3].PrefillValue = rpc

		defaultChallengerFields[2].PrefillValue = chainId
		defaultChallengerFields[3].PrefillValue = rpc

		if state.InitExecutorBot {
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewFieldInputModel(m.Ctx, defaultExecutorFields, NewSetDALayer), cmd
		} else if state.InitChallengerBot {
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewFieldInputModel(m.Ctx, defaultChallengerFields, NewStartingInitBot), cmd
		}

	}

	return m, cmd
}

func (m *L1PrefillSelector) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L1"}, styles.Question) + m.Selector.View()
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
	utils.BaseModel
	question string
}

func NewSetDALayer(ctx context.Context) tea.Model {
	return &SetDALayer{
		Selector: utils.Selector[DALayerNetwork]{
			Options: []DALayerNetwork{
				Initia,
				CelestiaMainnet,
				CelestiaTestnet,
			},
		},
		BaseModel: utils.BaseModel{Ctx: ctx},
		question:  "Which DA Layer would you like to use?",
	}
}

func (m *SetDALayer) GetQuestion() string {
	return m.question
}

func (m *SetDALayer) Init() tea.Cmd {
	return nil
}

func (m *SetDALayer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.Ctx = utils.CloneStateAndPushPage[OPInitBotsState](m.Ctx, m)
		state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"DA Layer"}, string(*selected)))
		switch *selected {
		case Initia:
			state.botConfig["da.chain_id"] = state.botConfig["l1_node.chain_id"]
			state.botConfig["da.rpc_address"] = state.botConfig["l1_node.rpc_address"]
			state.botConfig["da.bech32_prefix"] = "init"
			state.botConfig["da.gas_price"] = state.botConfig["l1_node.gas_price"]
		case CelestiaMainnet:
			state.botConfig["da.chain_id"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_mainnet.chain_id"))
			state.botConfig["da.rpc_address"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_mainnet.rpc"))
			state.botConfig["da.bech32_prefix"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_mainnet.bech32_prefix"))
			state.botConfig["da.gas_price"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_mainnet.gas_price"))
		case CelestiaTestnet:
			state.botConfig["da.chain_id"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_testnet.chain_id"))
			state.botConfig["da.rpc_address"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_testnet.rpc"))
			state.botConfig["da.bech32_prefix"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_testnet.bech32_prefix"))
			state.botConfig["da.gas_price"] = fmt.Sprintf("%s", utils.GetConfig("constants.da_layer.celestia_testnet.gas_price"))
		}
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NewFieldInputModel(m.Ctx, defaultDALayerFields, NewStartingInitBot), cmd
	}

	return m, cmd
}

func (m *SetDALayer) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"DA Layer"}, styles.Question) + m.Selector.View()
}

type StartingInitBot struct {
	utils.BaseModel
	loading utils.Loading
}

func NewStartingInitBot(ctx context.Context) tea.Model {
	state := utils.GetCurrentState[OPInitBotsState](ctx)
	var bot string
	if state.InitExecutorBot {
		bot = "executor"
	} else {
		bot = "challenger"
	}
	return &StartingInitBot{
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   utils.NewLoading(fmt.Sprintf("Setting up OPinit bot %s...", bot), WaitStartingInitBot(&state)),
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
		return NewOPinitBotSuccessful(m.Ctx), nil
	}
	return m, cmd
}

func (m *StartingInitBot) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + m.loading.View()
}

type OPinitBotSuccessful struct {
	utils.BaseModel
}

func NewOPinitBotSuccessful(ctx context.Context) *OPinitBotSuccessful {
	return &OPinitBotSuccessful{
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *OPinitBotSuccessful) Init() tea.Cmd {
	return nil
}

func (m *OPinitBotSuccessful) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m *OPinitBotSuccessful) View() string {
	state := utils.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt("OPInit bot setup successful.", []string{}, styles.Completed) + "\n"
}
