package weaveinit

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/analytics"
	weavecontext "github.com/initia-labs/weave/context"
	"github.com/initia-labs/weave/models/initia"
	"github.com/initia-labs/weave/models/minitia"
	"github.com/initia-labs/weave/models/opinit_bots"
	"github.com/initia-labs/weave/models/relayer"
	"github.com/initia-labs/weave/styles"
	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/ui"
)

type State struct {
	weave types.WeaveState
}

func NewWeaveInitState() State {
	return State{
		weave: types.NewWeaveState(),
	}
}

func (e State) Clone() State {
	return State{
		weave: e.weave.Clone(),
	}
}

type WeaveInit struct {
	ui.Selector[Option]
	weavecontext.BaseModel
}

type Option string

const (
	RunL1NodeOption       Option = "Run an L1 node"
	LaunchNewRollupOption Option = "Launch a new rollup"
	RunOPBotsOption       Option = "Run OPinit bots"
	RunRelayerOption      Option = "Run a relayer"
)

func GetWeaveInitOptions() []Option {
	options := []Option{
		RunL1NodeOption,
		LaunchNewRollupOption,
		RunOPBotsOption,
		RunRelayerOption,
	}

	return options
}

func NewWeaveInit() *WeaveInit {
	ctx := weavecontext.NewAppContext(NewWeaveInitState())
	tooltips := []ui.Tooltip{
		ui.NewTooltip(string(RunL1NodeOption), "Bootstrap an Initia Layer 1 full node to be able to join the network whether it's mainnet, testnet, or your own local network. Weave also make state-syncing and automatic upgrades super easy for you.", "", []string{}, []string{}, []string{}),
		ui.NewTooltip(string(LaunchNewRollupOption), "Customize and deploy a new rollup on Initia in less than 5 minutes. This process includes configuring your rollup components (chain ID, gas, optimistic bridge, etc.) and fund OPinit bots to facilitate communications between your rollup and the underlying Initia L1.", "", []string{}, []string{}, []string{}),
		ui.NewTooltip(string(RunOPBotsOption), "Configure and run OPinit bots, the glue between rollup and the underlying Initia L1.", "", []string{}, []string{}, []string{}),
		ui.NewTooltip(string(RunRelayerOption), "Run a relayer to facilitate communications between your rollup and the underlying Initia L1.", "", []string{}, []string{}, []string{}),
	}

	return &WeaveInit{
		BaseModel: weavecontext.BaseModel{Ctx: ctx, CannotBack: true},
		Selector: ui.Selector[Option]{
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
	if model, cmd, handled := weavecontext.HandleCommonCommands[State](m, msg); handled {
		return model, cmd
	}

	selected, cmd := m.Select(msg)
	if selected != nil {
		switch *selected {
		case RunL1NodeOption:
			ctx := weavecontext.NewAppContext(initia.RunL1NodeState{})
			analytics.TrackEvent("run_l1_node_selected", map[string]interface{}{"component": "node"})
			model, err := initia.NewRunL1NodeNetworkSelect(ctx)
			if err != nil {
				return m, m.HandlePanic(err)
			}
			return model, nil
		case LaunchNewRollupOption:
			analytics.TrackEvent("launch_new_rollup_selected", map[string]interface{}{"component": "rollup"})
			minitiaChecker := minitia.NewExistingMinitiaChecker(weavecontext.NewAppContext(*minitia.NewLaunchState()))
			return minitiaChecker, minitiaChecker.Init()
		case RunOPBotsOption:
			analytics.TrackEvent("run_op_bots_selected", map[string]interface{}{"component": "opinit"})
			model, err := opinit_bots.NewOPInitBotInitSelector(weavecontext.NewAppContext(opinit_bots.NewOPInitBotsState()))
			if err != nil {
				return m, m.HandlePanic(err)
			}
			return model, nil
		case RunRelayerOption:
			analytics.TrackEvent("run_relayer_selected", map[string]interface{}{"component": "relayer"})
			model, err := relayer.NewRollupSelect(weavecontext.NewAppContext(relayer.NewRelayerState()))
			if err != nil {
				return m, m.HandlePanic(err)
			}
			return model, nil
		}
	}

	return m, cmd
}

func (m *WeaveInit) View() string {
	m.Selector.ViewTooltip(m.Ctx)
	return styles.RenderPrompt("What do you want to do?", []string{}, styles.Question) + m.Selector.View()
}
