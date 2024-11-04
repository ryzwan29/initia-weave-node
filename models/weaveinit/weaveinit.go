package weaveinit

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/flags"
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

func GetWeaveInitOptions() []WeaveInitOption {
	options := []WeaveInitOption{
		RunL1NodeOption,
	}

	if flags.IsEnabled(flags.MinitiaLaunch) {
		options = append(options, LaunchNewMinitiaOption)
	}

	if flags.IsEnabled(flags.OPInitBots) {
		options = append(options, InitializeOPBotsOption)
	}

	if flags.IsEnabled(flags.Relayer) {
		options = append(options, StartRelayerOption)
	}

	return options
}

func NewWeaveInit() *WeaveInit {
	return &WeaveInit{
		Selector: utils.Selector[WeaveInitOption]{
			Options: GetWeaveInitOptions(),
			Cursor:  0,
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
			ctx := utils.NewAppContext(initia.RunL1NodeState{})
			return initia.NewRunL1NodeNetworkSelect(ctx), nil
		case LaunchNewMinitiaOption:
			minitiaChecker := minitia.NewExistingMinitiaChecker(utils.NewAppContext(*minitia.NewLaunchState()))
			return minitiaChecker, minitiaChecker.Init()
		}
	}

	return m, cmd
}

func (m *WeaveInit) View() string {
	return styles.RenderPrompt("What action would you like to perform?", []string{}, styles.Question) + m.Selector.View()
}
