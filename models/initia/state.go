package initia

import (
	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/types"
)

// RunL1NodeState represents the configuration state of a Layer 1 Node
type RunL1NodeState struct {
	weave                             types.WeaveState
	network                           string
	chainRegistry                     *registry.ChainRegistry // We can store the registry here since we only need one
	initiadVersion                    string
	initiadEndpoint                   string
	chainId                           string
	moniker                           string
	existingApp                       bool
	replaceExistingApp                bool
	minGasPrice                       string
	enableLCD                         bool
	enableGRPC                        bool
	seeds                             string
	persistentPeers                   string
	existingGenesis                   bool
	genesisEndpoint                   string
	existingData                      bool
	syncMethod                        string
	replaceExistingData               bool
	replaceExistingGenesisWithDefault bool
	snapshotEndpoint                  string
	stateSyncEndpoint                 string
	additionalStateSyncPeers          string
	allowAutoUpgrade                  bool
	pruning                           string
}

// NewRunL1NodeState initializes a new RunL1NodeState with default values.
func NewRunL1NodeState() RunL1NodeState {
	return RunL1NodeState{
		weave: types.NewWeaveState(),
	}
}

// Clone creates a deep copy of RunL1NodeState without pointers.
func (s RunL1NodeState) Clone() RunL1NodeState {
	return RunL1NodeState{
		weave:                             s.weave.Clone(), // Assuming WeaveState has a Clone method
		network:                           s.network,
		chainRegistry:                     s.chainRegistry,
		initiadVersion:                    s.initiadVersion,
		initiadEndpoint:                   s.initiadEndpoint,
		chainId:                           s.chainId,
		moniker:                           s.moniker,
		existingApp:                       s.existingApp,
		replaceExistingApp:                s.replaceExistingApp,
		minGasPrice:                       s.minGasPrice,
		enableLCD:                         s.enableLCD,
		enableGRPC:                        s.enableGRPC,
		seeds:                             s.seeds,
		persistentPeers:                   s.persistentPeers,
		existingGenesis:                   s.existingGenesis,
		genesisEndpoint:                   s.genesisEndpoint,
		existingData:                      s.existingData,
		syncMethod:                        s.syncMethod,
		replaceExistingData:               s.replaceExistingData,
		replaceExistingGenesisWithDefault: s.replaceExistingGenesisWithDefault,
		snapshotEndpoint:                  s.snapshotEndpoint,
		stateSyncEndpoint:                 s.stateSyncEndpoint,
		additionalStateSyncPeers:          s.additionalStateSyncPeers,
		allowAutoUpgrade:                  s.allowAutoUpgrade,
		pruning:                           s.pruning,
	}
}
