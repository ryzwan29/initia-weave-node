package minitia

import "github.com/initia-labs/weave/types"

type LaunchState struct {
	weave              types.WeaveState
	existingMinitiaApp bool
	l1ChainId          string
	l1RPC              string
	vmType             string
	minitiadVersion    string
	minitiadEndpoint   string
	chainId            string
	gasDenom           string
	moniker            string
	enableOracle       bool
	genesisAccounts    GenesisAccounts

	opBridgeSubmissionInterval       string
	opBridgeOutputFinalizationPeriod string
	opBridgeBatchSubmissionTarget    string

	generateKeys                     bool
	systemKeyOperatorMnemonic        string
	systemKeyBridgeExecutorMnemonic  string
	systemKeyOutputSubmitterMnemonic string
	systemKeyBatchSubmitterMnemonic  string
	systemKeyChallengerMnemonic      string

	systemKeyOperatorAddress        string
	systemKeyBridgeExecutorAddress  string
	systemKeyOutputSubmitterAddress string
	systemKeyBatchSubmitterAddress  string
	systemKeyChallengerAddress      string

	systemKeyL1OperatorBalance        string
	systemKeyL1BridgeExecutorBalance  string
	systemKeyL1OutputSubmitterBalance string
	systemKeyL1BatchSubmitterBalance  string
	systemKeyL1ChallengerBalance      string

	systemKeyL2OperatorBalance        string
	systemKeyL2BridgeExecutorBalance  string
	systemKeyL2OutputSubmitterBalance string
	systemKeyL2BatchSubmitterBalance  string
	systemKeyL2ChallengerBalance      string

	gasStationExist     bool
	downloadedNewBinary bool

	preGenesisAccountsResponsesCount int
	preL1BalancesResponsesCount      int
	preL2BalancesResponsesCount      int
}

func NewLaunchState() *LaunchState {
	return &LaunchState{
		weave: types.NewWeaveState(),
	}
}

func (ls *LaunchState) FinalizeGenesisAccounts() {
	ls.genesisAccounts = append(
		ls.genesisAccounts,
		GenesisAccount{
			Address: ls.systemKeyOperatorAddress,
			Coins:   ls.systemKeyL2OperatorBalance,
		},
		GenesisAccount{
			Address: ls.systemKeyBridgeExecutorAddress,
			Coins:   ls.systemKeyL2BridgeExecutorBalance,
		},
		GenesisAccount{
			Address: ls.systemKeyOutputSubmitterAddress,
			Coins:   ls.systemKeyL2OutputSubmitterBalance,
		},
		GenesisAccount{
			Address: ls.systemKeyBatchSubmitterAddress,
			Coins:   ls.systemKeyL2BatchSubmitterBalance,
		},
		GenesisAccount{
			Address: ls.systemKeyChallengerAddress,
			Coins:   ls.systemKeyL2ChallengerBalance,
		},
	)
}
