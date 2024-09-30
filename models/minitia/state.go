package minitia

import "github.com/initia-labs/weave/types"

type GenesisAccount struct {
	address string
	balance string
}

type LaunchState struct {
	weave                            types.WeaveState
	existingMinitiaApp               bool
	l1Network                        string
	vmType                           string
	minitiadVersion                  string
	minitiadEndpoint                 string
	chainId                          string
	gasDenom                         string
	moniker                          string
	genesisAccounts                  []GenesisAccount
	preGenesisAccountsResponsesCount int
	opBridgeSubmissionInterval       string
}

func NewLaunchState() *LaunchState {
	return &LaunchState{
		weave: types.NewWeaveState(),
	}
}
