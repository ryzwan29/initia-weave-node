package tooltip

import (
	"github.com/initia-labs/weave/ui"
)

var (
	RollupChainIdTooltip           = ui.NewTooltip("Rollup chain ID", ChainIDDescription("rollup"), "", []string{}, []string{}, []string{})
	RollupRPCEndpointTooltip       = ui.NewTooltip("Rollup RPC endpoint", RPCEndpointDescription("rollup"), "", []string{}, []string{}, []string{})
	RollupGRPCEndpointTooltip      = ui.NewTooltip("Rollup GRPC endpoint", GRPCEndpointDescription("rollup"), "", []string{}, []string{}, []string{})
	RollupWebSocketEndpointTooltip = ui.NewTooltip("Rollup WebSocket endpoint", WebSocketEndpointDescription("rollup"), "", []string{}, []string{}, []string{})
	RollupGasDenomTooltip          = ui.NewTooltip("Rollup gas denom", GasDenomDescription("rollup"), "", []string{}, []string{}, []string{})
	RollupGasPriceTooltip          = ui.NewTooltip("Rollup gas price", GasPriceDescription("rollup"), "", []string{}, []string{}, []string{})

	// OP Bridge Tooltips
	OpBridgeSubmissionIntervalTooltip       = ui.NewTooltip("Submission Interval", "The maximum waiting time before submitting the rollup output root to L1 again.", "", []string{}, []string{}, []string{})
	OpBridgeOutputFinalizationPeriodTooltip = ui.NewTooltip("Output Finalization Period", "The time period during which submitted output roots can be challenged before being considered final. After this period, the output becomes immutable.", "", []string{}, []string{}, []string{})
	OpBridgeBatchSubmissionTargetTooltip    = ui.NewTooltip("Batch Submission Target", "The target chain for submitting rollup blocks and transaction data to ensure Data Availability. Currently, submissions can be made to Initia L1 or Celestia.\n\nFor production use, we recommend Celestia due to its cost-effective block space and faster query capabilities.", "", []string{}, []string{}, []string{})
	EnableOracleTooltip                     = ui.NewTooltip("Oracle", "Oracle fetches and submits real-world price data to the blockchain, with validators running both an on-chain component and a sidecar process to gather and relay prices. \nEnabling this is recommended.", "", []string{}, []string{}, []string{})

	// System Key Tooltips
	SystemKeyOperatorMnemonicTooltip        = ui.NewTooltip("Rollup Operator", "Also known as Sequencer, is responsible for creating blocks, ordering and including transactions within each block, and maintaining the operation of the rollup network.", "", []string{}, []string{}, []string{})
	SystemKeyBridgeExecutorMnemonicTooltip  = ui.NewTooltip("Bridge Executor", "Monitors the L1 and rollup transactions, facilitates token bridging and withdrawals between the rollup and Initia L1 chain, and also relays oracle price feed to rollup.", "", []string{}, []string{}, []string{})
	SystemKeyOutputSubmitterMnemonicTooltip = ui.NewTooltip("Output Submitter", "Submits rollup output roots to L1 for verification and potential challenges. If the submitted output remains unchallenged beyond the output finalization period, it is considered finalized and immutable.", "", []string{}, []string{}, []string{})
	SystemKeyBatchSubmitterMnemonicTooltip  = ui.NewTooltip("Batch Submitter", "Submits block and transactions data in batches into a chain to ensure Data Availability. Currently, submissions can be made to Initia L1 or Celestia.", "", []string{}, []string{}, []string{})
	SystemKeyChallengerMnemonicTooltip      = ui.NewTooltip("Challenger", "Prevents misconduct and invalid rollup state submissions by monitoring for output proposals and challenging any that are invalid.", "", []string{}, []string{}, []string{})

	// System Accounts funding
	SystemAccountsFundingPresetTooltip = ui.NewTooltip(
		"OPInit Bots",
		"Bridge Executor: Monitors the L1 and rollup transactions, facilitates token bridging and withdrawals between the rollup and Initia L1 chain, and also relays oracle price feed to rollup.\n\nOutput Submitter: Submits rollup output roots to L1 for verification and potential challenges. If the submitted output remains unchallenged beyond the output finalization period, it is considered finalized and immutable.\n\nBatch Submitter: Submits block and transactions data in batches into a chain to ensure Data Availability. Currently, submissions can be made to Initia L1 or Celestia.\n\nChallenger: Prevents misconduct and invalid rollup state submissions by monitoring for output proposals and challenging any that are invalid.\n\nRollup Operator: Also known as Sequencer, is responsible for creating blocks, ordering and including transactions within each block, and maintaining the operation of the rollup network.",
		"", []string{"Bridge Executor:", "Output Submitter:", "Batch Submitter:", "Challenger:", "Rollup Operator:"}, []string{}, []string{},
	)
	GasStationInRollupGenesisTooltip        = ui.NewTooltip("Gas station in rollup genesis", "Adding gas station account as a genesis account grants initial balances to the gas station on the rollup at network launch. This ensures the gas station can fund relayer accounts during the Weave relayer flow, eliminating the need for later manual funding.", "", []string{}, []string{}, []string{})
	GasStationBalanceOnRollupGenesisTooltip = ui.NewTooltip("Gas station genesis balance", "A genesis balance is the amount of tokens allocated to specific accounts when a blockchain network launches, in this case the gas station account. It allows these accounts to have immediate resources for transactions, testing, or operational roles without needing to acquire tokens afterward.", "", []string{}, []string{}, []string{})

	// Genesis Accounts Tooltips
	GenesisAccountSelectTooltip = ui.NewTooltip("Genesis account", "Genesis accounts are accounts that are created at the genesis block of a blockchain network. They are used to grant initial balances to specific accounts at network launch, enabling early operations without later funding.", "", []string{}, []string{}, []string{})
	GenesisBalanceInputTooltip  = ui.NewTooltip("Genesis balance", "The amount of tokens allocated to specific accounts when a blockchain network launches, in this case the genesis account. It allows these accounts to have immediate resources for transactions, testing, or operational roles without needing to acquire tokens afterward.", "", []string{}, []string{}, []string{})
)
