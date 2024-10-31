package initia

import (
	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/types"
)

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
}

func NewRunL1NodeState() *RunL1NodeState {
	return &RunL1NodeState{
		weave: types.NewWeaveState(),
	}
}
