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
	GenerateKeyfileKey string = "generate-key-file"
	KeyFileKey         string = "key-file"
	WithConfigKey      string = "with-config"
	VmTypeKey          string = "vm-type"

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
	OpBridgeBatchSubmissionTargetSelected Event = "op-bridge-batch-sumission-selected"
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

	// GasStation
	GasStationMethodSelected Event = "gas-station-method-selected"

	// WithConfig

)

type Component string

type Event string
