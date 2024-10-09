package opinit_bots

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/utils"
)

type OPInitBotInitOption string

const (
	Executor_OPInitBotInitOption   OPInitBotInitOption = "Executor"
	Challenger_OPInitBotInitOption OPInitBotInitOption = "Challenger"
)

type OPInitBotInitSelector struct {
	utils.Selector[OPInitBotInitOption]
	state    *OPInitBotsState
	question string
}

func NewOPInitBotInitSelector(state *OPInitBotsState) *OPInitBotInitSelector {
	return &OPInitBotInitSelector{
		Selector: utils.Selector[OPInitBotInitOption]{
			Options: []OPInitBotInitOption{Executor_OPInitBotInitOption, Challenger_OPInitBotInitOption},
		},
		state:    state,
		question: "Which bot would you like to run?",
	}
}

func (m *OPInitBotInitSelector) GetQuestion() string {
	return m.question
}

func (m *OPInitBotInitSelector) Init() tea.Cmd {
	return nil
}

func (m *OPInitBotInitSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		m.state.weave.PushPreviousResponse(styles.RenderPreviousResponse(styles.ArrowSeparator, m.GetQuestion(), []string{"bot"}, string(*selected)))

		switch *selected {
		case Executor_OPInitBotInitOption:
			return m, cmd
		case Challenger_OPInitBotInitOption:
			return m, cmd

		}
	}

	return m, cmd
}

func (m *OPInitBotInitSelector) View() string {
	return styles.RenderPrompt(m.GetQuestion(), []string{"bot"}, styles.Question) + m.Selector.View()
}
