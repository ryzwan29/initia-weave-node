package tooltip

import (
	"github.com/initia-labs/weave/ui"
)

var (
	L1ChainIdTooltip        = ui.NewTooltip("L1 chain ID", ChainIDDescription("L1"), "", []string{}, []string{}, []string{})
	L1RPCEndpointTooltip    = ui.NewTooltip("L1 RPC endpoint", RPCEndpointDescription("L1"), "", []string{}, []string{}, []string{})
	L1NetworkSelectTooltip  = ui.NewTooltip("Network to participate", "Available options are Testnet, and local which means no network participation, no state syncing needed, but fully customizable (often used for development and testing purposes)", "", []string{}, []string{}, []string{})
	L1InitiadVersionTooltip = ui.NewTooltip("Initiad version", "Initiad version refers to the version of the Initia daemon, which is software used to run an Initia Layer 1 node.", "", []string{}, []string{}, []string{})
	L1ExistingAppTooltip    = ui.NewTooltip("app.toml / config.toml", "app.toml contains application-specific configurations for the blockchain node, such as transaction limits, gas price, state pruning strategy.\n\nconfig.toml contains core network and protocol settings for the node, such as peers to connect to, timeouts, consensus configurations, etc.", "", []string{"app.toml", "config.toml"}, []string{}, []string{})
	L1MinGasPriceTooltip    = ui.NewTooltip("Minimum Gas Price", MinGasPriceDescription("L1"), "", []string{}, []string{}, []string{})

	// Enable Features Tooltips
	L1EnableRESTTooltip = ui.NewTooltip("REST", "Enabling this option allows REST API calls to query data and submit transactions to your node. \n\nEnabling this is recommended.", "", []string{}, []string{}, []string{})
	L1EnablegRPCTooltip = ui.NewTooltip("gRPC", "Enabling this option allows gRPC calls to your node. \n\nEnabling this is recommended.", "", []string{}, []string{}, []string{})

	L1SeedsTooltip           = ui.NewTooltip("Seeds", "Configure known nodes (<node-id>@<IP>:<port>) as initial contact points, mainly used to discover other nodes. If you're don't need your node to participate in the network, seeds may be unnecessary.\n\nThis step is optional but can quickly get your node up to date.", "", []string{}, []string{}, []string{})
	L1PersistentPeersTooltip = ui.NewTooltip("Persistent Peers", "Configure nodes (<node-id>@<IP>:<port>) to maintain constant connections. This is particularly useful for fast syncing if you have access to a trusted, reliable node.\n\nThis step is optional but can expedite the process of getting your node up to date.", "", []string{}, []string{}, []string{})
	L1GenesisEndpointTooltip = ui.NewTooltip("genesis.json", "Provide the URL or network address where the genesis.json file can be accessed. This file should contains the initial state and configuration of the blockchain network, which is essential for new nodes to sync and participate in the network correctly.", "", []string{}, []string{}, []string{})

	// Sync Method Tooltips
	L1SnapshotSyncTooltip = ui.NewTooltip("Snapshot", "Downloads a recent state snapshot to quickly catch up without replaying all history. This is faster than full sync but relies on a trusted source for the snapshot.\n\nThis is necessary to participate in an existing network.", "", []string{}, []string{}, []string{})
	L1StateSyncTooltip    = ui.NewTooltip("State Sync", "Retrieves the latest blockchain state from peers without downloading the entire history. It's faster than syncing from genesis but may miss some historical data.\n\nThis is necessary to participate in an existing network.", "", []string{}, []string{}, []string{})
	L1NoSyncTooltip       = ui.NewTooltip("No Sync", "The node will not download data from any sources to replace the existing (if any). The node will start syncing from its current state, potentially genesis state if this is the first run.\n\nThis is best for local development / testing.", "", []string{}, []string{}, []string{})

	// Cosmovisor Tooltips
	L1CosmovisorAutoUpgradeEnableTooltip  = ui.NewTooltip("Enable", "Enable automatic downloading of new binaries and upgrades via Cosmovisor. \nSee more: https://docs.initia.xyz/run-initia-node/automating-software-updates-with-cosmovisor", "", []string{}, []string{}, []string{})
	L1CosmovisorAutoUpgradeDisableTooltip = ui.NewTooltip("Disable", "Disable automatic downloading of new binaries and upgrades via Cosmovisor. You will need to manually upgrade the binaries and restart the node to apply the upgrades.", "", []string{}, []string{}, []string{})
)
