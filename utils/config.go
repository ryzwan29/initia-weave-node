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

const DefaultConfigTemplate = `
{
	"endpoints": {
		"mainnet": {
			"rpc": "https://rpc.initia.xyz",
			"lcd": "https://lcd.initia.xyz"
		},
		"testnet": {
			"rpc": "https://rpc.testnet.initia.xyz",
			"lcd": "https://lcd.testnet.initia.xyz"
		}
	}
}
`
