package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/client"
	"github.com/initia-labs/weave/common"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/cosmosutils"
	weaveio "github.com/initia-labs/weave/io"
	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/service"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/ui"
)

var defaultL2ConfigLocal = []*Field{
	{Name: "l2.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: `Press tab to use "http://localhost:26657"`, DefaultValue: "http://localhost:26657", ValidateFn: common.ValidateURL},
	{Name: "l2.grpc_address", Type: StringField, Question: "Please specify the L2 grpc_address", Placeholder: `Press tab to use "http://localhost:9090"`, DefaultValue: "http://localhost:9090", ValidateFn: common.ValidateURL},
	{Name: "l2.websocket", Type: StringField, Question: "Please specify the L2 websocket", Placeholder: `Press tab to use ws://localhost:26657/websocket`, DefaultValue: "ws://localhost:26657/websocket", ValidateFn: common.ValidateURL},
}

var defaultL2ConfigManual = []*Field{
	{Name: "l2.chain_id", Type: StringField, Question: "Please specify the L2 chain_id", Placeholder: "Add alphanumeric", ValidateFn: common.ValidateEmptyString},
	{Name: "l2.rpc_address", Type: StringField, Question: "Please specify the L2 rpc_address", Placeholder: "Add RPC address ex. http://localhost:26657", ValidateFn: common.ValidateURL},
	{Name: "l2.grpc_address", Type: StringField, Question: "Please specify the L2 grpc_address", Placeholder: "Add RPC address ex. http://localhost:9090", ValidateFn: common.ValidateURL},
	{Name: "l2.websocket", Type: StringField, Question: "Please specify the L2 websocket", Placeholder: "Add RPC address ex. ws://localhost:26657/websocket", ValidateFn: common.ValidateURL},
	{Name: "l2.gas_price.denom", Type: StringField, Question: "Please specify the gas_price denom", Placeholder: "Add gas_price denom ex. umin", ValidateFn: common.ValidateDenom},
	{Name: "l2.gas_price.price", Type: StringField, Question: "Please specify the gas_price prie", Placeholder: "Add gas_price price ex. 0.15"},
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
	if weaveio.FileOrFolderExists(weavecontext.GetMinitiaArtifactsConfigJson(ctx)) {
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

			if minitiaConfig.L1Config.ChainID == InitiaTestnetChainId {
				testnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Testnet)
				state.Config["l1.chain_id"] = testnetRegistry.GetChainId()
				state.Config["l1.rpc_address"] = testnetRegistry.MustGetActiveRpc()
				state.Config["l1.grpc_address"] = testnetRegistry.MustGetActiveGrpc()
				state.Config["l1.lcd_address"] = testnetRegistry.MustGetActiveLcd()
				state.Config["l1.websocket"] = testnetRegistry.MustGetActiveWebsocket()
				state.Config["l1.gas_price.denom"] = DefaultGasPriceDenom
				state.Config["l1.gas_price.price"] = testnetRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)

				state.Config["l2.chain_id"] = minitiaConfig.L2Config.ChainID
				state.Config["l2.gas_price.denom"] = minitiaConfig.L2Config.Denom
				state.Config["l2.gas_price.price"] = DefaultGasPriceAmount
				state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, "L1 network is auto-detected", []string{}, minitiaConfig.L1Config.ChainID))

			} else {
				panic("not support L1")
			}

			//return NewL1KeySelect(weavecontext.SetCurrentState(m.Ctx, state)), nil
			return NewFieldInputModel(weavecontext.SetCurrentState(m.Ctx, state), defaultL2ConfigLocal, NewSelectSettingUpIBCChannelsMethod), nil
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
	l1ChainId := MustGetL1ChainId(ctx)
	options := []L1KeySelectOption{
		L1GenerateKey,
		L1ImportKey,
	}
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	if l1RelayerAddress, found := cosmosutils.GetHermesRelayerAddress(state.hermesBinaryPath, l1ChainId); found {
		state.l1RelayerAddress = l1RelayerAddress
		options = append([]L1KeySelectOption{L1ExistingKey}, options...)
	}

	return &L1KeySelect{
		Selector: ui.Selector[L1KeySelectOption]{
			Options:    options,
			CannotBack: true,
		},
		BaseModel: weavecontext.BaseModel{Ctx: weavecontext.SetCurrentState(ctx, state), CannotBack: true},
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
	l2ChainId := MustGetL2ChainId(ctx)
	options := []L2KeySelectOption{
		L2SameKey,
		L2GenerateKey,
		L2ImportKey,
	}
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	if l2RelayerAddress, found := cosmosutils.GetHermesRelayerAddress(state.hermesBinaryPath, l2ChainId); found {
		state.l2RelayerAddress = l2RelayerAddress
		options = append([]L2KeySelectOption{L2ExistingKey}, options...)
	}

	return &L2KeySelect{
		Selector: ui.Selector[L2KeySelectOption]{
			Options: options,
		},
		BaseModel: weavecontext.BaseModel{Ctx: weavecontext.SetCurrentState(ctx, state)},
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

		switch L1KeySelectOption(state.l1KeyMethod) {
		case L1ExistingKey:
			switch *selected {
			case L2ExistingKey:
				model := NewFetchingBalancesLoading(weavecontext.SetCurrentState(m.Ctx, state))
				return model, model.Init()
			case L2SameKey:
				state.l2RelayerAddress = state.l1RelayerAddress
				model := NewFetchingBalancesLoading(weavecontext.SetCurrentState(m.Ctx, state))
				return model, model.Init()
			case L2GenerateKey:
				model := NewGenerateL2RelayerKeyLoading(weavecontext.SetCurrentState(m.Ctx, state))
				return model, model.Init()
			case L2ImportKey:
				return NewImportL2RelayerKeyInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
			}
		case L1GenerateKey:
			model := NewGenerateL1RelayerKeyLoading(weavecontext.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		case L1ImportKey:
			return NewImportL1RelayerKeyInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
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

type GenerateL1RelayerKeyLoading struct {
	loading ui.Loading
	weavecontext.BaseModel
}

func NewGenerateL1RelayerKeyLoading(ctx context.Context) *GenerateL1RelayerKeyLoading {
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	layerText := "L1"
	if state.l1KeyMethod == string(L1GenerateKey) && state.l2KeyMethod == string(L2SameKey) {
		layerText = "L1 and L2"
	}

	return &GenerateL1RelayerKeyLoading{
		loading:   ui.NewLoading(fmt.Sprintf("Generating new relayer account key for %s ...", layerText), waitGenerateL1RelayerKeyLoading(ctx)),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *GenerateL1RelayerKeyLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func waitGenerateL1RelayerKeyLoading(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1500 * time.Millisecond)

		state := weavecontext.GetCurrentState[RelayerState](ctx)
		l1ChainId := MustGetL1ChainId(ctx)

		relayerKey, err := cosmosutils.GenerateAndReplaceHermesKey(state.hermesBinaryPath, l1ChainId)
		if err != nil {
			panic(err)
		}
		state.l1RelayerAddress = relayerKey.Address
		state.l1RelayerMnemonic = relayerKey.Mnemonic

		return ui.EndLoading{
			Ctx: weavecontext.SetCurrentState(ctx, state),
		}
	}
}

func (m *GenerateL1RelayerKeyLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = m.loading.EndContext
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		switch L2KeySelectOption(state.l2KeyMethod) {
		case L2ExistingKey, L2ImportKey:
			return NewKeysMnemonicDisplayInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
		case L2SameKey:
			state.l2RelayerAddress = state.l1RelayerAddress
			state.l2RelayerMnemonic = state.l1RelayerMnemonic
			return NewKeysMnemonicDisplayInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
		case L2GenerateKey:
			model := NewGenerateL2RelayerKeyLoading(weavecontext.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		}
	}
	return m, cmd
}

func (m *GenerateL1RelayerKeyLoading) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type GenerateL2RelayerKeyLoading struct {
	loading ui.Loading
	weavecontext.BaseModel
}

func NewGenerateL2RelayerKeyLoading(ctx context.Context) *GenerateL2RelayerKeyLoading {
	return &GenerateL2RelayerKeyLoading{
		loading:   ui.NewLoading("Generating new relayer account key for L2 ...", waitGenerateL2RelayerKeyLoading(ctx)),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *GenerateL2RelayerKeyLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func waitGenerateL2RelayerKeyLoading(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1500 * time.Millisecond)

		state := weavecontext.GetCurrentState[RelayerState](ctx)
		l2ChainId := MustGetL2ChainId(ctx)

		relayerKey, err := cosmosutils.GenerateAndReplaceHermesKey(state.hermesBinaryPath, l2ChainId)
		if err != nil {
			panic(err)
		}
		state.l2RelayerAddress = relayerKey.Address
		state.l2RelayerMnemonic = relayerKey.Mnemonic

		return ui.EndLoading{
			Ctx: weavecontext.SetCurrentState(ctx, state),
		}
	}
}

func (m *GenerateL2RelayerKeyLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = m.loading.EndContext
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		return NewKeysMnemonicDisplayInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
	}
	return m, cmd
}

func (m *GenerateL2RelayerKeyLoading) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type KeysMnemonicDisplayInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question string
}

func NewKeysMnemonicDisplayInput(ctx context.Context) *KeysMnemonicDisplayInput {
	model := &KeysMnemonicDisplayInput{
		TextInput: ui.NewTextInput(true),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please type `continue` to proceed after you have securely stored the mnemonic.",
	}
	model.WithPlaceholder("Type `continue` to continue, Ctrl+C to quit.")
	model.WithValidatorFn(common.ValidateExactString("continue"))
	return model
}

func (m *KeysMnemonicDisplayInput) GetQuestion() string {
	return m.question
}

func (m *KeysMnemonicDisplayInput) Init() tea.Cmd {
	return nil
}

func (m *KeysMnemonicDisplayInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		extraText := " has"
		if state.l1KeyMethod == string(L1GenerateKey) && state.l2KeyMethod == string(L2GenerateKey) {
			extraText = "s have"
		}
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, fmt.Sprintf("Relayer key%s been successfully generated.", extraText), []string{}, ""))

		switch L2KeySelectOption(state.l2KeyMethod) {
		case L2ExistingKey, L2GenerateKey, L2SameKey:
			model := NewFetchingBalancesLoading(weavecontext.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		case L2ImportKey:
			return NewImportL2RelayerKeyInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *KeysMnemonicDisplayInput) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	var mnemonicText string

	if state.l1KeyMethod == string(L1GenerateKey) {
		layerText := "L1"
		if state.l2KeyMethod == string(L2SameKey) {
			layerText = "L1 and L2"
		}
		mnemonicText += styles.RenderMnemonic(
			styles.RenderPrompt(fmt.Sprintf("Weave Relayer on %s", layerText), []string{layerText}, styles.Empty),
			state.l1RelayerAddress,
			state.l1RelayerMnemonic,
		)
	}

	if state.l2KeyMethod == string(L2GenerateKey) {
		mnemonicText += styles.RenderMnemonic(
			styles.RenderPrompt(fmt.Sprintf("Weave Relayer on L2"), []string{"L2"}, styles.Empty),
			state.l2RelayerAddress,
			state.l2RelayerMnemonic,
		)
	}

	return state.weave.Render() + "\n" +
		styles.BoldUnderlineText("Important", styles.Yellow) + "\n" +
		styles.Text("Write down these mnemonic phrases and store them in a safe place. \nIt is the only way to recover your system keys.", styles.Yellow) + "\n\n" +
		mnemonicText + styles.RenderPrompt(m.GetQuestion(), []string{"`continue`"}, styles.Question) + m.TextInput.View()
}

type ImportL1RelayerKeyInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question  string
	layerText string
}

func NewImportL1RelayerKeyInput(ctx context.Context) *ImportL1RelayerKeyInput {
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	layerText := "L1"
	if state.l1KeyMethod == string(L1ImportKey) && state.l2KeyMethod == string(L2SameKey) {
		layerText = "L1 and L2"
	}
	model := &ImportL1RelayerKeyInput{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  fmt.Sprintf("Please add mnemonic for relayer account key on %s", layerText),
		layerText: layerText,
	}
	model.WithPlaceholder("Enter the mnemonic")
	model.WithValidatorFn(common.ValidateMnemonic)
	return model
}

func (m *ImportL1RelayerKeyInput) GetQuestion() string {
	return m.question
}

func (m *ImportL1RelayerKeyInput) Init() tea.Cmd {
	return nil
}

func (m *ImportL1RelayerKeyInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		relayerKey, err := cosmosutils.RecoverAndReplaceHermesKey(state.hermesBinaryPath, MustGetL1ChainId(m.Ctx), input.Text)
		if err != nil {
			panic(err)
		}

		state.l1RelayerMnemonic = relayerKey.Mnemonic
		state.l1RelayerAddress = relayerKey.Address
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"relayer account key", m.layerText}, styles.HiddenMnemonicText))

		switch L2KeySelectOption(state.l2KeyMethod) {
		case L2ExistingKey:
			model := NewFetchingBalancesLoading(weavecontext.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		case L2SameKey:
			state.l2RelayerAddress = relayerKey.Address
			state.l2RelayerMnemonic = relayerKey.Mnemonic
			model := NewFetchingBalancesLoading(weavecontext.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		case L2GenerateKey:
			model := NewGenerateL2RelayerKeyLoading(weavecontext.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		case L2ImportKey:
			return NewImportL2RelayerKeyInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
	}
	m.TextInput = input
	return m, cmd
}

func (m *ImportL1RelayerKeyInput) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"relayer account key", m.layerText}, styles.Question) + m.TextInput.View()
}

type ImportL2RelayerKeyInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question string
}

func NewImportL2RelayerKeyInput(ctx context.Context) *ImportL2RelayerKeyInput {
	model := &ImportL2RelayerKeyInput{
		TextInput: ui.NewTextInput(false),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
		question:  "Please add mnemonic for relayer account key on L2",
	}
	model.WithPlaceholder("Enter the mnemonic")
	model.WithValidatorFn(common.ValidateMnemonic)
	return model
}

func (m *ImportL2RelayerKeyInput) GetQuestion() string {
	return m.question
}

func (m *ImportL2RelayerKeyInput) Init() tea.Cmd {
	return nil
}

func (m *ImportL2RelayerKeyInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		relayerKey, err := cosmosutils.RecoverAndReplaceHermesKey(state.hermesBinaryPath, MustGetL2ChainId(m.Ctx), input.Text)
		if err != nil {
			panic(err)
		}

		state.l2RelayerMnemonic = relayerKey.Mnemonic
		state.l2RelayerAddress = relayerKey.Address
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"relayer account key", "L2"}, styles.HiddenMnemonicText))

		model := NewFetchingBalancesLoading(weavecontext.SetCurrentState(m.Ctx, state))
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *ImportL2RelayerKeyInput) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"relayer account key", "L2"}, styles.Question) + m.TextInput.View()
}

type FetchingBalancesLoading struct {
	loading ui.Loading
	weavecontext.BaseModel
}

func NewFetchingBalancesLoading(ctx context.Context) *FetchingBalancesLoading {
	return &FetchingBalancesLoading{
		loading:   ui.NewLoading("Fetching relayer account balances ...", waitFetchingBalancesLoading(ctx)),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *FetchingBalancesLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func waitFetchingBalancesLoading(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := weavecontext.GetCurrentState[RelayerState](ctx)

		l1Rest := MustGetL1ActiveLcd(ctx)
		l1Balances, err := cosmosutils.QueryBankBalances(l1Rest, state.l1RelayerAddress)
		if err != nil {
			panic(fmt.Errorf("cannot fetch balance for l1: %v", err))
		}
		if l1Balances.IsZero() {
			state.l1NeedsFunding = true
		}

		l2ChainId := MustGetL2ChainId(ctx)
		l2Registry := registry.MustGetL2Registry(registry.InitiaL1Testnet, l2ChainId)
		l2Rest := l2Registry.MustGetActiveLcd()
		l2Balances, err := cosmosutils.QueryBankBalances(l2Rest, state.l2RelayerAddress)
		if err != nil {
			panic(fmt.Errorf("cannot fetch balance for l2: %v", err))
		}
		if l2Balances.IsZero() {
			state.l2NeedsFunding = true
		}

		return ui.EndLoading{
			Ctx: weavecontext.SetCurrentState(ctx, state),
		}
	}
}

func (m *FetchingBalancesLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = m.loading.EndContext
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		if !state.l1NeedsFunding && !state.l2NeedsFunding {
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "Your relayer has been setup successfully. 🎉", []string{}, ""))
			return NewTerminalState(weavecontext.SetCurrentState(m.Ctx, state)), tea.Quit
		}

		return NewFundingAmountSelect(weavecontext.SetCurrentState(m.Ctx, state)), nil
	}
	return m, cmd
}

func (m *FetchingBalancesLoading) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

type FundingAmountSelect struct {
	ui.Selector[FundingAmountSelectOption]
	weavecontext.BaseModel
	question string
}

type FundingAmountSelectOption string

const (
	FundingFillManually FundingAmountSelectOption = "○ Fill in an amount manually to fund from Gas Station Account"
	FundingUserTransfer FundingAmountSelectOption = "○ Transfer funds manually from other account"
)

var (
	FundingDefaultPreset FundingAmountSelectOption = ""
)

func NewFundingAmountSelect(ctx context.Context) *FundingAmountSelect {
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	FundingDefaultPreset = FundingAmountSelectOption(fmt.Sprintf(
		"○ Use the default preset\n    Total amount that will be transferred from Gas Station account:\n    %s %s on L1 %s\n    %s %s on L2 %s",
		styles.BoldText(fmt.Sprintf("• L1 (%s):", MustGetL1ChainId(ctx)), styles.Cyan),
		styles.BoldText(fmt.Sprintf("%s%s", DefaultL1RelayerBalance, MustGetL1GasDenom(ctx)), styles.White),
		styles.Text(fmt.Sprintf("(%s)", state.l1RelayerAddress), styles.Gray),
		styles.BoldText(fmt.Sprintf("• L2 (%s):", MustGetL2ChainId(ctx)), styles.Cyan),
		styles.BoldText(fmt.Sprintf("%s%s", DefaultL2RelayerBalance, MustGetL2GasDenom(ctx)), styles.White),
		styles.Text(fmt.Sprintf("(%s)", state.l2RelayerAddress), styles.Gray),
	))
	return &FundingAmountSelect{
		Selector: ui.Selector[FundingAmountSelectOption]{
			Options: []FundingAmountSelectOption{
				FundingDefaultPreset,
				FundingFillManually,
				FundingUserTransfer,
			},
			CannotBack: true,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  fmt.Sprintf("Please select the filling amount option"),
	}
}

func (m *FundingAmountSelect) GetQuestion() string {
	return m.question
}

func (m *FundingAmountSelect) Init() tea.Cmd {
	return nil
}

func (m *FundingAmountSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[RelayerState](m)

		switch *selected {
		case FundingDefaultPreset:
			// TODO: Continue
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, "Use the default preset"))
		case FundingFillManually:
			// TODO: Continue
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, "Fill in an amount manually to fund from Gas Station Account"))
		case FundingUserTransfer:
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{}, "Transfer funds manually from other account"))
			state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "Your relayer has been set up successfully! 🎉", []string{}, ""))
			state.weave.PushPreviousResponse(fmt.Sprintf(
				"%s %s\n  %s\n%s\n\n",
				styles.Text("i", styles.Yellow),
				styles.BoldUnderlineText("Important", styles.Yellow),
				styles.Text("However, to ensure the relayer functions properly, please make sure these accounts are funded.", styles.Yellow),
				styles.CreateFrame(fmt.Sprintf(
					"%s %s\n%s %s",
					styles.BoldText("• Relayer key on L1", styles.White),
					styles.Text(fmt.Sprintf("(%s)", state.l1RelayerAddress), styles.Gray),
					styles.BoldText("• Relayer key on L2", styles.White),
					styles.Text(fmt.Sprintf("(%s)", state.l2RelayerAddress), styles.Gray),
				), 65),
			))
			return NewTerminalState(weavecontext.SetCurrentState(m.Ctx, state)), tea.Quit
		}
	}

	return m, cmd
}

func (m *FundingAmountSelect) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)

	var informationLayer, warningLayer string
	if state.l1NeedsFunding && state.l2NeedsFunding {
		informationLayer = "both L1 and L2"
		warningLayer = "L1 and L2 have"
	} else if state.l1NeedsFunding {
		informationLayer = "L1"
		warningLayer = "L1 has"
	} else if state.l2NeedsFunding {
		informationLayer = "L2"
		warningLayer = "L2 has"
	}

	return state.weave.Render() + "\n" +
		styles.RenderPrompt(
			fmt.Sprintf("You will need to fund the relayer account on %s.\n  You can either transfer funds from created Gas Station Account or transfer manually.", informationLayer),
			[]string{informationLayer},
			styles.Information,
		) + "\n\n" +
		styles.BoldUnderlineText("Important", styles.Yellow) + "\n" +
		styles.Text(fmt.Sprintf("The relayer account on %s have no funds.\nYou will need to fund the account in order to run the relayer properly.", warningLayer), styles.Yellow) + "\n\n" +
		styles.RenderPrompt(
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
		question:  "Select the Initia L1 network you want to connect your rollup to",
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

func NewSelectingL2Network(ctx context.Context, chainType registry.ChainType) *SelectingL2Network {
	networks := registry.MustGetAllL2AvailableNetwork(chainType)

	options := make([]string, 0)
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

		re := regexp.MustCompile(`\(([^)]+)\)`)
		chainId := re.FindStringSubmatch(m.Options[m.Cursor])[1]
		l2Registry, err := registry.GetL2Registry(registry.InitiaL1Testnet, chainId)
		if err != nil {
			panic(err)
		}
		lcdAddress, err := l2Registry.GetActiveLcd()
		if err != nil {
			panic(err)
		}
		httpClient := client.NewHTTPClient()
		var res types.ChannelsResponse
		_, err = httpClient.Get(lcdAddress, "/ibc/core/channel/v1/channels", nil, &res)
		if err != nil {
			panic(err)
		}

		pairs := make([]types.IBCChannelPair, 0)
		for _, channel := range res.Channels {
			pairs = append(pairs, types.IBCChannelPair{
				L1: channel.Counterparty,
				L2: types.Channel{
					PortID:    channel.PortID,
					ChannelID: channel.ChannelID,
				},
			})
		}

		l2DefaultFeeToken := l2Registry.MustGetDefaultFeeToken()
		l2Rpc := l2Registry.MustGetActiveRpc()
		state.Config["l2.chain_id"] = chainId
		state.Config["l2.gas_price.denom"] = l2DefaultFeeToken.Denom
		state.Config["l2.gas_price.price"] = strconv.FormatFloat(l2DefaultFeeToken.FixedMinGasPrice, 'f', -1, 64)
		state.Config["l2.rpc_address"] = l2Rpc
		state.Config["l2.grpc_address"] = l2Registry.MustGetActiveGrpc()
		state.Config["l2.websocket"] = l2Registry.MustGetActiveWebsocket()

		return NewIBCChannelsCheckbox(weavecontext.SetCurrentState(m.Ctx, state), pairs), nil
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
		question:  "Select the Initia L1 network you want to connect your rollup to",
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
			state.Config["l1.websocket"] = testnetRegistry.MustGetActiveWebsocket()
			state.Config["l1.gas_price.denom"] = DefaultGasPriceDenom
			state.Config["l1.gas_price.price"] = testnetRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)

			return NewSelectingL2Network(weavecontext.SetCurrentState(m.Ctx, state), registry.InitiaL1Testnet), nil
		case Mainnet:
			mainnetRegistry := registry.MustGetChainRegistry(registry.InitiaL1Mainnet)
			state.Config["l1.chain_id"] = mainnetRegistry.GetChainId()
			state.Config["l1.rpc_address"] = mainnetRegistry.MustGetActiveRpc()
			state.Config["l1.grpc_address"] = mainnetRegistry.MustGetActiveGrpc()
			state.Config["l1.lcd_address"] = mainnetRegistry.MustGetActiveLcd()
			state.Config["l1.websocket"] = mainnetRegistry.MustGetActiveWebsocket()
			state.Config["l1.gas_price.denom"] = DefaultGasPriceDenom
			state.Config["l1.gas_price.price"] = mainnetRegistry.MustGetMinGasPriceByDenom(DefaultGasPriceDenom)

			return NewSelectingL2Network(weavecontext.SetCurrentState(m.Ctx, state), registry.InitiaL1Testnet), nil
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
			var metadata types.Metadata
			var networkRegistry *registry.ChainRegistry
			if state.Config["l1.chain_id"] == InitiaTestnetChainId {
				networkRegistry = registry.MustGetChainRegistry(registry.InitiaL1Testnet)
				info := networkRegistry.MustGetOpinitBridgeInfo(artifacts.BridgeID)
				metadata = types.MustDecodeBridgeMetadata(info.BridgeConfig.Metadata)
			} else {
				panic(fmt.Sprintf("not support for l1 %s", state.Config["l1.chain_id"]))
			}
			channelPairs := make([]types.IBCChannelPair, 0)
			for _, channel := range metadata.PermChannels {
				counterparty := networkRegistry.MustGetCounterPartyIBCChannel(channel.PortID, channel.ChannelID)
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

func MustGetL1ChainId(ctx context.Context) string {
	chainId, ok := GetL1ChainId(ctx)
	if !ok {
		panic("cannot get l1 chain id")
	}

	return chainId
}

func GetL2ChainId(ctx context.Context) (string, bool) {
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	if chainId, found := state.Config["l2.chain_id"]; found {
		return chainId, found
	}
	return "", false
}

func MustGetL2ChainId(ctx context.Context) string {
	chainId, ok := GetL2ChainId(ctx)
	if !ok {
		panic("cannot get l2 chain id")
	}

	return chainId
}

func GetL1ActiveLcd(ctx context.Context) (string, bool) {
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	if lcd, found := state.Config["l1.lcd_address"]; found {
		return lcd, found
	}
	return "", false
}

func MustGetL1ActiveLcd(ctx context.Context) string {
	lcd, ok := GetL1ActiveLcd(ctx)
	if !ok {
		panic("cannot get l1 active lcd from state")
	}

	return lcd
}

func GetL1GasDenom(ctx context.Context) (string, bool) {
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	if denom, found := state.Config["l1.gas_price.denom"]; found {
		return denom, found
	}

	return "", false
}

func MustGetL1GasDenom(ctx context.Context) string {
	denom, ok := GetL1GasDenom(ctx)
	if !ok {
		panic("cannot get l1 gas denom from state")
	}

	return denom
}

func GetL2GasDenom(ctx context.Context) (string, bool) {
	state := weavecontext.GetCurrentState[RelayerState](ctx)
	if denom, found := state.Config["l2.gas_price.denom"]; found {
		return denom, found
	}

	return "", false
}

func MustGetL2GasDenom(ctx context.Context) string {
	denom, ok := GetL2GasDenom(ctx)
	if !ok {
		panic("cannot get l2 gas denom from state")
	}

	return denom
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
		state.IBCChannels = append(state.IBCChannels, types.IBCChannelPair{})
		state.IBCChannels[m.idx].L1.PortID = m.TextInput.Text
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
		state.IBCChannels[m.idx].L1.ChannelID = m.TextInput.Text
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
		state.IBCChannels[m.idx].L2.PortID = m.TextInput.Text
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
		state.IBCChannels[m.idx].L2.ChannelID = m.TextInput.Text
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
		model := NewSettingUpRelayer(weavecontext.SetCurrentState(m.Ctx, state))
		return model, model.Init()
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
	err      error
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

		// TODO: should have loding state for this
		httpClient := client.NewHTTPClient()
		var res types.ChannelsResponse
		_, err := httpClient.Get(input.Text, "/ibc/core/channel/v1/channels", nil, &res)
		if err != nil {
			m.err = fmt.Errorf("Unable to call the LCD endpoint '%s'. Please verify that the address is correct and reachablem", input.Text)
			return m, cmd
		}

		pairs := make([]types.IBCChannelPair, 0)
		for _, channel := range res.Channels {
			pairs = append(pairs, types.IBCChannelPair{
				L1: channel.Counterparty,
				L2: types.Channel{
					PortID:    channel.PortID,
					ChannelID: channel.ChannelID,
				},
			})
		}

		return NewIBCChannelsCheckbox(weavecontext.SetCurrentState(m.Ctx, state), pairs), nil
	}
	m.TextInput = input
	return m, cmd
}

func (m *FillL2LCD) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	if m.err != nil {
		return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2", "LCD_address", m.extra}, styles.Question) + m.TextInput.ViewErr(m.err)
	}
	return state.weave.Render() + styles.RenderPrompt(m.GetQuestion(), []string{"L2", "LCD_address", m.extra}, styles.Question) + m.TextInput.View()

}

type SettingUpRelayer struct {
	loading ui.Loading
	weavecontext.BaseModel
}

func NewSettingUpRelayer(ctx context.Context) *SettingUpRelayer {
	return &SettingUpRelayer{
		loading:   ui.NewLoading("Setting up relayer...", WaitSettingUpRelayer(ctx)),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

// Utility function for determining Hermes binary URL, panicking on errors
func getHermesBinaryURL(version string) string {
	baseURL := "https://github.com/informalsystems/hermes/releases/download"
	arch := runtime.GOARCH
	os := runtime.GOOS

	var binaryType string
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		panic(fmt.Sprintf("Unsupported architecture: %s", arch))
	}

	switch os {
	case "darwin":
		binaryType = "apple-darwin"
	case "linux":
		binaryType = "unknown-linux-gnu"
	default:
		panic(fmt.Sprintf("Unsupported operating system: %s", os))
	}

	fileName := fmt.Sprintf("hermes-%s-%s-%s.tar.gz", version, arch, binaryType)
	return fmt.Sprintf("%s/%s/%s", baseURL, version, fileName)
}

func WaitSettingUpRelayer(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := weavecontext.GetCurrentState[RelayerState](ctx)

		// Get Hermes binary URL based on OS and architecture
		hermesURL := getHermesBinaryURL(HermesVersion)

		// Get the user's home directory
		userHome, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("Failed to get user home directory: %v", err))
		}

		// Define paths
		weaveDataPath := filepath.Join(userHome, common.WeaveDataDirectory)
		tarballPath := filepath.Join(weaveDataPath, fmt.Sprintf("hermes-%s.tar.gz", HermesVersion))
		state.hermesBinaryPath = filepath.Join(weaveDataPath, "hermes")

		// Ensure the data directory exists
		if err := os.MkdirAll(weaveDataPath, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create data directory: %v", err))
		}

		// Download and extract Hermes tarball
		if err := weaveio.DownloadAndExtractTarGz(hermesURL, tarballPath, weaveDataPath); err != nil {
			panic(fmt.Sprintf("Failed to download and extract Hermes: %v", err))
		}

		// Make the Hermes binary executable
		if err := os.Chmod(state.hermesBinaryPath, 0755); err != nil {
			panic(fmt.Sprintf("Failed to set executable permissions for Hermes: %v", err))
		}

		// Remove quarantine attribute on macOS
		if runtime.GOOS == "darwin" {
			if err := removeQuarantineAttribute(state.hermesBinaryPath); err != nil {
				panic(fmt.Sprintf("Failed to remove quarantine attribute on macOS: %v", err))
			}
		}

		// Create Hermes configuration
		createHermesConfig(state)

		srv, err := service.NewService(service.Relayer)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize service: %v", err))
		}

		if err = srv.Create("", filepath.Join(userHome, HermesHome)); err != nil {
			panic(fmt.Sprintf("failed to create service: %v", err))
		}

		// Return updated state
		return ui.EndLoading{Ctx: weavecontext.SetCurrentState(ctx, state)}
	}
}

func (m *SettingUpRelayer) Init() tea.Cmd {
	return m.loading.Init()
}

func (m *SettingUpRelayer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[RelayerState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = m.loading.EndContext
		state := weavecontext.PushPageAndGetState[RelayerState](m)
		return NewL1KeySelect(weavecontext.SetCurrentState(m.Ctx, state)), nil
	}
	return m, cmd
}

func (m *SettingUpRelayer) View() string {
	state := weavecontext.GetCurrentState[RelayerState](m.Ctx)
	return state.weave.Render() + "\n" + m.loading.View()
}

func removeQuarantineAttribute(filePath string) error {
	cmd := exec.Command("xattr", "-l", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If listing attributes fails, assume no quarantine attribute
		return nil
	}

	// Check if com.apple.quarantine exists
	if !containsQuarantine(string(output)) {
		return nil
	}

	// Remove the quarantine attribute
	cmd = exec.Command("xattr", "-d", "com.apple.quarantine", filePath)
	return cmd.Run()
}

func containsQuarantine(attrs string) bool {
	return stringContains(attrs, "com.apple.quarantine")
}

func stringContains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && haystack[len(haystack)-len(needle):] == needle
}
