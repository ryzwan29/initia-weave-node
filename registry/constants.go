package registry

import "fmt"

const (
	CelestiaTestnet ChainType = iota
	CelestiaMainnet
	InitiaL1Testnet
	InitiaL1Mainnet
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
		InitiaL1Mainnet: InitiaRegistryEndpoint,
	}
	ChainTypeToEndpointSlug = map[ChainType]string{
		CelestiaTestnet: "testnets/celestiatestnet3",
		CelestiaMainnet: "celestia",
		InitiaL1Testnet: "testnets/initia",
		InitiaL1Mainnet: "mainnets/initia",
	}
)

func GetRegistryEndpoint(chainType ChainType) string {
	return fmt.Sprintf(ChainTypeToEndpoint[chainType], ChainTypeToEndpointSlug[chainType])
}
