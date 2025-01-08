package registry

import (
	"fmt"
)

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

	InitiaTestnetRegistryAPI string = "https://registry.testnet.initia.xyz/chains.json"
	InitiaMainnetRegistryAPI string = "https://registry.initia.xyz/chains.json"
	InitiaL1PrettyName       string = "Initia"

	InitiaTestnetGraphQLAPI string = "https://graphql.testnet.initia.xyz/v1/graphql"
	InitiaMainnetGraphQLAPI string = "https://graphql.initia.xyz/v1/graphql"
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
	ChainTypeToInitiaRegistryAPI = map[ChainType]string{
		InitiaL1Testnet: InitiaTestnetRegistryAPI,
		InitiaL1Mainnet: InitiaMainnetRegistryAPI,
	}
	ChainTypeToInitiaGraphQLAPI = map[ChainType]string{
		InitiaL1Testnet: InitiaTestnetGraphQLAPI,
		InitiaL1Mainnet: InitiaMainnetGraphQLAPI,
	}
)

func GetRegistryEndpoint(chainType ChainType) string {
	return fmt.Sprintf(ChainTypeToEndpoint[chainType], ChainTypeToEndpointSlug[chainType])
}
