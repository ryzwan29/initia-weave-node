package registry

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/initia-labs/weave/client"
)

// LoadedChainRegistry contains a map of chain id to the chain.json
var LoadedChainRegistry = make(map[ChainType]*ChainRegistry)

// LoadedL2Registry contains a map of l2 network path to the chain.json
var LoadedL2Registry = make(map[string]*ChainRegistry)

type ChainRegistry struct {
	ChainId      string   `json:"chain_id"`
	Bech32Prefix string   `json:"bech32_prefix"`
	Fees         Fees     `json:"fees"`
	Codebase     Codebase `json:"codebase"`
	Apis         Apis     `json:"apis"`
	Peers        Peers    `json:"peers"`
}

type Fees struct {
	FeeTokens []FeeTokens `json:"fee_tokens"`
}

type FeeTokens struct {
	Denom            string  `json:"denom"`
	FixedMinGasPrice float64 `json:"fixed_min_gas_price"`
}

type Codebase struct {
	Genesis Genesis `json:"genesis"`
}

type Genesis struct {
	GenesisUrl string `json:"genesis_url"`
}

type Apis struct {
	Rpc  []Endpoint `json:"rpc"`
	Rest []Endpoint `json:"rest"`
	Grpc []Endpoint `json:"grpc"`
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

type L2GitHubContent struct {
	Name string `json:"name"`
	Path string `json:"path"`
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

func (cr *ChainRegistry) MustGetMinGasPriceByDenom(denom string) string {
	minGasPrice, err := cr.GetMinGasPriceByDenom(denom)
	if err != nil {
		panic(err)
	}

	return minGasPrice
}

func checkAndAddPort(addr string) (string, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("invalid address: %v", err)
	}

	if u.Port() == "" {
		if u.Scheme == "https" {
			u.Host = u.Host + ":443"
		} else if u.Scheme == "http" {
			u.Host = u.Host + ":80"
		}
	}

	return u.String(), nil
}

func (cr *ChainRegistry) GetActiveRpc() (string, error) {
	httpClient := client.NewHTTPClient()
	for _, rpc := range cr.Apis.Rpc {
		address, err := checkAndAddPort(rpc.Address)
		if err != nil {
			continue
		}

		_, err = httpClient.Get(address, "/health", nil, nil)
		if err != nil {
			continue
		}

		return address, nil
	}

	return "", fmt.Errorf("no active RPC endpoints available")
}

func (cr *ChainRegistry) MustGetActiveRpc() string {
	rpc, err := cr.GetActiveRpc()
	if err != nil {
		panic(err)
	}

	return rpc
}

func (cr *ChainRegistry) GetActiveLcd() (string, error) {
	httpClient := client.NewHTTPClient()
	for _, lcd := range cr.Apis.Rest {
		_, err := httpClient.Get(lcd.Address, "/cosmos/base/tendermint/v1beta1/syncing", nil, nil)
		if err != nil {
			continue
		}

		return lcd.Address, nil
	}

	return "", fmt.Errorf("no active LCD endpoints available")
}

func (cr *ChainRegistry) MustGetActiveLcd() string {
	lcd, err := cr.GetActiveLcd()
	if err != nil {
		panic(err)
	}

	return lcd
}

func (cr *ChainRegistry) GetActiveGrpc() (string, error) {
	grpcClient := client.NewGRPCClient()
	for _, grpc := range cr.Apis.Grpc {
		err := grpcClient.CheckHealth(grpc.Address)
		if err != nil {
			continue
		}

		return grpc.Address, nil
	}

	return "", fmt.Errorf("no active gRPC endpoints available")
}

func (cr *ChainRegistry) MustGetActiveGrpc() string {
	grpc, err := cr.GetActiveGrpc()
	if err != nil {
		panic(err)
	}

	return grpc
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

func (cr *ChainRegistry) GetGenesisUrl() string {
	return cr.Codebase.Genesis.GenesisUrl
}

func loadChainRegistry(chainType ChainType) error {
	httpClient := client.NewHTTPClient()
	endpoint := GetRegistryEndpoint(chainType)
	LoadedChainRegistry[chainType] = &ChainRegistry{}
	if _, err := httpClient.Get(endpoint, "", nil, LoadedChainRegistry[chainType]); err != nil {
		return err
	}

	return nil
}

func GetChainRegistry(chainType ChainType) (*ChainRegistry, error) {
	chainRegistry, ok := LoadedChainRegistry[chainType]
	if !ok {
		if err := loadChainRegistry(chainType); err != nil {
			return nil, fmt.Errorf("failed to load chain registry for %s: %v", chainType, err)
		}
		return LoadedChainRegistry[chainType], nil
	}

	return chainRegistry, nil
}

func MustGetChainRegistry(chainType ChainType) *ChainRegistry {
	chainRegistry, err := GetChainRegistry(chainType)
	if err != nil {
		panic(err)
	}

	return chainRegistry
}

func loadL2Registry(networkPath string) error {
	httpClient := client.NewHTTPClient()
	endpoint := fmt.Sprintf(InitiaRegistryEndpoint, networkPath)
	LoadedL2Registry[networkPath] = &ChainRegistry{}
	if _, err := httpClient.Get(endpoint, "", nil, LoadedL2Registry[networkPath]); err != nil {
		return err
	}

	return nil
}

func GetL2Registry(networkPath string) (*ChainRegistry, error) {
	l2Registry, ok := LoadedL2Registry[networkPath]
	if !ok {
		if err := loadL2Registry(networkPath); err != nil {
			return nil, fmt.Errorf("failed to load l2 registry for path %s: %v", networkPath, err)
		}
		return LoadedL2Registry[networkPath], nil
	}

	return l2Registry, nil
}

func MustGetL2Registry(networkPath string) *ChainRegistry {
	l2Registry, err := GetL2Registry(networkPath)
	if err != nil {
		panic(err)
	}

	return l2Registry
}

func GetAllL2Names(chainType ChainType) ([]L2GitHubContent, error) {
	if chainType == InitiaL1Testnet {
		client := client.NewHTTPClient()

		resp, err := client.Get("https://api.github.com", "/repos/initia-labs/initia-registry/contents/testnets", nil, nil)
		if err != nil {
			return []L2GitHubContent{}, fmt.Errorf("failed to fetch l2 registry: %v", err)
		}

		var l2 []L2GitHubContent
		// Decode the JSON
		if err := json.Unmarshal(resp, &l2); err != nil {
			return []L2GitHubContent{}, err
		}

		return l2, nil
	}

	return []L2GitHubContent{}, fmt.Errorf("failed to matched chain type")
}

func MustGetAllL2Contents(chainType ChainType) []L2GitHubContent {
	contents, err := GetAllL2Names(chainType)
	if err != nil {
		panic(err)
	}

	return contents
}

var OPInitBotsSpecVersion map[string]int

func loadOPInitBotsSpecVersion() error {
	httpClient := client.NewHTTPClient()
	if _, err := httpClient.Get(OPInitBotsSpecEndpoint, "", nil, &OPInitBotsSpecVersion); err != nil {
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

func MustGetOPInitBotsSpecVersion(chainId string) int {
	version, err := GetOPInitBotsSpecVersion(chainId)
	if err != nil {
		panic(err)
	}

	return version
}
