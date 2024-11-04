package weaveinit

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/flags"
	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/models/minitia"
	"github.com/initia-labs/weave/models/opinit_bots"
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
	SetupOPBotsKeys        WeaveInitOption = "Setup OPInit Bots Keys"
	InitializeOPBotsOption WeaveInitOption = "Initialize OPInit Bots"
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
		options = append(options, SetupOPBotsKeys)
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
			Options:    GetWeaveInitOptions(),
			Cursor:     0,
			CannotBack: true,
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
		case SetupOPBotsKeys:
			versions, currentVersion := utils.GetOPInitVersions()
			ctx := utils.NewAppContext(opinit_bots.NewOPInitBotsState())
			return opinit_bots.NewOPInitBotVersionSelector(ctx, versions, currentVersion), nil
		case InitializeOPBotsOption:
			return opinit_bots.NewOPInitBotInitSelector(utils.NewAppContext(opinit_bots.NewOPInitBotsState())), nil
		}
	}

	return m, cmd
}

func (m *WeaveInit) View() string {
	return styles.RenderPrompt("What action would you like to perform?", []string{}, styles.Question) + m.Selector.View()
}
