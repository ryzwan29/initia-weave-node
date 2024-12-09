package opinit_bots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/initia-labs/weave/cosmosutils"
	"github.com/initia-labs/weave/io"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/common"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/service"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/ui"
)

type OPInitBotInitOption string

const (
	ExecutorOPInitBotInitOption   OPInitBotInitOption = "Executor"
	ChallengerOPInitBotInitOption OPInitBotInitOption = "Challenger"
)

type OPInitBotInitSelector struct {
	weavecontext.BaseModel
	ui.Selector[OPInitBotInitOption]
	question string
}

var (
	ListenAddressTooltip          = ui.NewTooltip("listen_address", "The listen_address specifies the network address and port where the bot listens for incoming transaction execution requests, often formatted as tcp://0.0.0.0:<port>.", "", []string{}, []string{}, []string{})
	L1RPCAddressTooltip           = ui.NewTooltip("L1 rpc_address", "The rpc_address for the Executor defines the network address and port where the bot's RPC interface listens. This allows other network components to communicate with the Executor via RPC calls, typically formatted as tcp://0.0.0.0:<port>.", "", []string{}, []string{}, []string{})
	L2RPCAddressTooltip           = ui.NewTooltip("L2 rpc_address", "The rpc_address for the Executor defines the network address and port where the bot's RPC interface listens. This allows other network components to communicate with the Executor via RPC calls, typically formatted as tcp://0.0.0.0:<port>.", "", []string{}, []string{}, []string{})
	L2GasPriceTooltip             = ui.NewTooltip("L2 gas_price", "The L2 gas_price specifies the minimum gas price for transactions submitted on the Layer 2 (L2) network. This value helps ensure that L2 transactions are processed with adequate priority and aligns with  Minitias or other L2 environments.", "", []string{}, []string{}, []string{})
	InitiaDALayerTooltip          = ui.NewTooltip("Initia", "Ideal for projects that require close integration within the Initia network, offering streamlined communication and data handling within the Initia ecosystem.", "", []string{}, []string{}, []string{})
	CelestiaMainnetDALayerTooltip = ui.NewTooltip("Celestia Mainnet", "Suitable for production environments that need reliable and secure data availability with Celestia's decentralized architecture, ensuring robust support for live applications.", "", []string{}, []string{}, []string{})
	CelestiaTestnetDALayerTooltip = ui.NewTooltip("Celestia Testnet", "Best for testing purposes, allowing you to validate functionality and performance in a non-production setting before deploying to a mainnet environment.", "", []string{}, []string{}, []string{})
)

var defaultExecutorFields = []*Field{
	// Listen Address
	{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address", Placeholder: `Press tab to use "localhost:3000"`, DefaultValue: "localhost:3000", ValidateFn: common.ValidateEmptyString, Tooltip: &ListenAddressTooltip},

	// L1 Node Configuration
	{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", ValidateFn: common.ValidateURL, Tooltip: &L1RPCAddressTooltip},

	// L2 Node Configuration
	{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: common.ValidateEmptyString},
	{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: `Press tab to use "http://localhost:26657"`, DefaultValue: "http://localhost:26657", ValidateFn: common.ValidateURL, Tooltip: &L2RPCAddressTooltip},
	{Name: "l2_node.gas_price", Type: StringField, Question: "Please specify the L2 gas_price", Placeholder: `Press tab to use "0.015umin"`, DefaultValue: "0.015umin", ValidateFn: common.ValidateDecCoin, Tooltip: &L2GasPriceTooltip},
}

var defaultChallengerFields = []*Field{
	// Listen Address
	{Name: "listen_address", Type: StringField, Question: "Please specify the listen_address", Placeholder: `Press tab to use "localhost:3000"`, DefaultValue: "localhost:3000", ValidateFn: common.ValidateEmptyString, Tooltip: &ListenAddressTooltip},

	// L1 Node Configuration
	{Name: "l1_node.rpc_address", Type: StringField, Question: "Please specify the L1 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", ValidateFn: common.ValidateURL, Tooltip: &L1RPCAddressTooltip},

	// L2 Node Configuration
	{Name: "l2_node.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: common.ValidateEmptyString},
	{Name: "l2_node.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: `Press tab to use "http://localhost:26657"`, DefaultValue: "http://localhost:26657", ValidateFn: common.ValidateURL, Tooltip: &L2RPCAddressTooltip},
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
	tooltips := []ui.Tooltip{
		ui.NewTooltip("Executor", "Executes cross-chain transactions, ensuring that assets and data move securely between Initia and Minitias.", "", []string{}, []string{}, []string{}),
		ui.NewTooltip("Challenger", "Monitors for potential fraud, submitting proofs to dispute invalid state updates and maintaining network security.", "", []string{}, []string{}, []string{}),
	}
	return &OPInitBotInitSelector{
		Selector: ui.Selector[OPInitBotInitOption]{
			Options:    []OPInitBotInitOption{ExecutorOPInitBotInitOption, ChallengerOPInitBotInitOption},
			CannotBack: true,
			Tooltips:   &tooltips,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
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
	DANode struct {
		ChainID      string `json:"chain_id"`
		Bech32Prefix string `json:"bech32_prefix"`
	} `json:"da_node"`
}

func OPInitBotInitSelectExecutor(ctx context.Context) tea.Model {
	state := weavecontext.GetCurrentState[OPInitBotsState](ctx)
	state.InitExecutorBot = true
	minitiaConfigPath := weavecontext.GetMinitiaArtifactsConfigJson(ctx)

	if io.FileOrFolderExists(minitiaConfigPath) {
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

	opInitHome := weavecontext.GetOPInitHome(ctx)
	state.dbPath = filepath.Join(opInitHome, "executor.db")
	if io.FileOrFolderExists(state.dbPath) {
		ctx = weavecontext.SetCurrentState(ctx, state)
		return NewDeleteDBSelector(ctx, "executor")
	}

	executorJsonPath := filepath.Join(opInitHome, "executor.json")
	if io.FileOrFolderExists(executorJsonPath) {
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
		state.botConfig["da_node.chain_id"] = botConfigChainId.DANode.ChainID
		state.daIsCelestia = botConfigChainId.DANode.Bech32Prefix == "celestia"
		ctx = weavecontext.SetCurrentState(ctx, state)
		return NewUseCurrentConfigSelector(ctx, "executor")
	}

	state.ReplaceBotConfig = true

	if state.MinitiaConfig != nil {
		ctx = weavecontext.SetCurrentState(ctx, state)
		return NewPrefillMinitiaConfig(ctx)
	}

	ctx = weavecontext.SetCurrentState(ctx, state)
	return NewL1PrefillSelector(ctx)
}

func OPInitBotInitSelectChallenger(ctx context.Context) tea.Model {
	state := weavecontext.GetCurrentState[OPInitBotsState](ctx)
	state.InitChallengerBot = true

	minitiaConfigPath := weavecontext.GetMinitiaArtifactsConfigJson(ctx)
	if io.FileOrFolderExists(minitiaConfigPath) {
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

	opInitHome := weavecontext.GetOPInitHome(ctx)
	state.dbPath = filepath.Join(opInitHome, "challenger.db")
	if io.FileOrFolderExists(state.dbPath) {
		ctx = weavecontext.SetCurrentState(ctx, state)
		return NewDeleteDBSelector(ctx, "challenger")
	}

	challengerJsonPath := filepath.Join(opInitHome, "challenger.json")
	if io.FileOrFolderExists(challengerJsonPath) {
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
		ctx = weavecontext.SetCurrentState(ctx, state)
		return NewUseCurrentConfigSelector(ctx, "challenger")
	}

	state.ReplaceBotConfig = true

	if state.MinitiaConfig != nil {
		ctx = weavecontext.SetCurrentState(ctx, state)
		return NewPrefillMinitiaConfig(ctx)
	}
	ctx = weavecontext.SetCurrentState(ctx, state)
	return NewL1PrefillSelector(ctx)
}

func (m *OPInitBotInitSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[OPInitBotsState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bot"}, string(*selected)))
		switch *selected {
		case ExecutorOPInitBotInitOption:
			keyNames := make(map[string]bool)
			keyNames[BridgeExecutorKeyName] = true
			keyNames[OutputSubmitterKeyName] = true
			keyNames[BatchSubmitterKeyName] = true
			keyNames[OracleBridgeExecutorKeyName] = true

			finished := true

			state.BotInfos = CheckIfKeysExist(BotInfos)
			for idx, botInfo := range state.BotInfos {
				if keyNames[botInfo.KeyName] && botInfo.IsNotExist {
					state.BotInfos[idx].IsSetup = true
					finished = false
				} else {
					state.BotInfos[idx].IsSetup = false
				}
			}
			if finished {
				return OPInitBotInitSelectExecutor(weavecontext.SetCurrentState(m.Ctx, state)), cmd
			}
			ja := ""
			for _, botInfo := range state.BotInfos {
				ja += fmt.Sprintf("%s => %v\n", botInfo.KeyName, botInfo.IsSetup)
			}

			state.isSetupMissingKey = true
			return NextUpdateOpinitBotKey(weavecontext.SetCurrentState(m.Ctx, state))
		case ChallengerOPInitBotInitOption:
			return OPInitBotInitSelectChallenger(weavecontext.SetCurrentState(m.Ctx, state)), cmd
		}
	}
	return m, cmd
}

func (m *OPInitBotInitSelector) View() string {
	m.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return styles.RenderPrompt(m.GetQuestion(), []string{"bot"}, styles.Question) + m.Selector.View()
}

type DeleteDBOption string

const (
	DeleteDBOptionNo  = "No"
	DeleteDBOptionYes = "Yes, reset"
)

type DeleteDBSelector struct {
	ui.Selector[DeleteDBOption]
	weavecontext.BaseModel
	question string
	bot      string
}

func NewDeleteDBSelector(ctx context.Context, bot string) *DeleteDBSelector {
	return &DeleteDBSelector{
		Selector: ui.Selector[DeleteDBOption]{
			Options: []DeleteDBOption{
				DeleteDBOptionNo,
				DeleteDBOptionYes,
			},
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[OPInitBotsState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, string(*selected)))
		switch *selected {
		case DeleteDBOptionNo:
			state.isDeleteDB = false
		case DeleteDBOptionYes:
			state.isDeleteDB = true
		}

		opInitHome := weavecontext.GetOPInitHome(m.Ctx)
		executorJsonPath := filepath.Join(opInitHome, fmt.Sprintf("%s.json", m.bot))
		if io.FileOrFolderExists(executorJsonPath) {
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
			state.botConfig["da_node.chain_id"] = botConfigChainId.DANode.ChainID
			state.daIsCelestia = botConfigChainId.DANode.Bech32Prefix == "celestia"
			return NewUseCurrentConfigSelector(weavecontext.SetCurrentState(m.Ctx, state), m.bot), cmd
		}

		state.ReplaceBotConfig = true
		if state.MinitiaConfig != nil {
			return NewPrefillMinitiaConfig(weavecontext.SetCurrentState(m.Ctx, state)), cmd
		}
		return NewL1PrefillSelector(weavecontext.SetCurrentState(m.Ctx, state)), cmd
	}

	return m, cmd
}

func (m *DeleteDBSelector) View() string {
	state := weavecontext.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + m.Selector.View()
}

type UseCurrentConfigSelector struct {
	ui.Selector[string]
	weavecontext.BaseModel
	question   string
	configPath string
}

func NewUseCurrentConfigSelector(ctx context.Context, bot string) *UseCurrentConfigSelector {
	configPath := fmt.Sprintf("%s/%s.json", weavecontext.GetOPInitHome(ctx), bot)
	return &UseCurrentConfigSelector{
		Selector: ui.Selector[string]{
			Options: []string{
				"use current file",
				"replace",
			},
		},
		BaseModel:  weavecontext.BaseModel{Ctx: ctx},
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[OPInitBotsState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{m.configPath}, *selected))
		switch *selected {
		case "use current file":
			state.ReplaceBotConfig = false
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			model := NewStartingInitBot(m.Ctx)
			return model, model.Init()
		case "replace":
			state.ReplaceBotConfig = true
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			if state.MinitiaConfig != nil {
				return NewPrefillMinitiaConfig(m.Ctx), cmd
			}
			if state.InitExecutorBot || state.InitChallengerBot {
				return NewL1PrefillSelector(m.Ctx), cmd
			}
			return m, cmd
		}
	}

	return m, cmd
}

func (m *UseCurrentConfigSelector) View() string {
	state := weavecontext.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{m.configPath}, styles.Question) + m.Selector.View()
}

type PrefillMinitiaConfigOption string

const (
	PrefillMinitiaConfigYes = "Yes, prefill"
	PrefillMinitiaConfigNo  = "No, skip"
)

type PrefillMinitiaConfig struct {
	ui.Selector[PrefillMinitiaConfigOption]
	weavecontext.BaseModel
	question string
}

func NewPrefillMinitiaConfig(ctx context.Context) *PrefillMinitiaConfig {
	return &PrefillMinitiaConfig{
		Selector: ui.Selector[PrefillMinitiaConfigOption]{
			Options: []PrefillMinitiaConfigOption{
				PrefillMinitiaConfigYes,
				PrefillMinitiaConfigNo,
			},
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Existing %s detected. Would you like to use the data in this file to pre-fill some fields?", weavecontext.GetMinitiaArtifactsConfigJson(ctx)),
	}
}

func (m *PrefillMinitiaConfig) GetQuestion() string {
	return m.question
}

func (m *PrefillMinitiaConfig) Init() tea.Cmd {
	return nil
}

func (m *PrefillMinitiaConfig) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[OPInitBotsState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{weavecontext.GetMinitiaArtifactsConfigJson(m.Ctx)}, string(*selected)))

		switch *selected {
		case PrefillMinitiaConfigYes:
			minitiaConfig := state.MinitiaConfig
			state.botConfig["l1_node.chain_id"] = minitiaConfig.L1Config.ChainID
			state.botConfig["l1_node.rpc_address"] = minitiaConfig.L1Config.RpcUrl
			state.botConfig["l1_node.gas_price"] = minitiaConfig.L1Config.GasPrices
			GetField(defaultExecutorFields, "l1_node.rpc_address").PrefillValue = minitiaConfig.L1Config.RpcUrl
			GetField(defaultExecutorFields, "l2_node.chain_id").PrefillValue = minitiaConfig.L2Config.ChainID
			GetField(defaultExecutorFields, "l2_node.gas_price").PrefillValue = "0.015" + minitiaConfig.L2Config.Denom
			GetField(defaultExecutorFields, "l2_node.gas_price").Placeholder = "Press tab to use " + "\"0.015" + minitiaConfig.L2Config.Denom + "\""

			GetField(defaultChallengerFields, "l1_node.rpc_address").PrefillValue = minitiaConfig.L1Config.RpcUrl
			GetField(defaultChallengerFields, "l2_node.chain_id").PrefillValue = minitiaConfig.L2Config.ChainID

			if minitiaConfig.OpBridge.BatchSubmissionTarget == "CELESTIA" {
				var network registry.ChainType
				if registry.MustGetChainRegistry(registry.InitiaL1Testnet).GetChainId() == minitiaConfig.L1Config.ChainID {
					network = registry.CelestiaTestnet
				} else {
					network = registry.CelestiaMainnet
				}

				chainRegistry := registry.MustGetChainRegistry(network)

				state.botConfig["da_node.chain_id"] = chainRegistry.GetChainId()
				state.botConfig["da_node.rpc_address"] = chainRegistry.MustGetActiveRpc()
				state.botConfig["da_node.bech32_prefix"] = chainRegistry.GetBech32Prefix()
				state.botConfig["da_node.gas_price"] = DefaultCelestiaGasPrices
				state.daIsCelestia = true
			} else {
				state.botConfig["da_node.chain_id"] = state.botConfig["l1_node.chain_id"]
				state.botConfig["da_node.rpc_address"] = state.botConfig["l1_node.rpc_address"]
				state.botConfig["da_node.bech32_prefix"] = "init"
				state.botConfig["da_node.gas_price"] = state.botConfig["l1_node.gas_price"]
			}
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			if state.InitExecutorBot {
				return NewFieldInputModel(m.Ctx, defaultExecutorFields, NewStartingInitBot), cmd
			} else if state.InitChallengerBot {
				return NewFieldInputModel(m.Ctx, defaultChallengerFields, NewStartingInitBot), cmd

			}
		case PrefillMinitiaConfigNo:
			m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
			return NewL1PrefillSelector(m.Ctx), cmd
		}

	}

	return m, cmd
}

func (m *PrefillMinitiaConfig) View() string {
	state := weavecontext.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{weavecontext.GetMinitiaArtifactsConfigJson(m.Ctx)}, styles.Question) + m.Selector.View()
}

type L1PrefillOption string

var (
	L1PrefillOptionTestnet L1PrefillOption = ""
	L1PrefillOptionCustom  L1PrefillOption = "Custom"
)

type L1PrefillSelector struct {
	ui.Selector[L1PrefillOption]
	weavecontext.BaseModel
	question string
}

func NewL1PrefillSelector(ctx context.Context) *L1PrefillSelector {
	initiaTestnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
	L1PrefillOptionTestnet = L1PrefillOption(fmt.Sprintf("Testnet (%s)", initiaTestnetRegistry.GetChainId()))
	return &L1PrefillSelector{
		Selector: ui.Selector[L1PrefillOption]{
			Options: []L1PrefillOption{
				L1PrefillOptionTestnet,
				L1PrefillOptionCustom,
			},
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[OPInitBotsState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"L1"}, string(*selected)))

		var chainId, rpc, minGasPrice string
		switch *selected {
		case L1PrefillOptionTestnet:
			chainRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
			chainId = chainRegistry.GetChainId()
			rpc = chainRegistry.MustGetActiveRpc()
			minGasPrice = chainRegistry.MustGetMinGasPriceByDenom(DefaultInitiaGasDenom)
		}

		state.botConfig["l1_node.chain_id"] = chainId
		state.botConfig["l1_node.gas_price"] = minGasPrice

		GetField(defaultExecutorFields, "l1_node.rpc_address").PrefillValue = rpc
		GetField(defaultChallengerFields, "l1_node.rpc_address").PrefillValue = rpc

		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		if state.InitExecutorBot {
			return NewFieldInputModel(m.Ctx, defaultExecutorFields, NewSetDALayer), cmd
		} else if state.InitChallengerBot {
			return NewFieldInputModel(m.Ctx, defaultChallengerFields, NewStartingInitBot), cmd
		}

	}

	return m, cmd
}

func (m *L1PrefillSelector) View() string {
	state := weavecontext.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L1"}, styles.Question) + m.Selector.View()
}

type DALayerNetwork string

const (
	Initia   DALayerNetwork = "Initia"
	Celestia DALayerNetwork = "Celestia"
	// Add other types as needed
)

type SetDALayer struct {
	ui.Selector[DALayerNetwork]
	weavecontext.BaseModel
	question string

	chainRegistry *registry.ChainRegistry
}

func NewSetDALayer(ctx context.Context) tea.Model {
	tooltips := []ui.Tooltip{
		InitiaDALayerTooltip,
	}
	state := weavecontext.GetCurrentState[OPInitBotsState](ctx)
	var network registry.ChainType
	if registry.MustGetChainRegistry(registry.InitiaL1Testnet).GetChainId() == state.botConfig["l1_node.chain_id"] {
		network = registry.CelestiaTestnet
		tooltips = append(tooltips, CelestiaTestnetDALayerTooltip)
	} else {
		network = registry.CelestiaMainnet
		tooltips = append(tooltips, CelestiaMainnetDALayerTooltip)
	}

	chainRegistry := registry.MustGetChainRegistry(network)

	return &SetDALayer{
		Selector: ui.Selector[DALayerNetwork]{
			Options: []DALayerNetwork{
				Initia,
				Celestia,
			},
			CannotBack: true,
			Tooltips:   &tooltips,
		},
		BaseModel:     weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:      "Which DA Layer would you like to use?",
		chainRegistry: chainRegistry,
	}
}

func (m *SetDALayer) GetQuestion() string {
	return m.question
}

func (m *SetDALayer) Init() tea.Cmd {
	return nil
}

func (m *SetDALayer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[OPInitBotsState](m, msg); handled {
		return model, cmd
	}
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[OPInitBotsState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"DA Layer"}, string(*selected)))
		switch *selected {
		case Initia:
			state.botConfig["da_node.chain_id"] = state.botConfig["l1_node.chain_id"]
			state.botConfig["da_node.rpc_address"] = state.botConfig["l1_node.rpc_address"]
			state.botConfig["da_node.bech32_prefix"] = "init"
			state.botConfig["da_node.gas_price"] = state.botConfig["l1_node.gas_price"]
		case Celestia:
			state.botConfig["da_node.chain_id"] = m.chainRegistry.GetChainId()
			state.botConfig["da_node.rpc_address"] = m.chainRegistry.MustGetActiveRpc()
			state.botConfig["da_node.bech32_prefix"] = m.chainRegistry.GetBech32Prefix()
			state.botConfig["da_node.gas_price"] = DefaultCelestiaGasPrices
			state.daIsCelestia = true
		}
		model := NewStartingInitBot(weavecontext.SetCurrentState(m.Ctx, state))
		return model, model.Init()

	}

	return m, cmd
}

func (m *SetDALayer) View() string {
	state := weavecontext.GetCurrentState[OPInitBotsState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"DA Layer"}, styles.Question) + m.Selector.View()
}

type StartingInitBot struct {
	weavecontext.BaseModel
	loading ui.Loading
}

func NewStartingInitBot(ctx context.Context) tea.Model {
	state := weavecontext.GetCurrentState[OPInitBotsState](ctx)
	var bot string
	if state.InitExecutorBot {
		bot = "executor"
	} else {
		bot = "challenger"
	}

	return &StartingInitBot{
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   ui.NewLoading(fmt.Sprintf("Setting up OPinit bot %s...", bot), WaitStartingInitBot(ctx)),
	}
}

func WaitStartingInitBot(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1500 * time.Millisecond)

		state := weavecontext.GetCurrentState[OPInitBotsState](ctx)
		configMap := state.botConfig

		if state.isDeleteDB {
			err := io.DeleteDirectory(state.dbPath)
			if err != nil {
				panic(err)
			}
		}

		opInitHome := weavecontext.GetOPInitHome(ctx)
		weaveDummyKeyPath := filepath.Join(opInitHome, "weave-dummy")
		l1KeyPath := filepath.Join(opInitHome, configMap["l1_node.chain_id"])
		l2KeyPath := filepath.Join(opInitHome, configMap["l2_node.chain_id"])

		err := io.CopyDirectory(weaveDummyKeyPath, l1KeyPath)
		if err != nil {
			panic(err)
		}
		err = io.CopyDirectory(weaveDummyKeyPath, l2KeyPath)
		if err != nil {
			panic(err)
		}

		if state.daIsCelestia {
			daKeyPath := filepath.Join(opInitHome, configMap["da_node.chain_id"])
			err = io.CopyDirectory(weaveDummyKeyPath, daKeyPath)
			if err != nil {
				panic(err)
			}
		}

		// TODO: Remove these once our rpcs are compatible
		if strings.Contains(configMap["l1_node.rpc_address"], "initia.xyz") {
			configMap["l1_node.rpc_address"] = DefaultInitiaL1Rpc
		}
		if strings.Contains(configMap["da_node.rpc_address"], "initia.xyz") {
			configMap["da_node.rpc_address"] = DefaultInitiaL1Rpc
		}

		if state.InitExecutorBot {
			srv, err := service.NewService(service.OPinitExecutor)
			if err != nil {
				panic(fmt.Sprintf("failed to initialize service: %v", err))
			}

			if err = srv.Create("", opInitHome); err != nil {
				panic(fmt.Sprintf("failed to create service: %v", err))
			}

			if !state.ReplaceBotConfig {
				return ui.EndLoading{}
			}

			version := registry.MustGetOPInitBotsSpecVersion(state.botConfig["l1_node.chain_id"])

			config := ExecutorConfig{
				Version: version,
				Server: ServerConfig{
					Address:      configMap["listen_address"],
					AllowOrigins: "*",
					AllowHeaders: "Origin, Content-Type, Accept",
					AllowMethods: "GET",
				},
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
					ChainID:       configMap["da_node.chain_id"],
					RPCAddress:    configMap["da_node.rpc_address"],
					Bech32Prefix:  configMap["da_node.bech32_prefix"],
					GasPrice:      configMap["da_node.gas_price"],
					GasAdjustment: 1.5,
					TxTimeout:     60,
				},
				BridgeExecutor:                BridgeExecutorKeyName,
				OracleBridgeExecutor:          OracleBridgeExecutorKeyName,
				MaxChunks:                     5000,
				MaxChunkSize:                  300000,
				MaxSubmissionTime:             3600,
				L2StartHeight:                 0,
				BatchStartHeight:              0,
				DisableDeleteFutureWithdrawal: false,
				DisableAutoSetL1Height:        false,
				DisableBatchSubmitter:         false,
				DisableOutputSubmitter:        false,
			}
			configBz, err := json.MarshalIndent(config, "", " ")
			if err != nil {
				panic(fmt.Errorf("failed to marshal config: %v", err))
			}

			configFilePath := filepath.Join(opInitHome, "executor.json")
			if err = os.WriteFile(configFilePath, configBz, 0600); err != nil {
				panic(fmt.Errorf("failed to write config file: %v", err))
			}
		} else if state.InitChallengerBot {
			srv, err := service.NewService(service.OPinitChallenger)
			if err != nil {
				panic(fmt.Sprintf("failed to initialize service: %v", err))
			}

			if err = srv.Create("", opInitHome); err != nil {
				panic(fmt.Sprintf("failed to create service: %v", err))
			}

			if !state.ReplaceBotConfig {
				return ui.EndLoading{}
			}

			version := registry.MustGetOPInitBotsSpecVersion(state.botConfig["l1_node.chain_id"])
			config := ChallengerConfig{
				Version: version,
				Server: ServerConfig{
					Address:      configMap["listen_address"],
					AllowOrigins: "*",
					AllowHeaders: "Origin, Content-Type, Accept",
					AllowMethods: "GET",
				},
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

			configFilePath := filepath.Join(opInitHome, "challenger.json")
			if err = os.WriteFile(configFilePath, configBz, 0600); err != nil {
				panic(fmt.Errorf("failed to write config file: %v", err))
			}
		}
		return ui.EndLoading{}
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
	state := weavecontext.GetCurrentState[OPInitBotsState](m.Ctx)
	return state.weave.Render() + m.loading.View()
}

type OPinitBotSuccessful struct {
	weavecontext.BaseModel
}

func NewOPinitBotSuccessful(ctx context.Context) *OPinitBotSuccessful {
	return &OPinitBotSuccessful{
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *OPinitBotSuccessful) Init() tea.Cmd {
	return nil
}

func (m *OPinitBotSuccessful) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m *OPinitBotSuccessful) View() string {
	state := weavecontext.GetCurrentState[OPInitBotsState](m.Ctx)

	botConfigFileName := "executor"
	if state.InitChallengerBot {
		botConfigFileName = "challenger"
	}

	return state.weave.Render() + styles.RenderPrompt(fmt.Sprintf("OPInit bot setup successfully. Config file is saved at %s. Feel free to modify it as needed.", filepath.Join(weavecontext.GetOPInitHome(m.Ctx), fmt.Sprintf("%s.json", botConfigFileName))), []string{}, styles.Completed) + "\n" + styles.RenderPrompt("You can start the bot by running `weave opinit-bots start "+botConfigFileName+"`", []string{}, styles.Completed) + "\n"
}

// SetupOPInitBotsMissingKey handles the loading and setup of OPInit bots
type SetupOPInitBotsMissingKey struct {
	weavecontext.BaseModel
	loading ui.Loading
}

// NewSetupOPInitBots initializes a new SetupOPInitBots with context
func NewSetupOPInitBotsMissingKey(ctx context.Context) *SetupOPInitBotsMissingKey {
	return &SetupOPInitBotsMissingKey{
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		loading:   ui.NewLoading("Downloading binary and adding keys...", WaitSetupOPInitBotsMissingKey(ctx)),
	}
}

func (m *SetupOPInitBotsMissingKey) Init() tea.Cmd {
	return m.loading.Init()
}

func (m *SetupOPInitBotsMissingKey) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				state := weavecontext.GetCurrentState[OPInitBotsState](m.loading.EndContext)
				if state.InitExecutorBot {
					return OPInitBotInitSelectExecutor(m.loading.EndContext), nil
				} else if state.InitChallengerBot {
					return OPInitBotInitSelectChallenger(m.loading.EndContext), nil
				}
			}
		}
	}
	return m, cmd
}

func (m *SetupOPInitBotsMissingKey) View() string {
	state := weavecontext.GetCurrentState[OPInitBotsState](m.Ctx)
	if len(state.SetupOpinitResponses) > 0 {
		mnemonicText := ""
		for _, botName := range BotNames {
			if res, ok := state.SetupOpinitResponses[botName]; ok {
				keyInfo := strings.Split(res, "\n")
				address := strings.Split(keyInfo[0], ": ")
				mnemonicText += renderMnemonic(string(botName), address[1], keyInfo[1])
			}
		}

		return state.weave.Render() + "\n" + styles.BoldUnderlineText("Important", styles.Yellow) + "\n" +
			styles.Text("Write down these mnemonic phrases and store them in a safe place. \nIt is the only way to recover your system keys.", styles.Yellow) + "\n\n" +
			mnemonicText + "\nPress enter to go next step\n"
	}
	return state.weave.Render() + "\n"
}

func WaitSetupOPInitBotsMissingKey(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := weavecontext.GetCurrentState[OPInitBotsState](ctx)
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("failed to get user home directory: %v", err))
		}

		binaryPath := filepath.Join(userHome, common.WeaveDataDirectory, fmt.Sprintf("opinitd@%s", OpinitBotBinaryVersion), AppName)

		opInitHome := weavecontext.GetOPInitHome(ctx)
		for _, info := range state.BotInfos {
			if info.Mnemonic != "" {
				res, err := cosmosutils.OPInitRecoverKeyFromMnemonic(binaryPath, info.KeyName, info.Mnemonic, info.DALayer == string(CelestiaLayerOption), opInitHome)
				if err != nil {
					return ui.ErrorLoading{Err: err}
				}
				state.SetupOpinitResponses[info.BotName] = res
				continue
			}
			if info.IsGenerateKey {
				res, err := cosmosutils.OPInitAddOrReplace(binaryPath, info.KeyName, info.DALayer == string(CelestiaLayerOption), opInitHome)
				if err != nil {
					return ui.ErrorLoading{Err: err}

				}
				state.SetupOpinitResponses[info.BotName] = res
				continue
			}
		}

		return ui.EndLoading{
			Ctx: ctx,
		}
	}
}
