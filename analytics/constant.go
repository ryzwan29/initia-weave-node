package analytics

const (
	AmplitudeKey = "aba1be3e2335dd5b8b060e977d93410b"

	// Component
	AnalyticsComponent  Component = "analytics"
	GasStationComponent Component = "gas-station"
	L1NodeComponent     Component = "l1-node"
	RollupComponent     Component = "rollup"
	OPinitComponent     Component = "opinit"
	RelayerComponent    Component = "relayer"
	UpgradeComponent    Component = "upgrade"

	// EventKeys
	ComponentEventKey  string = "component"
	FeatureEventKey    string = "feature"
	CommandEventKey    string = "command"
	OptionEventKey     string = "option"
	L1ChainIdEventKey  string = "l1-chain-id"
	ErrorEventKey      string = "panic-error"
	ModelNameKey       string = "model-name"
	GenerateKeyfileKey string = "generate-key-file"
	KeyFileKey         string = "key-file"
	WithConfigKey      string = "with-config"
	VmKey              string = "vm"
	BotTypeKey         string = "bot-type"

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

	// Rollup Event
	VmTypeSelected                        Event = "vm-type-selected"
	OpBridgeBatchSubmissionTargetSelected Event = "op-bridge-batch-submission-selected"
	EnableOracleSelected                  Event = "enable-oracle-selected"
	SystemKeysSelected                    Event = "system-keys-selected"
	AccountsFundingPresetSelected         Event = "accounts-funding-preset-selected"
	AddGasStationToGenesisSelected        Event = "add-gas-station-to-genesis-selected"
	AddGenesisAccountsSelected            Event = "add-genesis-accounts-selected"

	// Opinit Event
	OPInitBotInitSelected           Event = "opinit-bot-init-selected"
	ResetDBSelected                 Event = "reset-db-selected"
	UseCurrentConfigSelected        Event = "use-current-config-selected"
	PrefillFromArtifactsSelected    Event = "prefill-from-artifacts-selected"
	L1PrefillSelected               Event = "l1-prefill-selected"
	DALayerSelected                 Event = "da-layer-selected"
	ImportKeysFromArtifactsSelected Event = "import-keys-from-artifacts-selected"
	RecoverKeySelected              Event = "recover-key-selected"

	// Relayer
	RelayerRollupSelected              Event = "relayer-rollup-selected"
	RelayerL2Selected                  Event = "relayer-l2-selected"
	SettingUpIBCChannelsMethodSelected Event = "setting-up-ibc-channels-method-selected"
	IBCChannelsSelected                Event = "ibc-channel-selected"
	UseChallengerKeySelected           Event = "use-challenger-key-selected"

	// GasStation
	GasStationMethodSelected Event = "gas-station-method-selected"
)

var (
	SetupL1NodeFeature     Feature = Feature{Name: "setup-l1-node", Component: L1NodeComponent}
	RollupLaunchFeature    Feature = Feature{Name: "launch-rollup", Component: RollupComponent}
	SetupOPinitBotFeature  Feature = Feature{Name: "setup-opinit-bot", Component: OPinitComponent}
	SetupOPinitKeysFeature Feature = Feature{Name: "setup-opinit-keys", Component: OPinitComponent}
	ResetOPinitBotFeature  Feature = Feature{Name: "reset-opinit-bot", Component: OPinitComponent}
	SetupGasStationFeature Feature = Feature{Name: "setup-gas-station", Component: GasStationComponent}
	SetupRelayerFeature    Feature = Feature{Name: "setup-relayer", Component: RelayerComponent}
	OptOutAnalyticsFeature Feature = Feature{Name: "opt-out-analytics", Component: AnalyticsComponent}
)

type Component string

type Feature struct {
	Name      string
	Component Component
}

type Event string
