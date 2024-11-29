package models

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/common"
	"github.com/initia-labs/weave/config"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/ui"
)

type ExistingCheckerState struct {
	weave types.WeaveState
}

func NewExistingCheckerState() ExistingCheckerState {
	return ExistingCheckerState{
		weave: types.NewWeaveState(),
	}
}

func (e ExistingCheckerState) Clone() ExistingCheckerState {
	return ExistingCheckerState{
		weave: e.weave.Clone(),
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
			return NewGasStationMnemonicInput(m.Ctx), nil
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
