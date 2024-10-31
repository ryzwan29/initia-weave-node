package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func InitializeConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, WeaveDirectory, "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	dataPath := filepath.Join(homeDir, WeaveDataDirectory)
	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	logPath := filepath.Join(homeDir, WeaveLogDirectory)
	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfigFile(configPath); err != nil {
			return fmt.Errorf("failed to create default config file: %v", err)
		}
	}

	return LoadConfig()
}

func createDefaultConfigFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(DefaultConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to write to config file: %v", err)
	}

	return nil
}

func LoadConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	return nil
}

func GetConfig(key string) interface{} {
	return viper.Get(key)
}

func SetConfig(key string, value interface{}) error {
	viper.Set(key, value)
	return WriteConfig()
}

func WriteConfig() error {
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	return nil
}

func IsFirstTimeSetup() bool {
	return viper.Get("common.gas_station_mnemonic") == nil
}

func GetGasStationMnemonic() string {
	return GetConfig("common.gas_station_mnemonic").(string)
}

const DefaultConfigTemplate = `
{
  "constants": {
    "chain_id": {
      "mainnet": "initia-1",
      "testnet": "initiation-2"
    },
    "endpoints": {
      "mainnet": {
        "rpc": "https://rpc.initia.xyz:443",
        "lcd": "https://lcd.initia.xyz",
        "genesis": "https://initia.s3.ap-southeast-1.amazonaws.com/initia-1/genesis.json"
      },
      "testnet": {
        "rpc": "https://rpc.initiation-2.initia.xyz:443",
        "lcd": "https://lcd.initiation-2.initia.xyz",
        "genesis": "https://storage.googleapis.com/initia-binaries/genesis.json"
      }
    },
    "da_layer": {
      "celestia_testnet": {
        "chain_id": "mocha-4",
        "rpc": "https://celestia-testnet-rpc.publicnode.com:443",
		"lcd": "https://celestia-mocha-rest.publicnode.com",
        "bech32_prefix": "celestia",
        "gas_price": "0.02utia"
      },
      "celestia_mainnet": {
        "chain_id": "celestia",
        "rpc": "https://celestia-rpc.mesa.newmetric.xyz",
		"lcd": "https://celestia-rest.publicnode.com",
        "bech32_prefix": "celestia",
        "gas_price": "0.02utia"
      }
    }
  }
}
`
