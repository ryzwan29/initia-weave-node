package minitia

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/utils"
)

type ExistingMinitiaChecker struct {
	state   *LaunchState
	loading utils.Loading
}

func NewExistingMinitiaChecker(state *LaunchState) *ExistingMinitiaChecker {
	return &ExistingMinitiaChecker{
		state:   state,
		loading: utils.NewLoading("Checking for an existing Minitia app...", WaitExistingMinitiaChecker(state)),
	}
}

func (m *ExistingMinitiaChecker) Init() tea.Cmd {
	return m.loading.Init()
}

func WaitExistingMinitiaChecker(state *LaunchState) tea.Cmd {
	return func() tea.Msg {
		_, err := os.UserHomeDir()
		if err != nil {
			return utils.ErrorLoading{Err: err}
		}

		// TODO: Implement this
		time.Sleep(1500 * time.Millisecond)
		return utils.EndLoading{}

	}
}

func (m *ExistingMinitiaChecker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	loader, cmd := m.loading.Update(msg)
	m.loading = loader
	if m.loading.Completing {
		// TODO: Continue
	}
	return m, cmd
}

func (m *ExistingMinitiaChecker) View() string {
	return m.state.weave.Render() + "\n" + m.loading.View()
}
