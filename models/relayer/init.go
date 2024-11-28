package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/common"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/cosmosutils"
	"github.com/initia-labs/weave/io"
	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/ui"
)

var defaultL2ConfigLocal = []*Field{
	{Name: "l2.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", DefaultValue: "http://localhost:26657", ValidateFn: common.ValidateURL},
	{Name: "l2.grpc_address", Type: StringField, Question: "Please specify the L2 grpc_address", Placeholder: "Add RPC address ex. http://localhost:9090", DefaultValue: "http://localhost:9090", ValidateFn: common.ValidateURL},
	{Name: "l2.websocket", Type: StringField, Question: "Please specify the L2 websocket", Placeholder: "Add RPC address ex. ws://localhost:26657/websocket", DefaultValue: "ws://localhost:26657/websocket", ValidateFn: common.ValidateURL},
}

var defaultL2ConfigManual = []*Field{
	{Name: "l2.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: common.ValidateEmptyString},
	{Name: "l2.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", ValidateFn: common.ValidateURL},
	{Name: "l2.grpc_address", Type: StringField, Question: "Please specify the L2 grpc_address", Placeholder: "Add RPC address ex. http://localhost:9090", ValidateFn: common.ValidateURL},
	{Name: "l2.websocket", Type: StringField, Question: "Please specify the L2 websocket", Placeholder: "Add RPC address ex. ws://localhost:26657/websocket", ValidateFn: common.ValidateURL},
	{Name: "l2.gas_price.denom", Type: StringField, Question: "Please specify the gas_price denom", Placeholder: "Add gas_price denom ex. umin", ValidateFn: common.ValidateDenom},
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
	options := make([]RollupSelectOption, 0)
	if io.FileOrFolderExists(weavecontext.GetMinitiaArtifactsConfigJson(ctx)) {
		options = append(options, Whitelisted, Local, Manual)
	} else {
		options = append(options, Whitelisted, Manual)
	}

	return &RollupSelect{
		Selector: ui.Selector[RollupSelectOption]{
			Options:    options,
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
			return NewSelectingL1NetworkRegistry(weavecontext.SetCurrentState(m.Ctx, state)), nil
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
				state.Config["l1.rpc_address"] = testnetRegistry.MustGetActiveRpc()
				state.Config["l1.grpc_address"] = testnetRegistry.MustGetActiveGrpc()
				state.Config["l1.lcd_address"] = testnetRegistry.MustGetActiveLcd()
				state.Config["l1.gas_price.denom"] = DefaultGasPriceDenom
				state.Config["l1.gas_price.price"] = testnetRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)

				state.Config["l2.chain_id"] = minitiaConfig.L2Config.ChainID
				state.Config["l2.gas_price.denom"] = DefaultGasPriceDenom
				state.Config["l2.gas_price.price"] = DefaultGasPriceAmount
				state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, "L1 network is auto-detected", []string{}, "initiation-2"))

			} else {
				panic("not support L1")
			}
			return NewL1KeySelect(weavecontext.SetCurrentState(m.Ctx, state)), nil
			//return NewFieldInputModel(weavecontext.SetCurrentState(m.Ctx, state), defaultL2ConfigLocal, NewSelectSettingUpIBCChannelsMethod), nil
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
	chainId  string
}

type L1KeySelectOption string

const (
	L1GenerateKey L1KeySelectOption = "Generate new system key"
)

var (
	L1ExistingKey = L1KeySelectOption("Use an existing key " + styles.Text("(previously setup in Weave)", styles.Gray))
	L1ImportKey   = L1KeySelectOption("Import existing key " + styles.Text("(you will be prompted to enter your mnemonic)", styles.Gray))
)

func NewL1KeySelect(ctx context.Context) *L1KeySelect {
	l1ChainId, ok := GetL1ChainId(ctx)
	if !ok {
		panic("cannot get l1 chain id")
	}

	options := []L1KeySelectOption{
		L1GenerateKey,
		L1ImportKey,
	}
	// TODO: fix appname
	if l1RelayerAddress, found := cosmosutils.GetHermesRelayerAddress("hermes", l1ChainId); found {
		state := weavecontext.GetCurrentState[RelayerState](ctx)
		state.l1RelayerAddress = l1RelayerAddress
		options = append([]L1KeySelectOption{L1ExistingKey}, options...)
	}

	return &L1KeySelect{
		Selector: ui.Selector[L1KeySelectOption]{
			Options:    options,
			CannotBack: true,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  fmt.Sprintf("Please select an option for setting up the relayer account key on L1 (%s)", l1ChainId),
		chainId:   l1ChainId,
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

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"relayer account key", fmt.Sprintf("L1 (%s)", m.chainId)}, string(*selected)))
		state.l1KeyMethod = string(*selected)
		return NewL2KeySelect(weavecontext.SetCurrentState(m.Ctx, state)), nil
	}

	return m, cmd
}

func (m *L1KeySelect) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + "\n" + styles.InformationMark + styles.BoldText(
		"Relayer account keys with funds",
		styles.White,
	) + " are required to setup and run the relayer properly." + "\n" + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"relayer account key", fmt.Sprintf("L1 (%s)", m.chainId)},
		styles.Question,
	) + m.Selector.View()
}

type L2KeySelect struct {
	ui.Selector[L2KeySelectOption]
	weavecontext.BaseModel
	question string
	chainId  string
}

type L2KeySelectOption string

const (
	L2SameKey     L2KeySelectOption = "Use the same key with L1"
	L2GenerateKey L2KeySelectOption = "Generate new system key"
)

var (
	L2ExistingKey = L2KeySelectOption("Use an existing key " + styles.Text("(previously setup in Weave)", styles.Gray))
	L2ImportKey   = L2KeySelectOption("Import existing key " + styles.Text("(you will be prompted to enter your mnemonic)", styles.Gray))
)

func NewL2KeySelect(ctx context.Context) *L2KeySelect {
	l2ChainId, ok := GetL2ChainId(ctx)
	if !ok {
		panic("cannot get l2 chain id")
	}

	options := []L2KeySelectOption{
		L2SameKey,
		L2GenerateKey,
		L2ImportKey,
	}
	// TODO: fix appname
	if l2RelayerAddress, found := cosmosutils.GetHermesRelayerAddress("hermes", l2ChainId); found {
		state := weavecontext.GetCurrentState[RelayerState](ctx)
		state.l2RelayerAddress = l2RelayerAddress
		options = append([]L2KeySelectOption{L2ExistingKey}, options...)
	}

	return &L2KeySelect{
		Selector: ui.Selector[L2KeySelectOption]{
			Options: options,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please select an option for setting up the relayer account key on L2 (%s)", l2ChainId),
		chainId:   l2ChainId,
	}
}

func (m *L2KeySelect) GetQuestion() string {
	return m.question
}

func (m *L2KeySelect) Init() tea.Cmd {
	return nil
}

func (m *L2KeySelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"relayer account key", fmt.Sprintf("L2 (%s)", m.chainId)}, string(*selected)))
		state.l2KeyMethod = string(*selected)
		return m, tea.Quit
	}

	return m, cmd
}

func (m *L2KeySelect) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(
		m.GetQuestion(),
		[]string{"relayer account key", fmt.Sprintf("L2 (%s)", m.chainId)},
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
			state.Config["l1.rpc_address"] = testnetRegistry.MustGetActiveRpc()
			state.Config["l1.grpc_address"] = testnetRegistry.MustGetActiveGrpc()
			state.Config["l1.lcd_address"] = testnetRegistry.MustGetActiveLcd()
			state.Config["l1.gas_price.denom"] = DefaultGasPriceDenom
			state.Config["l1.gas_price.price"] = testnetRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)

			return NewFieldInputModel(m.Ctx, defaultL2ConfigManual, NewSelectSettingUpIBCChannelsMethod), nil
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
	// TODO: dynamic network
	networks := registry.MustGetAllL2AvailableNetwork(registry.InitiaL1Testnet)

	var options []string
	for _, network := range networks {
		options = append(options, fmt.Sprintf("%s (%s)", network.PrettyName, network.ChainId))
	}
	sort.Slice(options, func(i, j int) bool { return strings.ToLower(options[i]) < strings.ToLower(options[j]) })

	return &SelectingL2Network{
		Selector: ui.Selector[string]{
			Options: options,
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

type SelectingL1NetworkRegistry struct {
	ui.Selector[NetworkSelectOption]
	weavecontext.BaseModel
	question string
}

func NewSelectingL1NetworkRegistry(ctx context.Context) *SelectingL1NetworkRegistry {
	testnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
	//mainnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Mainnet)
	Testnet = NetworkSelectOption(fmt.Sprintf("Testnet (%s)", testnetRegistry.GetChainId()))
	//Mainnet = NetworkSelectOption(fmt.Sprintf("Mainnet (%s)", mainnetRegistry.GetChainId()))
	return &SelectingL1NetworkRegistry{
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

func (m *SelectingL1NetworkRegistry) GetQuestion() string {
	return m.question
}

func (m *SelectingL1NetworkRegistry) Init() tea.Cmd {
	return nil
}

func (m *SelectingL1NetworkRegistry) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			state.Config["l1.rpc_address"] = testnetRegistry.MustGetActiveRpc()
			state.Config["l1.grpc_address"] = testnetRegistry.MustGetActiveGrpc()
			state.Config["l1.lcd_address"] = testnetRegistry.MustGetActiveLcd()
			state.Config["l1.gas_price.denom"] = DefaultGasPriceDenom
			state.Config["l1.gas_price.price"] = testnetRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)

			return NewSelectingL2Network(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
	}

	return m, cmd
}

func (m *SelectingL1NetworkRegistry) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"Initia L1 network"}, styles.Question) + m.Selector.View()
}

type SettingUpIBCChannelOption string

var (
	Basic       SettingUpIBCChannelOption = "Open IBC Channels for transfer and nft-transfer automatically"
	FillFromLCD SettingUpIBCChannelOption = "Fill in L2 LCD endpoint to detect available IBC Channel pairs"
	Manually    SettingUpIBCChannelOption = "Setup each IBC Channel manually"
)

type SelectSettingUpIBCChannelsMethod struct {
	ui.Selector[SettingUpIBCChannelOption]
	weavecontext.BaseModel
	question string
}

func NewSelectSettingUpIBCChannelsMethod(ctx context.Context) tea.Model {
	options := make([]SettingUpIBCChannelOption, 0)
	options = append(options, Basic)
	options = append(options, FillFromLCD, Manually)

	return &SelectSettingUpIBCChannelsMethod{
		Selector: ui.Selector[SettingUpIBCChannelOption]{
			Options: options,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Please select method to setup IBC channels for the relayer.",
	}
}

func (m *SelectSettingUpIBCChannelsMethod) GetQuestion() string {
	return m.question
}

func (m *SelectSettingUpIBCChannelsMethod) Init() tea.Cmd {
	return nil
}

func (m *SelectSettingUpIBCChannelsMethod) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	// Handle selection logic
	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{""}, string(*selected)))
		switch *selected {
		case Basic:

			// Read the file content
			data, err := os.ReadFile(weavecontext.GetMinitiaArtifactsJson(m.Ctx))
			if err != nil {
				panic(err)
			}
			// Decode the JSON into a struct
			var artifacts types.Artifacts
			if err := json.Unmarshal(data, &artifacts); err != nil {
				panic(err)
			}

			// TODO: switch registry
			testnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
			info := testnetRegistry.MustGetOpinitBridgeInfo(artifacts.BridgeID)
			metadata := types.MustDecodeBridgeMetadata(info.BridgeConfig.Metadata)

			channelPairs := make([]types.IBCChannelPair, 0)
			for _, channel := range metadata.PermChannels {
				counterparty := testnetRegistry.MustGetCounterPartyIBCChannel(channel.PortID, channel.ChannelID)
				channelPairs = append(channelPairs, types.IBCChannelPair{
					L1: channel,
					L2: counterparty,
				})
			}
			return NewIBCChannelsCheckbox(weavecontext.SetCurrentState(m.Ctx, state), channelPairs), nil
		case FillFromLCD:
			return NewFillL2LCD(weavecontext.SetCurrentState(m.Ctx, state)), nil
		case Manually:
			return NewFillPortOnL1(weavecontext.SetCurrentState(m.Ctx, state), 0), nil
		}
	}

	return m, cmd
}

func (m *SelectSettingUpIBCChannelsMethod) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{""}, styles.Question) + m.Selector.View()
}

func GetL1ChainId(ctx context.Context) (string, bool) {
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	if chainId, ok := state.Config["l1.chain_id"]; ok {
		return chainId, ok
	}
	return "", false
}

func GetL2ChainId(ctx context.Context) (string, bool) {
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	if chainId, found := state.Config["l2.chain_id"]; found {
		return chainId, found
	}
	return "", false
}

type FillPortOnL1 struct {
	weavecontext.BaseModel
	ui.TextInput
	idx      int
	question string
	extra    string
}

func NewFillPortOnL1(ctx context.Context, idx int) *FillPortOnL1 {
	extra := ""
	if chainId, found := GetL1ChainId(ctx); found {
		extra = fmt.Sprintf("(%s)", chainId)
	}
	model := &FillPortOnL1{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify the port name in L1 %s", extra),
		idx:       idx,
		extra:     extra,
	}
	model.WithPlaceholder("ex. transfer")
	return model
}

func (m *FillPortOnL1) GetQuestion() string {
	return m.question
}

func (m *FillPortOnL1) Init() tea.Cmd {
	return nil
}

func (m *FillPortOnL1) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"L1", m.extra}, m.TextInput.Text))
		return NewFillChannelL1(weavecontext.SetCurrentState(m.Ctx, state), m.TextInput.Text, m.idx), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *FillPortOnL1) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L1", m.extra}, styles.Question) + m.TextInput.View()
}

type FillChannelL1 struct {
	weavecontext.BaseModel
	ui.TextInput
	idx      int
	port     string
	question string
	extra    string
}

func NewFillChannelL1(ctx context.Context, port string, idx int) *FillChannelL1 {
	extra := ""
	if chainId, found := GetL1ChainId(ctx); found {
		extra = fmt.Sprintf("(%s)", chainId)
	}
	model := &FillChannelL1{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify the %s channel in L1 %s", port, extra),
		idx:       idx,
		port:      port,
		extra:     extra,
	}
	model.WithPlaceholder("ex. channel-1")
	return model
}

func (m *FillChannelL1) GetQuestion() string {
	return m.question
}

func (m *FillChannelL1) Init() tea.Cmd {
	return nil
}

func (m *FillChannelL1) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"L1", m.port, m.extra}, m.TextInput.Text))

		return NewFillPortOnL2(weavecontext.SetCurrentState(m.Ctx, state), m.idx), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *FillChannelL1) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L1", m.port, m.extra}, styles.Question) + m.TextInput.View()
}

type FillPortOnL2 struct {
	weavecontext.BaseModel
	ui.TextInput
	idx      int
	question string
	extra    string
}

func NewFillPortOnL2(ctx context.Context, idx int) *FillPortOnL2 {
	extra := ""
	if chainId, found := GetL2ChainId(ctx); found {
		extra = fmt.Sprintf("(%s)", chainId)
	}
	model := &FillPortOnL2{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify the port name in L2 %s", extra),
		idx:       idx,
		extra:     extra,
	}
	model.WithPlaceholder("ex. transfer")
	return model
}

func (m *FillPortOnL2) GetQuestion() string {
	return m.question
}

func (m *FillPortOnL2) Init() tea.Cmd {
	return nil
}

func (m *FillPortOnL2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"L2", m.extra}, m.TextInput.Text))

		return NewFillChannelL2(weavecontext.SetCurrentState(m.Ctx, state), m.TextInput.Text, m.idx), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *FillPortOnL2) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2", m.extra}, styles.Question) + m.TextInput.View()
}

type FillChannelL2 struct {
	weavecontext.BaseModel
	ui.TextInput
	idx      int
	port     string
	extra    string
	question string
}

func NewFillChannelL2(ctx context.Context, port string, idx int) *FillChannelL2 {
	extra := ""
	if chainId, found := GetL2ChainId(ctx); found {
		extra = fmt.Sprintf("(%s)", chainId)
	}
	model := &FillChannelL2{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify the %s channel in L2 %s", port, extra),
		idx:       idx,
		port:      port,
		extra:     extra,
	}
	model.WithPlaceholder("ex. channel-1")
	return model
}

func (m *FillChannelL2) GetQuestion() string {
	return m.question
}

func (m *FillChannelL2) Init() tea.Cmd {
	return nil
}

func (m *FillChannelL2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"L2", m.port, m.extra}, m.TextInput.Text))
		return NewAddMoreIBCChannels(weavecontext.SetCurrentState(m.Ctx, state), m.idx), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *FillChannelL2) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2", m.port, m.extra}, styles.Question) + m.TextInput.View()
}

type AddMoreIBCChannels struct {
	ui.Selector[string]
	weavecontext.BaseModel
	question string
	idx      int
}

func NewAddMoreIBCChannels(ctx context.Context, idx int) *AddMoreIBCChannels {
	return &AddMoreIBCChannels{
		Selector: ui.Selector[string]{
			Options: []string{
				"Yes",
				"No",
			},
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Do you want to open more IBC Channels?",
		idx:       idx,
	}
}

func (m *AddMoreIBCChannels) GetQuestion() string {
	return m.question
}

func (m *AddMoreIBCChannels) Init() tea.Cmd {
	return nil
}

func (m *AddMoreIBCChannels) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{""}, string(*selected)))

		return NewFillPortOnL1(weavecontext.SetCurrentState(m.Ctx, state), m.idx), nil
	}
	return m, cmd
}

func (m *AddMoreIBCChannels) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{""}, styles.Question) + m.Selector.View()
}

type IBCChannelsCheckbox struct {
	ui.CheckBox[string]
	weavecontext.BaseModel
	question string
	pairs    []types.IBCChannelPair
}

func NewIBCChannelsCheckbox(ctx context.Context, pairs []types.IBCChannelPair) *IBCChannelsCheckbox {
	prettyPairs := []string{"Open all IBC channels"}
	for _, pair := range pairs {
		prettyPairs = append(prettyPairs, fmt.Sprintf("(L1) %s : %s ◀ ▶︎ (L2) %s : %s", pair.L1.PortID, pair.L1.ChannelID, pair.L2.PortID, pair.L2.ChannelID))
	}
	cb := ui.NewCheckBox(prettyPairs)
	cb.EnableSelectAll()
	pairs = append([]types.IBCChannelPair{pairs[0]}, pairs...)
	return &IBCChannelsCheckbox{
		CheckBox:  *cb,
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Please select the IBC channels you would like to open",
		pairs:     pairs,
	}
}

func (m *IBCChannelsCheckbox) GetQuestion() string {
	return m.question
}

func (m *IBCChannelsCheckbox) Init() tea.Cmd {
	return nil
}

func (m *IBCChannelsCheckbox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}
	cb, cmd, done := m.Select(msg)
	_ = cb
	if done {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		ibcChannels := make([]types.IBCChannelPair, 0)
		for idx := 1; idx < len(m.pairs); idx++ {
			if m.Selected[idx] {
				ibcChannels = append(ibcChannels, m.pairs[idx])
			}
			state.IBCChannels = ibcChannels
		}
		// TODO: setup
	}
	return m, cmd
}

func (m *IBCChannelsCheckbox) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + "\n" + m.CheckBox.View()
}

type FillL2LCD struct {
	weavecontext.BaseModel
	ui.TextInput
	question string
	extra    string
}

func NewFillL2LCD(ctx context.Context) *FillL2LCD {
	chainId, _ := GetL2ChainId(ctx)
	extra := fmt.Sprintf("(%s)", chainId)
	return &FillL2LCD{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please specify the L2 LCD_address %s", extra),
		extra:     extra,
	}
}

func (m *FillL2LCD) GetQuestion() string {
	return m.question
}

func (m *FillL2LCD) Init() tea.Cmd {
	return nil
}

func (m *FillL2LCD) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"L2", "LCD_address", m.extra}, m.TextInput.Text))

		// TODO: if cannot get channels should show error
		// return NewFillChannelL1(weavecontext.SetCurrentState(m.Ctx, state), m.TextInput.Text, m.idx), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *FillL2LCD) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2", "LCD_address", m.extra}, styles.Question) + m.TextInput.View()
}
