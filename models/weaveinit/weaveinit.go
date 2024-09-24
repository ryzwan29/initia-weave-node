package weaveinit

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/models/minitia"
	"github.com/initia-labs/weave/styles"
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
			return initia.NewRunL1NodeNetworkSelect(&initia.RunL1NodeState{}), nil
		case LaunchNewMinitiaOption:
			minitiaChecker := minitia.NewExistingMinitiaChecker(&minitia.LaunchState{})
			return minitiaChecker, minitiaChecker.Init()
		}
	}

	return m, cmd
}

func (m *WeaveInit) View() string {
	return styles.RenderPrompt("What action would you like to perform?", []string{}, styles.Question) + m.Selector.View()
}
