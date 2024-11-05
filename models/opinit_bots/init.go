package opinit_bots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/registry"
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
	utils.BaseModel
	utils.Selector[OPInitBotInitOption]
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
	{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address", Placeholder: `Press tab to use "localhost:3000"`, DefaultValue: "localhost:3000", ValidateFn: utils.ValidateEmptyString},

	// L1 Node Configuration
	{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", ValidateFn: utils.ValidateURL},

	// L2 Node Configuration
	{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: utils.ValidateEmptyString},
	{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: `Press tab to use "http://localhost:26657"`, DefaultValue: "http://localhost:26657", ValidateFn: utils.ValidateURL},
}

func GetField(fields []*Field, name string) *Field {
	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}
	panic(fmt.Sprintf("field %s not found", name))
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

	state.ReplaceBotConfig = true

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
		return NewUseCurrentConfigSelector(ctx, "challenger")
	}

	state.ReplaceBotConfig = true

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

		userHome, _ := os.UserHomeDir()
		executorJsonPath := filepath.Join(userHome, utils.OPinitDirectory, fmt.Sprintf("%s.json", m.bot))
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
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewUseCurrentConfigSelector(utils.SetCurrentState(m.Ctx, state), m.bot), cmd
		}

		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		state.ReplaceBotConfig = true
		if state.MinitiaConfig != nil {
			m.Ctx = utils.SetCurrentState(m.Ctx, state)
			return NewPrefillMinitiaConfig(m.Ctx), cmd
		}
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		return NewL1PrefillSelector(m.Ctx), cmd
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
			if state.InitExecutorBot || state.InitChallengerBot {
				m.Ctx = utils.SetCurrentState(m.Ctx, state)
				return NewL1PrefillSelector(m.Ctx), cmd
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
			state.botConfig["l1_node.chain_id"] = minitiaConfig.L1Config.ChainID
			GetField(defaultExecutorFields, "l1_node.rpc_address").PrefillValue = minitiaConfig.L1Config.RpcUrl
			GetField(defaultExecutorFields, "l2_node.chain_id").PrefillValue = minitiaConfig.L2Config.ChainID
			GetField(defaultExecutorFields, "l2_node.gas_price").PrefillValue = "0.015" + minitiaConfig.L2Config.Denom
			GetField(defaultExecutorFields, "l2_node.gas_price").Placeholder = "Press tab to use " + "\"0.015" + minitiaConfig.L2Config.Denom + "\""

			GetField(defaultChallengerFields, "l1_node.rpc_address").PrefillValue = minitiaConfig.L1Config.RpcUrl
			GetField(defaultChallengerFields, "l2_node.chain_id").PrefillValue = minitiaConfig.L2Config.ChainID
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
	L1PrefillOptionTestnet L1PrefillOption = ""
	L1PrefillOptionCustom  L1PrefillOption = "Custom"
)

type L1PrefillSelector struct {
	utils.Selector[L1PrefillOption]
	utils.BaseModel
	question string
}

func NewL1PrefillSelector(ctx context.Context) *L1PrefillSelector {
	initiaTestnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
	L1PrefillOptionTestnet = L1PrefillOption(fmt.Sprintf("Testnet (%s)", initiaTestnetRegistry.GetChainId()))
	return &L1PrefillSelector{
		Selector: utils.Selector[L1PrefillOption]{
			Options: []L1PrefillOption{
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

		case L1PrefillOptionTestnet:
			chainRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
			chainId = chainRegistry.GetChainId()
			rpc = chainRegistry.MustGetActiveRpc()
		}
		state.botConfig["l1_node.chain_id"] = chainId

		// To be replaced with information from registry
		state.botConfig["l1_node.gas_price"] = "0.015uinit"

		GetField(defaultExecutorFields, "l1_node.rpc_address").PrefillValue = rpc
		GetField(defaultChallengerFields, "l1_node.rpc_address").PrefillValue = rpc

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
	Initia   DALayerNetwork = "Initia"
	Celestia DALayerNetwork = "Celestia"
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
				Celestia,
			},
			CannotBack: true,
		},
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
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
		case Celestia:
			var network registry.ChainType
			if registry.MustGetChainRegistry(registry.InitiaL1Testnet).GetChainId() == state.botConfig["l1_node.chain_id"] {
				network = registry.CelestiaTestnet
			} else {
				network = registry.CelestiaMainnet
			}
			chainRegistry := registry.MustGetChainRegistry(network)
			state.botConfig["da.chain_id"] = chainRegistry.GetChainId()
			state.botConfig["da.rpc_address"] = chainRegistry.MustGetActiveRpc()
			state.botConfig["da.bech32_prefix"] = chainRegistry.GetBech32Prefix()
			state.botConfig["da.gas_price"] = chainRegistry.MustGetMinGasPriceByDenom(DefaultCelestiaGasDenom)
			state.daIsCelestia = true
		}
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
		model := NewStartingInitBot(m.Ctx)
		return model, model.Init()

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

		if state.daIsCelestia {
			daKeyPath := filepath.Join(userHome, utils.OPinitDirectory, configMap["da.chain_id"])
			err = utils.CopyDirectory(weaveDummyKeyPath, daKeyPath)
			if err != nil {
				panic(err)
			}
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

			version := registry.MustGetOPInitBotsSpecVersion(state.botConfig["l1_node.chain_id"])

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

			version := registry.MustGetOPInitBotsSpecVersion(state.botConfig["l1_node.chain_id"])
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

	botConfigFileName := "executor"
	if state.InitChallengerBot {
		botConfigFileName = "challenger"
	}

	return state.weave.Render() + styles.RenderPrompt("OPInit bot setup successfully. Config file is saved at ~/.opinit/"+botConfigFileName+".json", []string{}, styles.Completed) + "\n" + styles.RenderPrompt("You can start the bot by running `weave opinit-bots start "+botConfigFileName+"`", []string{}, styles.Completed) + "\n"

}
