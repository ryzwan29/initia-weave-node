package weaveinit

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/initia-labs/weave/utils"
)

type WeaveInit struct {
	utils.Selector[WeaveInitOption]
}

type WeaveInitOption string

const (
	RunL1NodeOption        WeaveInitOption = "Run L1 Node"
	LaunchNewMinitiaOption WeaveInitOption = "Launch New Minitia"
	InitializeOPBotsOption WeaveInitOption = "Initialize OP Bots"
	StartRelayerOption     WeaveInitOption = "Start a Relayer"
)

func NewWeaveInit() *WeaveInit {
	return &WeaveInit{
		Selector: utils.Selector[WeaveInitOption]{
			Options: []WeaveInitOption{
				RunL1NodeOption,
				LaunchNewMinitiaOption,
				InitializeOPBotsOption,
				StartRelayerOption,
			},
			Cursor: 0,
		},
	}
}

func (m *WeaveInit) Init() tea.Cmd {
	return nil
}

func (m *WeaveInit) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case RunL1NodeOption:
			return NewRunL1NodeNetworkSelect(&RunL1NodeState{}), nil
		}
	}

	return m, cmd
}

func (m *WeaveInit) View() string {
	view := "? What action would you like to perform?\n"
	for i, option := range m.Options {
		if i == m.Cursor {
			view += "(■) " + string(option) + "\n"
		} else {
			view += "( ) " + string(option) + "\n"
		}
	}
	return view + "\nPress Enter to select, or q to quit."
}
