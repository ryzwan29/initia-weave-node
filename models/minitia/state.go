package minitia

import "github.com/initia-labs/weave/types"

type GenesisAccount struct {
	address string
	balance string
}

type LaunchState struct {
	weave              types.WeaveState
	existingMinitiaApp bool
	l1Network          string
	vmType             string
	minitiadVersion    string
	minitiadEndpoint   string
	chainId            string
	gasDenom           string
	moniker            string
	genesisAccounts    []GenesisAccount

	opBridgeSubmissionInterval       string
	opBridgeOutputFinalizationPeriod string
	opBridgeBatchSubmissionTarget    string

	systemKeyOperatorMnemonic        string
	systemKeyBridgeExecutorMnemonic  string
	systemKeyOutputSubmitterMnemonic string
	systemKeyBatchSubmitterMnemonic  string
	systemKeyChallengerMnemonic      string

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

	gasStationExist bool

	preGenesisAccountsResponsesCount int
	preL1BalancesResponsesCount      int
	preL2BalancesResponsesCount      int
}

func NewLaunchState() *LaunchState {
	return &LaunchState{
		weave: types.NewWeaveState(),
	}
}
