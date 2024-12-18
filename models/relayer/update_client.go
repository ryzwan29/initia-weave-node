package relayer

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/initia-labs/weave/common"
	"github.com/initia-labs/weave/cosmosutils"
	"github.com/initia-labs/weave/registry"
)

func UpdateClientFromConfig() error {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(userHome, HermesHome, "config.toml")
	weaveDataPath := filepath.Join(userHome, common.WeaveDataDirectory)
	hermesBinaryPath := filepath.Join(weaveDataPath, "hermes")

	tomlData, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config Config
	err = toml.Unmarshal([]byte(tomlData), &config)
	if err != nil {
		return err
	}
	var chainRegistry *registry.ChainRegistry

	if config.Chains[0].ID == registry.MustGetChainRegistry(registry.InitiaL1Testnet).GetChainId() {
		chainRegistry = registry.MustGetChainRegistry(registry.InitiaL1Testnet)
	} else if config.Chains[0].ID == registry.MustGetChainRegistry(registry.InitiaL1Mainnet).GetChainId() {
		chainRegistry = registry.MustGetChainRegistry(registry.InitiaL1Mainnet)
	}

	clientIds := make(map[string]bool)
	for _, channel := range config.Chains[0].PacketFilter.List {
		connection := chainRegistry.MustGetCounterpartyClientId(channel[0], channel[1])
		clientIds[connection.Connection.Counterparty.ClientID] = true
	}
	te := cosmosutils.NewHermesTxExecutor(hermesBinaryPath)

	for clientId := range clientIds {
		_, err := te.UpdateClient(clientId, config.Chains[1].ID)
		if err != nil {
			return err
		}
	}
	return nil
}
