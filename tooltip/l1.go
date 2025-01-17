package tooltip

import (
	"github.com/initia-labs/weave/ui"
)

var (
	L1ChainIdTooltip        = ui.NewTooltip("L1 chain ID", ChainIDDescription("L1"), "", []string{}, []string{}, []string{})
	L1RPCEndpointTooltip    = ui.NewTooltip("L1 RPC endpoint", RPCEndpointDescription("L1"), "", []string{}, []string{}, []string{})
	L1NetworkSelectTooltip  = ui.NewTooltip("Network to connect to", "Available options are Testnet (initiation-2) and local (no network participation).", "", []string{}, []string{}, []string{})
	L1InitiadVersionTooltip = ui.NewTooltip("Initiad version", "Initiad version refers to the version of the Initia daemon CLI used to run the Initia L1 node.", "", []string{}, []string{}, []string{})
	L1ExistingAppTooltip    = ui.NewTooltip("app.toml / config.toml", "The app.toml file contains the node's configuration, including transaction limits, gas price, and state pruning strategy.\n\nThe config.toml file includes core network and protocol settings for the node, such as peers to connect to, timeouts, and consensus configurations.", "", []string{"app.toml", "config.toml"}, []string{}, []string{})
	L1MinGasPriceTooltip    = ui.NewTooltip("Minimum Gas Price", MinGasPriceDescription("L1"), "", []string{}, []string{}, []string{})

	// Enable Features Tooltips
	L1EnableRESTTooltip = ui.NewTooltip("REST", "Enabling this option allows REST API calls to query data and submit transactions to your node. (Recommended)", "", []string{}, []string{}, []string{})
	L1EnablegRPCTooltip = ui.NewTooltip("gRPC", "Enabling this option allows gRPC calls to your node. (Recommended)", "", []string{}, []string{}, []string{})

	L1SeedsTooltip           = ui.NewTooltip("Seeds", "Enter a list of known node addresses (<node-id>@<IP>:<port>) to be used as initial contact points to discover other nodes. If you don't need your node to participate in the network (e.g. local development), seeds are not required.", "", []string{}, []string{}, []string{})
	L1PersistentPeersTooltip = ui.NewTooltip("Persistent Peers", "Enter a list of known node addresses (<node-id>@<IP>:<port>) to maintain constant connections to. This is particularly useful for fast syncing if you have access to a trusted, reliable node.", "", []string{}, []string{}, []string{})
	L1GenesisEndpointTooltip = ui.NewTooltip("genesis.json", "Provide the URL or network address where the genesis.json file can be accessed. This file contains the initial state and configuration of the blockchain network, which is essential for new nodes to sync and participate in the network correctly.", "", []string{}, []string{}, []string{})

	// Sync Method Tooltips
	L1SnapshotSyncTooltip = ui.NewTooltip("Snapshot", "Downloads a recent snapshot of the chain state to quickly catch up without replaying the entire chain history. This is faster than full state sync but relies on a trusted source for the snapshot.\n\nThis is necessary to participate in an existing network.", "", []string{}, []string{}, []string{})
	L1StateSyncTooltip    = ui.NewTooltip("State Sync", "Retrieves the latest blockchain state from peers without downloading the entire history. It's faster than syncing from genesis but may miss some historical data.\n\nThis is necessary to participate in an existing network.", "", []string{}, []string{}, []string{})
	L1NoSyncTooltip       = ui.NewTooltip("No Sync", "The node will not download data from any sources to replace the existing (if any). The node will start syncing from its current state, potentially genesis state if this is the first run.\n\nThis is best for local development / testing.", "", []string{}, []string{}, []string{})

	// Cosmovisor Tooltips
	L1CosmovisorAutoUpgradeEnableTooltip  = ui.NewTooltip("Enable", "Enable automatic downloading of new binaries and upgrades via Cosmovisor. \nSee more: https://docs.initia.xyz/run-initia-node/automating-software-updates-with-cosmovisor", "", []string{}, []string{}, []string{})
	L1CosmovisorAutoUpgradeDisableTooltip = ui.NewTooltip("Disable", "Disable automatic downloading of new binaries and upgrades via Cosmovisor. You will need to manually upgrade the binaries and restart the node to apply the upgrades.", "", []string{}, []string{}, []string{})

	L1DefaultPruningStrategiesTooltip    = ui.NewTooltip("Default", "Keep the last 100 states in addition to every 500th state, and prune on 10-block intervals. This configuration is safe to use on all types of nodes, especially validator nodes.", "", []string{}, []string{"recommended"}, []string{})
	L1NothingPruningStrategiesTooltip    = ui.NewTooltip("Nothing", "Disable node state pruning, essentially making your node an archival node. This mode consumes the highest disk usage.", "", []string{}, []string{"disable"}, []string{})
	L1EverythingPruningStrategiesTooltip = ui.NewTooltip("Everything", "Keep the current state and also prune on 10 blocks intervals. This settings is useful for nodes such as seed/sentry nodes, as long as they are not used to query RPC/REST API requests. This mode is not recommended when running validator nodes.", "", []string{}, []string{"not recommended "}, []string{})
)
