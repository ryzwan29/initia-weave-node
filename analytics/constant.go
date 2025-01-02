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
	ComponentEventKey  string = "component"
	CommandEventKey    string = "command"
	OptionEventKey     string = "option"
	L1ChainIdEventKey  string = "l1-chain-id"
	L1ExistingEventKey string = "existing-l1-app"
	ErrorEventKey      string = "panic-error"
	L1NodeVersionKey   string = "l1-node-version"
	EmptyInputKey      string = "empty-input"
	ModelNameKey       string = "model-name"

	// Event
	RunEvent       Event = "run"
	CompletedEvent Event = "completed"

	// Init Event
	InitActionSelected             Event = "init-action-selected"
	ExistingAppReplaceSelected     Event = "existing-app-replace-selected"
	L1NetworkSelected              Event = "l1-network-selected"
	Interrupted                    Event = "interrupted"
	Panicked                       Event = "panicked"
	L1NodeVersionSelected          Event = "l1-node-version-selected"
	PruningStrategySelected        Event = "pruning-strategy-selected"
	ExistingGenesisReplaceSelected Event = "existing-genesis-replace-selected"
	SyncMethodSelected             Event = "sync-method-selected"
	CosmovisorAutoUpgradeSelected  Event = "cosmovisor-auto-upgrade-selected"
	ExistingDataReplaceSelected    Event = "existing-data-replace-selected"
	FeaturesEnabled                Event = "feature-enabled"
)

type Component string

type Event string

type EventAtrriutes map[string]interface{}
