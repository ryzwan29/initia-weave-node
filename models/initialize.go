package models

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

func InitHeader() string {
	return styles.FadeText("Welcome to Weave! 🪢\n\n") +
		styles.RenderPrompt("As this is your first time using Weave, we ask that you set up your Gas Station account,\nwhich will hold the necessary funds for the OPinit-bots or relayer to send transactions.\n\n", []string{"Gas Station account"}, styles.Empty) +
		styles.BoldText("Please note that Weave will not send any transactions without your confirmation.\n", styles.Yellow) +
		styles.Text("While you can complete this setup later, we recommend doing it now to ensure a smoother experience.\n\n", styles.Gray)
}

type ExistingWeaveChecker struct {
	utils.Selector[ExistingWeaveOption]
	skipToModel tea.Model
}

type ExistingWeaveOption string

const (
	Yes ExistingWeaveOption = "Yes"
	No  ExistingWeaveOption = "No"
)

func NewExistingAppChecker(skipToModel tea.Model) *ExistingWeaveChecker {
	return &ExistingWeaveChecker{
		Selector: utils.Selector[ExistingWeaveOption]{
			Options: []ExistingWeaveOption{
				Yes,
				No,
			},
		},
		skipToModel: skipToModel,
	}
}

func (m *ExistingWeaveChecker) Init() tea.Cmd {
	return utils.DoTick()
}

func (m *ExistingWeaveChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case Yes:
			view := styles.RenderPreviousResponse(styles.ArrowSeparator, "Would you like to set up a Gas Station account", []string{"Gas Station account"}, string(*selected))
			return NewGasStationMnemonicInput(view), nil
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
	previousResponse string
	firstTime        bool
	utils.TextInput
}

func NewGasStationMnemonicInput(previousResponse string) *GasStationMnemonicInput {
	model := &GasStationMnemonicInput{
		previousResponse: previousResponse,
		firstTime:        utils.IsFirstTimeSetup(),
		TextInput:        utils.NewTextInput(true),
	}
	model.WithPlaceholder("Add mnemonic")
	model.WithValidatorFn(utils.ValidateMnemonic)
	return model
}

func (m *GasStationMnemonicInput) Init() tea.Cmd {
	return nil
}

func (m *GasStationMnemonicInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.previousResponse += styles.RenderPreviousResponse(styles.DotsSeparator, "Please set up a Gas Station account", []string{"Gas Station account"}, styles.HiddenMnemonicText)
		model := NewWeaveAppInitialization(m.previousResponse, input.Text)
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
	return header + m.previousResponse + styles.RenderPrompt("Please set up a Gas Station account", []string{"Gas Station account"}, styles.Question) + m.TextInput.View()
}

type WeaveAppInitialization struct {
	previousResponse string
	loading          utils.Loading
	mnemonic         string
}

func NewWeaveAppInitialization(prevRes, mnemonic string) tea.Model {
	return &WeaveAppInitialization{
		previousResponse: prevRes,
		loading:          utils.NewLoading("Initializing Weave...", WaitSetGasStation(mnemonic)),
		mnemonic:         mnemonic,
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
		model := NewWeaveAppSettingUpGasStation(hi.previousResponse)
		return model, model.Init()
	}
	return hi, cmd
}

func (hi *WeaveAppInitialization) View() string {
	return hi.previousResponse + "\n" + hi.loading.View()
}

type WeaveAppSettingUpGasStation struct {
	previousResponse string
	loading          utils.Loading
}

func NewWeaveAppSettingUpGasStation(prevRes string) tea.Model {
	return &WeaveAppSettingUpGasStation{
		previousResponse: prevRes,
		loading:          utils.NewLoading("Setting up Gas Station account...", utils.DefaultWait()),
	}
}

func (hi *WeaveAppSettingUpGasStation) Init() tea.Cmd {
	return hi.loading.Init()
}

func (hi *WeaveAppSettingUpGasStation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := hi.loading.Update(msg)
	hi.loading = loader
	if hi.loading.Completing {
		model := NewWeaveAppSuccessfullyInitialized(hi.previousResponse)
		return model, model.Init()
	}
	return hi, cmd
}

func (hi *WeaveAppSettingUpGasStation) View() string {
	return hi.previousResponse + "\n" + hi.loading.View()
}

type WeaveAppSuccessfullyInitialized struct {
	previousResponse string
	loadingEnded     bool
}

func NewWeaveAppSuccessfullyInitialized(prevRes string) tea.Model {
	return &WeaveAppSuccessfullyInitialized{
		previousResponse: prevRes,
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
	return hi.previousResponse + "\n" + styles.FadeText("Initia is a network for interwoven rollups ") + "🪢\n"
}
