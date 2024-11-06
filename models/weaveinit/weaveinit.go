package weaveinit

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/flags"
	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/models/minitia"
	"github.com/initia-labs/weave/models/opinit_bots"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/utils"
)

type WeaveInitState struct {
	weave types.WeaveState
}

func NewWeaveInitState() WeaveInitState {
	return WeaveInitState{
		weave: types.NewWeaveState(),
	}
}

func (e WeaveInitState) Clone() WeaveInitState {
	return WeaveInitState{
		weave: e.weave.Clone(),
	}
}

type WeaveInit struct {
	utils.Selector[WeaveInitOption]
	utils.BaseModel
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
	ctx := utils.NewAppContext(NewWeaveInitState())
	tooltips := []styles.Tooltip{
		styles.NewTooltip(string(RunL1NodeOption), "Bootstrap an Initia Layer 1 full node to be able to join the network whether it's Mainnet, Testnet, or your own local network. Weave also make state-syncing super easy for you.", "", []string{}, []string{}, []string{}),
		styles.NewTooltip(string(LaunchNewMinitiaOption), "Customize and deploy a new Minitia, an L2 rollup on Initia in less than 5 minutes. This process includes configuring your L2 components (chain-id, gas, optimistic bridge, etc.) and fund OPinit Bots to facilitate communications between your Minitia and the underlying Initia L1.", "", []string{}, []string{}, []string{}),
		styles.NewTooltip(string(SetupOPBotsKeys), "TBD", "", []string{}, []string{}, []string{}),
		styles.NewTooltip(string(InitializeOPBotsOption), "Configure and run OPinit Bots, the glue between Minitia and the underlying Initia L1.", "", []string{}, []string{}, []string{}),
	}

	return &WeaveInit{
		BaseModel: utils.BaseModel{Ctx: ctx, CannotBack: true},
		Selector: utils.Selector[WeaveInitOption]{
			Options:    GetWeaveInitOptions(),
			Cursor:     0,
			CannotBack: true,
			Tooltips:   &tooltips,
		},
	}
}

func (m *WeaveInit) Init() tea.Cmd {
	return nil
}

func (m *WeaveInit) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if model, cmd, handled := utils.HandleCommonCommands[WeaveInitState](m, msg); handled {
		return model, cmd
	}

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
	m.Selector.ToggleTooltip = utils.GetTooltip(m.Ctx)
	return styles.RenderPrompt("What action would you like to perform?", []string{}, styles.Question) + m.Selector.View()
}
