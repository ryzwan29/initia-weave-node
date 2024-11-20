package models

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"time"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/utils"
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
		styles.RenderPrompt("As this is your first time using Weave, we ask that you set up your Gas Station account,\nwhich will hold the necessary funds for the OPinit-bots or relayer to send transactions.\n\n", []string{"Gas Station account"}, styles.Empty) +
		styles.BoldText("Please note that Weave will not send any transactions without your confirmation.\n", styles.Yellow) +
		styles.Text("While you can complete this setup later, we recommend doing it now to ensure a smoother experience.\n\n", styles.Gray)
}

type ExistingWeaveChecker struct {
	utils.BaseModel
	utils.Selector[ExistingWeaveOption]
	skipToModel tea.Model
}

type ExistingWeaveOption string

const (
	Yes ExistingWeaveOption = "Yes"
	No  ExistingWeaveOption = "No"
)

func NewExistingAppChecker(ctx context.Context, skipToModel tea.Model) *ExistingWeaveChecker {
	return &ExistingWeaveChecker{
		Selector: utils.Selector[ExistingWeaveOption]{
			Options: []ExistingWeaveOption{
				Yes,
				No,
			},
			CannotBack: true,
		},
		BaseModel:   utils.BaseModel{Ctx: ctx, CannotBack: true},
		skipToModel: skipToModel,
	}
}

func (m *ExistingWeaveChecker) Init() tea.Cmd {
	return utils.DoTick()
}

func (m *ExistingWeaveChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[ExistingCheckerState](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		state := utils.PushPageAndGetState[ExistingCheckerState](m)

		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.ArrowSeparator, "Would you like to set up a Gas Station account", []string{"Gas Station account"}, string(*selected)),
		)
		m.Ctx = utils.SetCurrentState(m.Ctx, state)
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
	view += styles.RenderPrompt("Would you like to set up a Gas Station account", []string{"Gas Station account"}, styles.Question) +
		" " + styles.Text("(The account that will hold the funds required by the OPinit-bots or relayer to send transactions)", styles.Gray)
	view += m.Selector.View()

	return view
}

type GasStationMnemonicInput struct {
	utils.BaseModel
	firstTime bool
	utils.TextInput
}

func NewGasStationMnemonicInput(ctx context.Context) *GasStationMnemonicInput {
	tooltip := styles.NewTooltip(
		"Gas station account",
		"Gas Station account is the account from which Weave will use to fund necessary system accounts, enabling easier and more seamless experience while setting up things using Weave.",
		"** Weave will NOT automatically send transactions without asking for your confirmation. **",
		[]string{}, []string{}, []string{})
	model := &GasStationMnemonicInput{
		firstTime: utils.IsFirstTimeSetup(),
		TextInput: utils.NewTextInput(true),
		BaseModel: utils.BaseModel{Ctx: ctx},
	}
	model.WithPlaceholder("Add mnemonic")
	model.WithValidatorFn(utils.ValidateMnemonic)
	model.WithTooltip(&tooltip)
	return model
}

func (m *GasStationMnemonicInput) Init() tea.Cmd {
	return nil
}

func (m *GasStationMnemonicInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[ExistingCheckerState](m, msg); handled {
		return model, cmd
	}
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		state := utils.PushPageAndGetState[ExistingCheckerState](m)

		state.weave.PushPreviousResponse(
			styles.RenderPreviousResponse(styles.DotsSeparator, "Please set up a Gas Station account", []string{"Gas Station account"}, styles.HiddenMnemonicText),
		)
		model := NewWeaveAppInitialization(utils.SetCurrentState(m.Ctx, state), input.Text)
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
	state := utils.GetCurrentState[ExistingCheckerState](m.Ctx)
	m.TextInput.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return header + state.weave.Render() + styles.RenderPrompt("Please set up a Gas Station account", []string{"Gas Station account"}, styles.Question) + m.TextInput.View()
}

type WeaveAppInitialization struct {
	utils.BaseModel
	loading  utils.Loading
	mnemonic string
}

func NewWeaveAppInitialization(ctx context.Context, mnemonic string) tea.Model {
	return &WeaveAppInitialization{
		loading:  utils.NewLoading("Initializing Weave...", WaitSetGasStation(mnemonic)),
		mnemonic: mnemonic,
		BaseModel: utils.BaseModel{
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
		err := utils.SetConfig("common.gas_station_mnemonic", mnemonic)
		if err != nil {
			return utils.ErrorLoading{Err: err}
		}
		time.Sleep(1500 * time.Millisecond)
		return utils.EndLoading{}
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
	state := utils.GetCurrentState[ExistingCheckerState](hi.Ctx)
	return state.weave.Render() + "\n" + hi.loading.View()
}

type WeaveAppSettingUpGasStation struct {
	utils.BaseModel
	loading utils.Loading
}

func NewWeaveAppSettingUpGasStation(ctx context.Context) tea.Model {
	return &WeaveAppSettingUpGasStation{
		BaseModel: utils.BaseModel{
			Ctx:        ctx,
			CannotBack: true,
		},
		loading: utils.NewLoading("Setting up Gas Station account...", utils.DefaultWait()),
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
	state := utils.GetCurrentState[ExistingCheckerState](hi.Ctx)
	return state.weave.Render() + "\n" + hi.loading.View() + "\n"
}

type WeaveAppSuccessfullyInitialized struct {
	utils.BaseModel
	loadingEnded bool
}

func NewWeaveAppSuccessfullyInitialized(ctx context.Context) tea.Model {
	return &WeaveAppSuccessfullyInitialized{
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
	}
}

func (hi *WeaveAppSuccessfullyInitialized) Init() tea.Cmd {
	return utils.DefaultWait()
}

func (hi *WeaveAppSuccessfullyInitialized) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case utils.EndLoading:
		hi.loadingEnded = true // Set loadingEnded to true
		return hi, tea.Quit
	}
	return hi, nil
}

func (hi *WeaveAppSuccessfullyInitialized) View() string {
	if hi.loadingEnded {
		return ""
	}
	state := utils.GetCurrentState[ExistingCheckerState](hi.Ctx)
	return state.weave.Render() + "\n" + styles.FadeText("Initia is a network for interwoven rollups ") + "🪢\n"
}
