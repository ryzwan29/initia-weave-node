package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/common"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/ui"
)

var defaultL2ConfigLocal = []*Field{
	{Name: "l2.rpc", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", DefaultValue: "http://localhost:26657", ValidateFn: common.ValidateURL},
	{Name: "l2.grpc", Type: StringField, Question: "Please specify the L2 grpc_address", Placeholder: "Add RPC address ex. http://localhost:9090", DefaultValue: "http://localhost:9090", ValidateFn: common.ValidateURL},
	{Name: "l2.websocket", Type: StringField, Question: "Please specify the L2 websocket", Placeholder: "Add RPC address ex. ws://localhost:26657/websocket", DefaultValue: "", ValidateFn: common.ValidateURL},
}

var defaultL2ConfigManaul = []*Field{
	{Name: "l2.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: common.ValidateEmptyString},
	{Name: "l2.rpc", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", ValidateFn: common.ValidateURL},
	{Name: "l2.grpc", Type: StringField, Question: "Please specify the L2 grpc_address", Placeholder: "Add RPC address ex. http://localhost:9090", ValidateFn: common.ValidateURL},
	{Name: "l2.websocket", Type: StringField, Question: "Please specify the L2 websocket", Placeholder: "Add RPC address ex. ws://localhost:26657/websocket", ValidateFn: common.ValidateURL},
}

type RollupSelect struct {
	ui.Selector[RollupSelectOption]
	weavecontext.BaseModel
	question string
}

type RollupSelectOption string

const (
	Whitelisted RollupSelectOption = "Whitelisted Interwoven Rollups"
	Local       RollupSelectOption = "Local Interwoven Rollups"
	Manual      RollupSelectOption = "Manual Relayer Setup"
)

func NewRollupSelect(ctx context.Context) *RollupSelect {
	return &RollupSelect{
		Selector: ui.Selector[RollupSelectOption]{
			Options: []RollupSelectOption{
				Whitelisted,
				Local,
				Manual,
			},
			CannotBack: true,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please select the type of Interwoven Rollups you want to start a Relayer",
	}
}

func (m *RollupSelect) GetQuestion() string {
	return m.question
}

func (m *RollupSelect) Init() tea.Cmd {
	return nil
}

func (m *RollupSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, string(*selected)))
		switch *selected {
		case Whitelisted:
			return NewSelectingL2Network(weavecontext.SetCurrentState(m.Ctx, state)), nil
		case Local:
			minitiaConfigPath := weavecontext.GetMinitiaArtifactsConfigJson(m.Ctx)
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

			if minitiaConfig.L1Config.ChainID == "initiation-2" {
				testnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
				state.Config["l1.chain_id"] = testnetRegistry.GetChainId()
				state.Config["l1.rpc"] = testnetRegistry.MustGetActiveRpc()
				state.Config["l1.grpc"] = testnetRegistry.MustGetActiveGrpc()
				state.Config["l1.lcd"] = testnetRegistry.MustGetActiveLcd()
				state.Config["l1.gas_price.denom"] = DefaultGasPriceDenom
				state.Config["l1.gas_price.price"] = testnetRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)

				state.Config["l2.chain_id"] = minitiaConfig.L2Config.ChainID
				state.Config["l2.gas_price.denom"] = DefaultGasPriceDenom
				state.Config["l2.gas_price.price"] = DefaultGasPriceAmount

			} else {
				panic("not support L1")
			}
			return NewFieldInputModel(weavecontext.SetCurrentState(m.Ctx, state), defaultL2ConfigLocal, NewTerminalState), nil
		case Manual:
			return NewSelectingL1Network(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
		return m, tea.Quit
	}

	return m, cmd
}

func (m *RollupSelect) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{},
		styles.Question,
	) + m.Selector.View()
}

type L1KeySelect struct {
	ui.Selector[L1KeySelectOption]
	weavecontext.BaseModel
	question string
}

type L1KeySelectOption string

const (
	L1GenerateKey L1KeySelectOption = "Generate new system key"
	L1ImportKey   L1KeySelectOption = "Import existing key (you will be prompted to enter your mnemonic)"
)

func NewL1KeySelect(ctx context.Context) *L1KeySelect {
	return &L1KeySelect{
		Selector: ui.Selector[L1KeySelectOption]{
			Options: []L1KeySelectOption{
				L1GenerateKey,
				L1ImportKey,
			},
			CannotBack: true,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please select an option for setting up the relayer account keys on L1",
	}
}

func (m *L1KeySelect) GetQuestion() string {
	return m.question
}

func (m *L1KeySelect) Init() tea.Cmd {
	return nil
}

func (m *L1KeySelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, string(*selected)))
		// TODO: Implement this
		return m, tea.Quit
	}

	return m, cmd
}

func (m *L1KeySelect) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{},
		styles.Question,
	) + m.Selector.View()
}

type NetworkSelectOption string

func (n NetworkSelectOption) ToChainType() registry.ChainType {
	switch n {
	case Mainnet:
		return registry.InitiaL1Mainnet
	case Testnet:
		return registry.InitiaL1Testnet
	default:
		panic("invalid case for NetworkSelectOption")
	}
}

var (
	Testnet NetworkSelectOption = ""
	Mainnet NetworkSelectOption = ""
)

type SelectingL1Network struct {
	ui.Selector[NetworkSelectOption]
	weavecontext.BaseModel
	question string
}

func NewSelectingL1Network(ctx context.Context) *SelectingL1Network {
	testnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
	//mainnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Mainnet)
	Testnet = NetworkSelectOption(fmt.Sprintf("Testnet (%s)", testnetRegistry.GetChainId()))
	//Mainnet = NetworkSelectOption(fmt.Sprintf("Mainnet (%s)", mainnetRegistry.GetChainId()))
	return &SelectingL1Network{
		Selector: ui.Selector[NetworkSelectOption]{
			Options: []NetworkSelectOption{
				Testnet,
				// Mainnet,
			},
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Which Initia L1 network would you like to connect to?",
	}
}

func (m *SelectingL1Network) GetQuestion() string {
	return m.question
}

func (m *SelectingL1Network) Init() tea.Cmd {
	return nil
}

func (m *SelectingL1Network) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	// Handle selection logic
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"Initia L1 network"}, string(*selected)))
		switch *selected {
		case Testnet:
			testnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
			state.Config["l1.chain_id"] = testnetRegistry.GetChainId()
			state.Config["l1.rpc"] = testnetRegistry.MustGetActiveRpc()
			state.Config["l1.grpc"] = testnetRegistry.MustGetActiveGrpc()
			state.Config["l1.lcd"] = testnetRegistry.MustGetActiveLcd()
			state.Config["l1.gas_price.denom"] = DefaultGasPriceDenom
			state.Config["l1.gas_price.price"] = testnetRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)

			return NewFieldInputModel(m.Ctx, defaultL2ConfigManaul, NewTerminalState), nil
		}
	}

	return m, cmd
}

func (m *SelectingL1Network) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Initia L1 network"}, styles.Question) + m.Selector.View()
}

type SelectingL2Network struct {
	ui.Selector[string]
	weavecontext.BaseModel
	question string
}

func NewSelectingL2Network(ctx context.Context) *SelectingL2Network {
	// TODO: fix network
	contents := registry.MustGetAllL2Contents(registry.InitiaL1Testnet)
	names := make([]string, 0)
	for _, content := range contents {
		if content.Name != "initia" {
			names = append(names, content.Name)
		}
	}

	//Mainnet = NetworkSelectOption(fmt.Sprintf("Mainnet (%s)", mainnetRegistry.GetChainId()))
	return &SelectingL2Network{
		Selector: ui.Selector[string]{
			Options: names,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Please specify the L2 network",
	}
}

func (m *SelectingL2Network) Init() tea.Cmd {
	return nil
}

func (m *SelectingL2Network) GetQuestion() string {
	return m.question
}

func (m *SelectingL2Network) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	// Handle selection logic
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"L2 network"}, string(*selected)))
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		// TODO: implement
	}

	return m, cmd
}

func (m *SelectingL2Network) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2 network"}, styles.Question) + m.Selector.View()
}

type TerminalState struct {
	weavecontext.BaseModel
}

func NewTerminalState(ctx context.Context) tea.Model {
	return &TerminalState{
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
	}
}

func (m *TerminalState) Init() tea.Cmd {
	return nil
}

func (m *TerminalState) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m *TerminalState) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render()
}
