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
	genesisAccounts    types.GenesisAccounts

	opBridgeSubmissionInterval       string
	opBridgeOutputFinalizationPeriod string
	opBridgeBatchSubmissionTarget    string
	batchSubmissionIsCelestia        bool

	generateKeys                     bool
	systemKeyOperatorMnemonic        string
	systemKeyBridgeExecutorMnemonic  string
	systemKeyOutputSubmitterMnemonic string
	systemKeyBatchSubmitterMnemonic  string
	systemKeyChallengerMnemonic      string

	systemKeyOperatorAddress         string
	systemKeyBridgeExecutorAddress   string
	systemKeyOutputSubmitterAddress  string
	systemKeyBatchSubmitterAddress   string
	systemKeyL2BatchSubmitterAddress string
	systemKeyChallengerAddress       string

	systemKeyL1OperatorBalance        string
	systemKeyL1BridgeExecutorBalance  string
	systemKeyL1OutputSubmitterBalance string
	systemKeyL1BatchSubmitterBalance  string
	systemKeyL1ChallengerBalance      string
	systemKeyL1FundingTxHash          string
	systemKeyCelestiaFundingTxHash    string

	systemKeyL2OperatorBalance        string
	systemKeyL2BridgeExecutorBalance  string
	systemKeyL2OutputSubmitterBalance string
	systemKeyL2BatchSubmitterBalance  string
	systemKeyL2ChallengerBalance      string

	gasStationExist             bool
	downloadedNewBinary         bool
	downloadedNewCelestiaBinary bool

	preGenesisAccountsResponsesCount int
	preL1BalancesResponsesCount      int
	preL2BalancesResponsesCount      int

	minitiadLaunchStreamingLogs []string

	binaryPath         string
	celestiaBinaryPath string

	launchFromExistingConfig bool
	existingConfigPath       string
}

func NewLaunchState() *LaunchState {
	return &LaunchState{
		weave: types.NewWeaveState(),
	}
}

func (ls *LaunchState) FinalizeGenesisAccounts() {
	accounts := []types.GenesisAccount{
		{Address: ls.systemKeyOperatorAddress, Coins: ls.systemKeyL2OperatorBalance},
		{Address: ls.systemKeyBridgeExecutorAddress, Coins: ls.systemKeyL2BridgeExecutorBalance},
		{Address: ls.systemKeyOutputSubmitterAddress, Coins: ls.systemKeyL2OutputSubmitterBalance},
		{Address: ls.systemKeyChallengerAddress, Coins: ls.systemKeyL2ChallengerBalance},
	}

	if !ls.batchSubmissionIsCelestia {
		accounts = append(accounts, types.GenesisAccount{
			Address: ls.systemKeyBatchSubmitterAddress,
			Coins:   ls.systemKeyL2BatchSubmitterBalance,
		})
	}

	ls.genesisAccounts = append(ls.genesisAccounts, accounts...)
}

func (ls *LaunchState) PrepareLaunchingWithConfig(vm, minitiadVersion, minitiadEndpoint, configPath string) {
	vmType, err := ParseVMType(vm)
	if err != nil {
		panic(err)
	}
	ls.vmType = string(vmType)
	ls.minitiadVersion = minitiadVersion
	ls.minitiadEndpoint = minitiadEndpoint
	ls.launchFromExistingConfig = true
	ls.existingConfigPath = configPath
}
