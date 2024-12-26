package analytics

const (
	AmplitudeKey = "aba1be3e2335dd5b8b060e977d93410b"

	// Component
	InitComponent       Component = "init"
	AnalyticsComponent  Component = "analytics"
	GasStationComponent Component = "gas-station"
	L1NodeComponent     Component = "l1-node"
	RollupComponent     Component = "rollup"
	OPinitComponent     Component = "opinit"
	RelayerComponent    Component = "relayer"
	UpgradeComponent    Component = "upgrade"

	// EventKeys
	ComponentEventKey string = "component"
	CommandEventKey   string = "command"
	OptionEventKey    string = "option"

	// Event
	RunEvent Event = "run"

	// Init Event
	InitActionSelected Event = "init-action-selected"
)

type Component string

type Event string
