package models

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

type ExistingAppChecker struct {
	TextInput utils.TextInput
}

func NewExistingAppChecker() *ExistingAppChecker {
	return &ExistingAppChecker{}
}

func (m *ExistingAppChecker) Init() tea.Cmd {
	return utils.DoTick()
}

func (m *ExistingAppChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		view := styles.RenderPreviousResponse(styles.ArrowSeparator, "Please set up a Gas Station account", []string{"Gas Station account"}, input.Text)
		model := NewWeaveAppInitialization(view, input.Text)
		return model, model.Init()
	}
	m.TextInput = input
	return m, cmd
}

func (m *ExistingAppChecker) View() string {
	view := styles.DefaultText("Hi 👋🏻 Weave is a CLI for managing Initia deployments.\n\n")

	view += styles.RenderPrompt("Please set up a Gas Station account", []string{"Gas Station account"}, styles.Question) +
		" " + styles.Text("(The account that will hold the funds required by the OPinit-bots or relayer to send transactions)", styles.Gray)
	view += "\n" + m.TextInput.View()

	return view
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
