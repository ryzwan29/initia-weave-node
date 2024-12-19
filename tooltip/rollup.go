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
	OpBridgeOutputFinalizationPeriodTooltip = ui.NewTooltip("Output Finalization Period", "The maximum waiting time before submitting the rollup output root to L1 again.", "", []string{}, []string{}, []string{})
	OpBridgeBatchSubmissionTargetTooltip    = ui.NewTooltip("Batch Submission Target", "The target chain for submitting rollup blocks and transaction data to ensure Data Availability. Currently, submissions can be made to Initia L1 or Celestia.\n\nFor production use, we recommend Celestia due to its cost-effective block space and faster query capabilities.", "", []string{}, []string{}, []string{})
	EnableOracleTooltip                     = ui.NewTooltip("Oracle", "Oracle fetches and submits real-world price data to the blockchain, with validators running both an on-chain component and a sidecar process to gather and relay prices. \nEnabling this is recommended.", "", []string{}, []string{}, []string{})

	// System Key Tooltips
	SystemKeyOperatorMnemonicTooltip = ui.NewTooltip("Rollup Operator", "Also known as Sequencer, is responsible for creating blocks, ordering and including transactions within each block, and maintaining the operation of the rollup network.", "", []string{}, []string{}, []string{})
)
