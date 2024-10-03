package opinit_bots

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

type OPInitBotVersionSelector struct {
	utils.Selector[string]
	state    *OPInitBotsState
	question string
	versions utils.BinaryVersionWithDownloadURL
}

func NewOPInitBotVersionSelector(state *OPInitBotsState, versions utils.BinaryVersionWithDownloadURL) *OPInitBotVersionSelector {
	return &OPInitBotVersionSelector{
		Selector: utils.Selector[string]{
			Options: utils.SortVersions(versions),
		},
		state:    state,
		question: "Which OPInit bots version would you like to use?",
	}
}

func (m *OPInitBotVersionSelector) GetQuestion() string {
	return m.question
}

func (m *OPInitBotVersionSelector) Init() tea.Cmd {
	return nil
}

func (m *OPInitBotVersionSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.OPInitBotEndpoint = *selected
		m.state.OPInitBotVersion = m.versions[*selected]
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.DotsSeparator, m.GetQuestion(), []string{"OPInit bots"}, *selected))
		return NewSetupOPInitBotKeySelector(m.state), nil
	}

	return m, cmd
}

func (m *OPInitBotVersionSelector) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{"OPInit bots"}, styles.Question) + m.Selector.View()
}

type SetupOPInitBotKeySelector struct {
	utils.Selector[string]
	state    *OPInitBotsState
	question string
}

func NewSetupOPInitBotKeySelector(state *OPInitBotsState) *SetupOPInitBotKeySelector {
	return &SetupOPInitBotKeySelector{
		state: state,
		Selector: utils.Selector[string]{
			Options: []string{
				"Yes",
				"No",
			},
		},
		question: "Would you like to set up OPInit bot keys?",
	}
}

func (m *SetupOPInitBotKeySelector) GetQuestion() string {
	return m.question
}

func (m *SetupOPInitBotKeySelector) Init() tea.Cmd {
	return nil
}

func (m *SetupOPInitBotKeySelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case "Yes":
			return NewSetupBotCheckbox(m.state), nil
		case "No":
			return NewSetupOPInitBots(m.state), nil
		}
	}
	return m, cmd
}

func (m *SetupOPInitBotKeySelector) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + m.Selector.View()
}

func NextUpdateOpinitBotKey(state *OPInitBotsState) (tea.Model, tea.Cmd) {
	for idx := 0; idx < len(state.BotInfos); idx++ {
		if state.BotInfos[idx].IsSetup {
			return NewRecoverKeySelector(state, idx), nil
		}
	}
	model := NewSetupOPInitBots(state)
	return model, model.Init()
}

type SetupBotCheckbox struct {
	utils.CheckBox[string]
	state    *OPInitBotsState
	question string
}

func NewSetupBotCheckbox(state *OPInitBotsState) *SetupBotCheckbox {
	checkBlock := make([]string, 0)
	for idx, botInfo := range state.BotInfos {
		if !botInfo.IsNotExist {
			checkBlock = append(checkBlock, fmt.Sprintf("%s (already exist in keyring)", BotNames[idx]))
		} else {
			checkBlock = append(checkBlock, string(BotNames[idx]))
		}
	}

	return &SetupBotCheckbox{
		CheckBox: *utils.NewCheckBox(checkBlock),
		state:    state,
		question: "Which bots you wanna setup?",
	}
}

func (m *SetupBotCheckbox) GetQuestion() string {
	return m.question
}

func (m *SetupBotCheckbox) Init() tea.Cmd {
	return nil
}

func (m *SetupBotCheckbox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cb, cmd, done := m.Select(msg)
	if done {
		empty := true
		for idx, isSelected := range cb.Selected {
			if isSelected {
				empty = false
				m.state.BotInfos[idx].IsSetup = true
			}
		}
		if empty {
			return m, tea.Quit
		}
		return NextUpdateOpinitBotKey(m.state)
	}

	return m, cmd
}

func (m *SetupBotCheckbox) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{}, styles.Question) + "\n" + m.CheckBox.View()
}

type RecoverOption string

const (
	GenerateOption     RecoverOption = "Generate"
	FromMnemonicOption RecoverOption = "From Mnemonic"
)

type RecoverKeySelector struct {
	utils.Selector[RecoverOption]
	state    *OPInitBotsState
	idx      int
	question string
}

func NewRecoverKeySelector(state *OPInitBotsState, idx int) *RecoverKeySelector {
	return &RecoverKeySelector{
		Selector: utils.Selector[RecoverOption]{
			Options: []RecoverOption{
				GenerateOption,
				FromMnemonicOption,
			},
		},
		state:    state,
		idx:      idx,
		question: `Recover mode for key %s`,
	}
}

func (m *RecoverKeySelector) GetQuestion() string {
	return m.question
}

func (m *RecoverKeySelector) Init() tea.Cmd {
	return nil
}

func (m *RecoverKeySelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case GenerateOption:
			m.state.BotInfos[m.idx].IsGenerateKey = true
			m.state.BotInfos[m.idx].IsSetup = false
			if m.state.BotInfos[m.idx].BotName == BatchSubmitter {
				return NewDALayerSelector(m.state, m.idx), nil
			}
			return NextUpdateOpinitBotKey(m.state)
		case FromMnemonicOption:
			return NewRecoverFromMnemonic(m.state, m.idx), nil
		}
	}

	return m, cmd
}

func (m *RecoverKeySelector) View() string {
	botname := m.state.BotInfos[m.idx].BotName
	return styles.RenderPrompt(fmt.Sprintf(m.GetQuestion(), botname), []string{}, styles.Question) + "\n" + m.Selector.View()
}

type RecoverFromMnemonic struct {
	utils.TextInput
	question string
	state    *OPInitBotsState
	idx      int
}

func NewRecoverFromMnemonic(state *OPInitBotsState, idx int) *RecoverFromMnemonic {
	model := &RecoverFromMnemonic{
		TextInput: utils.NewTextInput(),
		state:     state,
		question:  "Input mnenomonic",
		idx:       idx,
	}
	model.WithPlaceholder("Enter in your mnemonic")
	return model
}

func (m *RecoverFromMnemonic) GetQuestion() string {
	return m.question
}

func (m *RecoverFromMnemonic) Init() tea.Cmd {
	return nil
}

func (m *RecoverFromMnemonic) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	input, cmd, done := m.TextInput.Update(msg)
	if done {
		m.state.BotInfos[m.idx].Mnemonic = input.Text
		m.state.BotInfos[m.idx].IsSetup = false
		if m.state.BotInfos[m.idx].BotName == BatchSubmitter {
			return NewDALayerSelector(m.state, m.idx), nil
		}
		return NextUpdateOpinitBotKey(m.state)
	}
	m.TextInput = input
	return m, cmd
}

func (m *RecoverFromMnemonic) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{"moniker"}, styles.Question) + m.TextInput.View()
}

type SetupOPInitBots struct {
	loading utils.Loading
	state   *OPInitBotsState
}

func NewSetupOPInitBots(state *OPInitBotsState) *SetupOPInitBots {
	return &SetupOPInitBots{
		state:   state,
		loading: utils.NewLoading("Checking for an existing Initia genesis file...", WaitSetupOPInitBots(state)),
	}
}

func (m *SetupOPInitBots) Init() tea.Cmd {
	return m.loading.Init()
}

func (m *SetupOPInitBots) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		return m, tea.Quit
	}
	return m, cmd
}

func (m *SetupOPInitBots) View() string {
	if m.loading.Completing {
		return strings.Join(m.state.SetupOpinitResponses, "\n")
	}
	return m.loading.View()
}

type DALayerOption string

const (
	InitiaLayerOption   DALayerOption = "initia"
	CelestiaLayerOption DALayerOption = "celestia"
)

type DALayerSelector struct {
	utils.Selector[DALayerOption]
	state *OPInitBotsState
	idx   int
}

func NewDALayerSelector(state *OPInitBotsState, idx int) *DALayerSelector {
	return &DALayerSelector{
		Selector: utils.Selector[DALayerOption]{
			Options: []DALayerOption{
				InitiaLayerOption,
				CelestiaLayerOption,
			},
		},
		state: state,
		idx:   idx,
	}
}

func (m *DALayerSelector) Init() tea.Cmd {
	return nil
}

func (m *DALayerSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.BotInfos[m.idx].DALayer = string(*selected)
		return NextUpdateOpinitBotKey(m.state)
	}

	return m, cmd
}

func (m *DALayerSelector) View() string {
	return styles.RenderPrompt("Please select DA layer", []string{}, styles.Question) + "\n" + m.Selector.View()
}

func WaitSetupOPInitBots(state *OPInitBotsState) tea.Cmd {
	return func() tea.Msg {
		for _, info := range state.BotInfos {
			if info.Mnemonic != "" {
				// TODO: recover celestia
				res, err := utils.RecoverKeyFromMnemonic("initiad", info.KeyName, info.Mnemonic)
				if err != nil {
					return utils.ErrorLoading{Err: err}
				}
				state.SetupOpinitResponses = append(state.SetupOpinitResponses, res)
			}
			if info.IsGenerateKey {
				res, err := utils.AddOrReplace("initiad", info.KeyName)
				if err != nil {
					return utils.ErrorLoading{Err: err}

				}
				state.SetupOpinitResponses = append(state.SetupOpinitResponses, res)
			}
		}

		return utils.EndLoading{}
	}
}
