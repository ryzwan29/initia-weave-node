package registry

const (
	CelestiaTestnet ChainType = iota
	CelestiaMainnet
	InitiaL1Testnet
)

const (
	InitiaRegistryEndpoint   string = "https://raw.githubusercontent.com/initia-labs/initia-registry/refs/heads/main/%s/chain.json"
	CelestiaRegistryEndpoint string = "https://raw.githubusercontent.com/cosmos/chain-registry/refs/heads/master/%s/chain.json"
	OPInitBotsSpecEndpoint   string = "https://raw.githubusercontent.com/initia-labs/opinit-bots/refs/heads/main/spec_version.json"
)

var (
	ChainTypeToEndpoint = map[ChainType]string{
		CelestiaTestnet: CelestiaRegistryEndpoint,
		CelestiaMainnet: CelestiaRegistryEndpoint,
		InitiaL1Testnet: InitiaRegistryEndpoint,
	}
	ChainTypeToEndpointSlug = map[ChainType]string{
		CelestiaTestnet: "testnets/celestiatestnet3",
		CelestiaMainnet: "celestia",
		InitiaL1Testnet: "testnets/initia",
	}
)
