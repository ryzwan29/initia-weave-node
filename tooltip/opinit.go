package tooltip

import "github.com/initia-labs/weave/ui"

var (
	ListenAddressTooltip          = ui.NewTooltip("Listen address", "The network address and port where the bot listens for incoming queries regarding deposits, withdrawals, and challenges.", "", []string{}, []string{}, []string{})
	InitiaDALayerTooltip          = ui.NewTooltip("Initia", "Ideal for projects that require close integration within the Initia network, offering streamlined communication and data handling within the Initia ecosystem.", "", []string{}, []string{}, []string{})
	CelestiaMainnetDALayerTooltip = ui.NewTooltip("Celestia Mainnet", "Suitable for production environments that need reliable and secure data availability with Celestia's decentralized architecture, ensuring robust support for live applications.", "", []string{}, []string{}, []string{})
	CelestiaTestnetDALayerTooltip = ui.NewTooltip("Celestia Testnet", "Best for testing purposes, allowing you to validate functionality and performance in a non-production setting before deploying to a mainnet environment.", "", []string{}, []string{}, []string{})
)
