package models

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/common"
	"github.com/initia-labs/weave/config"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/crypto"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/ui"
)

type ExistingCheckerState struct {
	weave             types.WeaveState
	generatedMnemonic string
}

func NewExistingCheckerState() ExistingCheckerState {
	return ExistingCheckerState{
		weave: types.NewWeaveState(),
	}
}

func (e ExistingCheckerState) Clone() ExistingCheckerState {
	return ExistingCheckerState{
		weave:             e.weave.Clone(),
		generatedMnemonic: e.generatedMnemonic,
	}

}

func InitHeader() string {
	return styles.FadeText("Welcome to Weave! 🪢\n\n") +
		styles.RenderPrompt("As this is your first time using Weave, we recommend setting up a Gas Station account.\nThis account will hold funds to be distributed to OPinit Bots, IBC Relayers, and other services for gas fees.\n\n", []string{"Gas Station account"}, styles.Empty) +
		styles.BoldText("Please note that Weave will not send any transactions without your confirmation.\n", styles.Yellow) +
		styles.Text("While you can complete this setup later, we recommend doing it now to ensure a smoother experience.\n\n", styles.Gray)
}

type ExistingWeaveChecker struct {
	weavecontext.BaseModel
	ui.Selector[ExistingWeaveOption]
	skipToModel tea.Model
}

type ExistingWeaveOption string

const (
	Yes ExistingWeaveOption = "Yes"
	No  ExistingWeaveOption = "No"
)

func NewExistingAppChecker(ctx context.Context, skipToModel tea.Model) *ExistingWeaveChecker {
	return &ExistingWeaveChecker{
		Selector: ui.Selector[ExistingWeaveOption]{
			Options: []ExistingWeaveOption{
				Yes,
				No,
			},
			CannotBack: true,
		},
		BaseModel:   weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		skipToModel: skipToModel,
	}
}

func (m *ExistingWeaveChecker) Init() tea.Cmd {
	return ui.DoTick()
}

func (m *ExistingWeaveChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[ExistingCheckerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[ExistingCheckerState](m)

		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, "Would you like to set up a Gas Station account", []string{"Gas Station account"}, string(*selected)),
		)
		m.Ctx = weavecontext.SetCurrentState(m.Ctx, state)
		switch *selected {
		case Yes:
			return NewGasStationMethodSelect(m.Ctx), nil
		case No:
			return m.skipToModel, m.skipToModel.Init()
		}
	}
	return m, cmd
}

func (m *ExistingWeaveChecker) View() string {
	view := InitHeader()
	view += styles.RenderPrompt("Would you like to set up a Gas Station account", []string{"Gas Station account"}, styles.Question)
	view += m.Selector.View()

	return view
}

type GasStationMethodSelect struct {
	weavecontext.BaseModel
	ui.Selector[GasStationMethodOption]
}

type GasStationMethodOption string

const (
	Generate GasStationMethodOption = "Generate new account (recommended)"
	Import   GasStationMethodOption = "Import existing account using mnemonic"
)

func NewGasStationMethodSelect(ctx context.Context) *GasStationMethodSelect {
	tooltip := ui.NewTooltipSlice(ui.NewTooltip(
		"Gas Station",
		"Gas Station account is the account from which Weave will use to fund necessary system accounts, enabling easier and more seamless experience while setting up things using Weave.",
		"** Weave will NOT automatically send transactions without asking for your confirmation. **",
		[]string{}, []string{}, []string{}), 2)
	return &GasStationMethodSelect{
		Selector: ui.Selector[GasStationMethodOption]{
			Options: []GasStationMethodOption{
				Generate,
				Import,
			},
			Tooltips:   &tooltip,
			CannotBack: true,
		},
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *GasStationMethodSelect) Init() tea.Cmd {
	return nil
}

func (m *GasStationMethodSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[ExistingCheckerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := weavecontext.PushPageAndGetState[ExistingCheckerState](m)

		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, "How would you like to setup your Gas station account?", []string{"Gas Station account"}, string(*selected)),
		)
		switch *selected {
		case Generate:
			model := NewGenerateOrRecoverSystemKeysLoading(weavecontext.SetCurrentState(m.Ctx, state))
			return model, model.Init()
		case Import:
			return NewGasStationMnemonicInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
		}
	}
	return m, cmd
}

func (m *GasStationMethodSelect) View() string {
	state := weavecontext.GetCurrentState[ExistingCheckerState](m.Ctx)
	m.Selector.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return InitHeader() + state.weave.Render() +
		styles.RenderPrompt("How would you like to setup your Gas station account?", []string{"Gas Station account"}, styles.Question) +
		m.Selector.View()
}

type GenerateGasStationLoading struct {
	loading ui.Loading
	weavecontext.BaseModel
}

func NewGenerateOrRecoverSystemKeysLoading(ctx context.Context) *GenerateGasStationLoading {
	return &GenerateGasStationLoading{
		loading:   ui.NewLoading("Generating new Gas Station account...", generateGasStationAccount(ctx)),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (m *GenerateGasStationLoading) Init() tea.Cmd {
	return m.loading.Init()
}

func generateGasStationAccount(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state := weavecontext.GetCurrentState[ExistingCheckerState](ctx)

		mnemonic, err := crypto.GenerateMnemonic()
		if err != nil {
			panic(fmt.Errorf("failed to generate gas station mnemonic: %w", err))
		}
		state.generatedMnemonic = mnemonic

		time.Sleep(1500 * time.Millisecond)

		return ui.EndLoading{
			Ctx: weavecontext.SetCurrentState(ctx, state),
		}
	}
}

func (m *GenerateGasStationLoading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[ExistingCheckerState](m, msg); handled {
		return model, cmd
	}

	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		m.Ctx = m.loading.EndContext
		state := weavecontext.PushPageAndGetState[ExistingCheckerState](m)
		state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.NoSeparator, "Gas Station account has been successfully generated.", []string{}, ""))
		return NewSystemKeysMnemonicDisplayInput(weavecontext.SetCurrentState(m.Ctx, state)), nil
	}
	return m, cmd
}

func (m *GenerateGasStationLoading) View() string {
	state := weavecontext.GetCurrentState[ExistingCheckerState](m.Ctx)
	return InitHeader() + state.weave.Render() + "\n" + m.loading.View()
}

type GasStationMnemonicDisplayInput struct {
	ui.TextInput
	weavecontext.BaseModel
	question string
}

func NewSystemKeysMnemonicDisplayInput(ctx context.Context) *GasStationMnemonicDisplayInput {
	model := &GasStationMnemonicDisplayInput{
		TextInput: ui.NewTextInput(true),
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		question:  "Please type `continue` to proceed after you have securely stored the mnemonic.",
	}
	model.WithPlaceholder("Type `continue` to continue, Ctrl+C to quit.")
	model.WithValidatorFn(common.ValidateExactString("continue"))
	return model
}

func (m *GasStationMnemonicDisplayInput) GetQuestion() string {
	return m.question
}

func (m *GasStationMnemonicDisplayInput) Init() tea.Cmd {
	return nil
}

func (m *GasStationMnemonicDisplayInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[ExistingCheckerState](m, msg); handled {
		return model, cmd
	}

	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[ExistingCheckerState](m)
		model := NewWeaveAppInitialization(m.Ctx, state.generatedMnemonic)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *GasStationMnemonicDisplayInput) View() string {
	state := weavecontext.GetCurrentState[ExistingCheckerState](m.Ctx)
	gasStationAddress, err := crypto.MnemonicToBech32Address("init", state.generatedMnemonic)
	if err != nil {
		panic(fmt.Errorf("failed to convert mnemonic to bech32 address: %w", err))
	}

	var mnemonicText string
	mnemonicText += styles.RenderMnemonic("Gas Station", gasStationAddress, state.generatedMnemonic)

	return InitHeader() + state.weave.Render() + "\n" +
		styles.BoldUnderlineText("Important", styles.Yellow) + "\n" +
		styles.Text("Write down these mnemonic phrases and store them in a safe place. \nIt is the only way to recover your system keys.", styles.Yellow) + "\n\n" +
		mnemonicText + styles.RenderPrompt(m.GetQuestion(), []string{"`continue`"}, styles.Question) + m.TextInput.View()
}

type GasStationMnemonicInput struct {
	weavecontext.BaseModel
	firstTime bool
	ui.TextInput
}

func NewGasStationMnemonicInput(ctx context.Context) *GasStationMnemonicInput {
	tooltip := ui.NewTooltip(
		"Gas station account",
		"Gas Station account is the account from which Weave will use to fund necessary system accounts, enabling easier and more seamless experience while setting up things using Weave.",
		"** Weave will NOT automatically send transactions without asking for your confirmation. **",
		[]string{}, []string{}, []string{})
	model := &GasStationMnemonicInput{
		firstTime: config.IsFirstTimeSetup(),
		TextInput: ui.NewTextInput(true),
		BaseModel: weavecontext.BaseModel{Ctx: ctx},
	}
	model.WithPlaceholder("Add mnemonic")
	model.WithValidatorFn(common.ValidateMnemonic)
	model.WithTooltip(&tooltip)
	return model
}

func (m *GasStationMnemonicInput) Init() tea.Cmd {
	return nil
}

func (m *GasStationMnemonicInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := weavecontext.HandleCommonCommands[ExistingCheckerState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := weavecontext.PushPageAndGetState[ExistingCheckerState](m)

		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.DotsSeparator, "Please set up a Gas Station account", []string{"Gas Station account"}, styles.HiddenMnemonicText),
		)
		model := NewWeaveAppInitialization(weavecontext.SetCurrentState(m.Ctx, state), input.Text)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *GasStationMnemonicInput) View() string {
	var header string
	if m.firstTime {
		header = InitHeader()
	} else {
		header = styles.FadeText("Welcome to Weave! 🪢\n\n") +
			styles.RenderPrompt("Since you've previously set up your Gas Station account, you have the option to override it with a new one.\nThis account will continue to hold the necessary funds for the OPinit-bots or relayer to send transactions.\n\n", []string{"Gas Station account"}, styles.Empty) +
			styles.BoldText("Please remember, Weave will only send transactions after your confirmation.\n", styles.Yellow)
	}
	state := weavecontext.GetCurrentState[ExistingCheckerState](m.Ctx)
	m.TextInput.ToggleTooltip = weavecontext.GetTooltip(m.Ctx)
	return header + state.weave.Render() + styles.RenderPrompt("Please set up a Gas Station account", []string{"Gas Station account"}, styles.Question) + m.TextInput.View()
}

type WeaveAppInitialization struct {
	weavecontext.BaseModel
	loading  ui.Loading
	mnemonic string
}

func NewWeaveAppInitialization(ctx context.Context, mnemonic string) tea.Model {
	return &WeaveAppInitialization{
		loading:  ui.NewLoading("Initializing Weave...", WaitSetGasStation(mnemonic)),
		mnemonic: mnemonic,
		BaseModel: weavecontext.BaseModel{
			Ctx:        ctx,
			CannotBack: true,
		},
	}
}

func (hi *WeaveAppInitialization) Init() tea.Cmd {
	return hi.loading.Init()
}

func WaitSetGasStation(mnemonic string) tea.Cmd {
	return func() tea.Msg {
		err := config.SetConfig("common.gas_station_mnemonic", mnemonic)
		if err != nil {
			return ui.ErrorLoading{Err: err}
		}
		time.Sleep(1500 * time.Millisecond)
		return ui.EndLoading{}
	}
}

func (hi *WeaveAppInitialization) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := hi.loading.Update(msg)
	hi.loading = loader
	if hi.loading.Completing {
		model := NewWeaveAppSettingUpGasStation(hi.Ctx)
		return model, model.Init()
	}
	return hi, cmd
}

func (hi *WeaveAppInitialization) View() string {
	state := weavecontext.GetCurrentState[ExistingCheckerState](hi.Ctx)
	return state.weave.Render() + "\n" + hi.loading.View()
}

type WeaveAppSettingUpGasStation struct {
	weavecontext.BaseModel
	loading ui.Loading
}

func NewWeaveAppSettingUpGasStation(ctx context.Context) tea.Model {
	return &WeaveAppSettingUpGasStation{
		BaseModel: weavecontext.BaseModel{
			Ctx:        ctx,
			CannotBack: true,
		},
		loading: ui.NewLoading("Setting up Gas Station account...", ui.DefaultWait()),
	}
}

func (hi *WeaveAppSettingUpGasStation) Init() tea.Cmd {
	return hi.loading.Init()
}

func (hi *WeaveAppSettingUpGasStation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := hi.loading.Update(msg)
	hi.loading = loader
	if hi.loading.Completing {
		model := NewWeaveAppSuccessfullyInitialized(hi.Ctx)
		return model, model.Init()
	}
	return hi, cmd
}

func (hi *WeaveAppSettingUpGasStation) View() string {
	state := weavecontext.GetCurrentState[ExistingCheckerState](hi.Ctx)
	return state.weave.Render() + "\n" + hi.loading.View() + "\n"
}

type WeaveAppSuccessfullyInitialized struct {
	weavecontext.BaseModel
	loadingEnded bool
}

func NewWeaveAppSuccessfullyInitialized(ctx context.Context) tea.Model {
	return &WeaveAppSuccessfullyInitialized{
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (hi *WeaveAppSuccessfullyInitialized) Init() tea.Cmd {
	return ui.DefaultWait()
}

func (hi *WeaveAppSuccessfullyInitialized) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case ui.EndLoading:
		hi.loadingEnded = true // Set loadingEnded to true
		return hi, tea.Quit
	}
	return hi, nil
}

func (hi *WeaveAppSuccessfullyInitialized) View() string {
	if hi.loadingEnded {
		return ""
	}
	state := weavecontext.GetCurrentState[ExistingCheckerState](hi.Ctx)
	return state.weave.Render() + "\n" + styles.FadeText("Initia is a network for interwoven rollups ") + "🪢\n"
}
