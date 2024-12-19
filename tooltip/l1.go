package tooltip

import (
	"github.com/initia-labs/weave/ui"
)

var (
	L1ChainIdTooltip     = ui.NewTooltip("L1 chain ID", ChainIDDescription("L1"), "", []string{}, []string{}, []string{})
	L1RPCEndpointTooltip = ui.NewTooltip("L1 RPC endpoint", RPCEndpointDescription("L1"), "", []string{}, []string{}, []string{})
)
