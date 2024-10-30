package registry

import (
	"fmt"
	"strings"

	"github.com/initia-labs/weave/utils"
)

// LoadedChainRegistry contains a map of chain id to the chain.json
var LoadedChainRegistry = make(map[ChainType]*ChainRegistry)

type ChainRegistry struct {
	ChainId      string `json:"chain_id"`
	Bech32Prefix string `json:"bech32_prefix"`
	Fees         Fees   `json:"fees"`
	Apis         Apis   `json:"apis"`
	Peers        Peers  `json:"peers"`
}

type Fees struct {
	FeeTokens []FeeTokens `json:"fee_tokens"`
}

type FeeTokens struct {
	Denom            string  `json:"denom"`
	FixedMinGasPrice float64 `json:"fixed_min_gas_price"`
}

type Apis struct {
	Rpc  []Endpoint `json:"rpc"`
	Rest []Endpoint `json:"rest"`
}

type Endpoint struct {
	Address        string `json:"address"`
	Provider       string `json:"provider"`
	AuthorizedUser string `json:"authorizedUser,omitempty"`
	IndexForSkip   int    `json:"indexForSkip,omitempty"`
}

type Peers struct {
	Seeds           []Peer `json:"seeds,omitempty"`
	PersistentPeers []Peer `json:"persistent_peers,omitempty"`
}

type Peer struct {
	Id       string `json:"id"`
	Address  string `json:"address"`
	Provider string `json:"provider,omitempty"`
}

func (cr *ChainRegistry) GetChainId() string {
	return cr.ChainId
}

func (cr *ChainRegistry) GetBech32Prefix() string {
	return cr.Bech32Prefix
}

func (cr *ChainRegistry) GetMinGasPriceByDenom(denom string) (string, error) {
	for _, feeToken := range cr.Fees.FeeTokens {
		if feeToken.Denom == denom {
			return fmt.Sprintf("%g%s", feeToken.FixedMinGasPrice, denom), nil
		}
	}
	return "", fmt.Errorf("denomination %s not found in fee tokens", denom)
}

func (cr *ChainRegistry) GetActiveRpc() (string, error) {
	client := utils.NewHTTPClient()
	for _, rpc := range cr.Apis.Rpc {
		_, err := client.Get(rpc.Address, "/health", nil, nil)
		if err != nil {
			continue
		}

		return rpc.Address, nil
	}

	return "", fmt.Errorf("no active RPC endpoints available")
}

func (cr *ChainRegistry) GetActiveLcd() (string, error) {
	client := utils.NewHTTPClient()
	for _, lcd := range cr.Apis.Rest {
		_, err := client.Get(lcd.Address, "/cosmos/base/tendermint/v1beta1/syncing", nil, nil)
		if err != nil {
			continue
		}

		return lcd.Address, nil
	}

	return "", fmt.Errorf("no active LCD endpoints available")
}

func (cr *ChainRegistry) GetSeeds() string {
	var seeds []string
	for _, seed := range cr.Peers.Seeds {
		seeds = append(seeds, fmt.Sprintf("%s@%s", seed.Id, seed.Address))
	}
	return strings.Join(seeds, ",")
}

func (cr *ChainRegistry) GetPersistentPeers() string {
	var persistentPeers []string
	for _, seed := range cr.Peers.PersistentPeers {
		persistentPeers = append(persistentPeers, fmt.Sprintf("%s@%s", seed.Id, seed.Address))
	}
	return strings.Join(persistentPeers, ",")
}

func loadChainRegistry(chainType ChainType) error {
	client := utils.NewHTTPClient()
	endpoint := GetRegistryEndpoint(chainType)
	LoadedChainRegistry[chainType] = &ChainRegistry{}
	if _, err := client.Get(endpoint, "", nil, LoadedChainRegistry[chainType]); err != nil {
		return err
	}

	return nil
}

func GetChainRegistry(chainType ChainType) (*ChainRegistry, error) {
	if _, ok := LoadedChainRegistry[chainType]; !ok {
		if err := loadChainRegistry(chainType); err != nil {
			return nil, fmt.Errorf("failed to load chain registry for %s: %v", chainType, err)
		}
	}

	chainRegistry, ok := LoadedChainRegistry[chainType]
	if !ok {
		return nil, fmt.Errorf("cannot retrieve chain registry from map")
	}

	return chainRegistry, nil
}

var OPInitBotsSpecVersion = make(map[string]int)

func loadOPInitBotsSpecVersion() error {
	client := utils.NewHTTPClient()
	if _, err := client.Get(OPInitBotsSpecEndpoint, "", nil, &OPInitBotsSpecVersion); err != nil {
		return fmt.Errorf("failed to load opinit-bots spec_version: %v", err)
	}
	return nil
}

func GetOPInitBotsSpecVersion(chainId string) (int, error) {
	if OPInitBotsSpecVersion == nil {
		if err := loadOPInitBotsSpecVersion(); err != nil {
			return 0, err
		}
	}

	version, ok := OPInitBotsSpecVersion[chainId]
	if !ok {
		return 0, fmt.Errorf("chain id not found in the spec_version")
	}

	return version, nil
}
