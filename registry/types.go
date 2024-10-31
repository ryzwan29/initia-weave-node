package registry

type ChainType int

func (ct ChainType) String() string {
	switch ct {
	case CelestiaTestnet:
		return "Celestia Testnet"
	case CelestiaMainnet:
		return "Celestia Mainnet"
	case InitiaL1Testnet:
		return "Initia L1 Testnet"
	case InitiaL1Mainnet:
		return "Initia L1 Mainnet"
	default:
		return "Unknown"
	}
}
