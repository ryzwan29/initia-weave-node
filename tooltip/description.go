package tooltip

import "fmt"

func ChainIDDescription(networkType string) string {
	return fmt.Sprintf("Identifier for the %s network.", networkType)
}

func RPCEndpointDescription(networkType string) string {
	return fmt.Sprintf("The network address and port that an %s RPC node listens. This allows the bots to communicate with %s network.", networkType, networkType)
}

func GRPCEndpointDescription(networkType string) string {
	return fmt.Sprintf("The network address and port that an %s GRPC node listens. This allows the bots to query additional data from the %s network.", networkType, networkType)
}

func WebSocketEndpointDescription(networkType string) string {
	return fmt.Sprintf("The network address and port that an %s WebSocket node listens. This allows the bots to listen to events from the %s network.", networkType, networkType)
}

func GasDenomDescription(networkType string) string {
	return fmt.Sprintf("The gas denom to be used for submitting transactions to the %s node.", networkType)
}

func GasPriceDescription(networkType string) string {
	return fmt.Sprintf("The gas price to be used for submitting transactions to the %s node. This value should be set to the minimum gas price for the %s node.", networkType, networkType)
}

func MinGasPriceDescription(networkType string) string {
	return fmt.Sprintf("The minimum gas price for transactions submitted on the %s network. This value helps ensure that %s transactions are processed with adequate priority.", networkType, networkType)
}
