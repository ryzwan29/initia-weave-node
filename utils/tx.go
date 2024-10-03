package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

const DefaultGasAdjustment = "1.4"

type CliTxResponse struct {
	Height    string         `json:"height"`
	TxHash    string         `json:"txhash"`
	Codespace string         `json:"codespace"`
	Code      int            `json:"code"`
	Data      string         `json:"data"`
	RawLog    string         `json:"raw_log"`
	Logs      *[]interface{} `json:"logs"`
	Info      string         `json:"info"`
	GasWanted string         `json:"gas_wanted"`
	GasUsed   string         `json:"gas_used"`
	Tx        *[]interface{} `json:"tx"`
	Timestamp string         `json:"timestamp"`
	Events    *[]interface{} `json:"events"`
}

type TxExecutor struct {
	appName       string
	chainId       string
	senderKeyName string
	gasPrices     string
	rpc           string
}

func NewTxExecutor(appName, chainId, senderKeyName, gasPrices, rpc string) *TxExecutor {
	return &TxExecutor{
		appName:       appName,
		chainId:       chainId,
		senderKeyName: senderKeyName,
		gasPrices:     gasPrices,
		rpc:           gasPrices,
	}
}

func (te *TxExecutor) BroadcastMsgSend(recipientAddress, amount string) (*CliTxResponse, error) {
	cmd := exec.Command(te.appName, "tx", "bank", "send", te.senderKeyName, recipientAddress, amount, "--from",
		te.senderKeyName, "--chain-id", te.chainId, "--gas", "auto", "--gas-adjustment", DefaultGasAdjustment,
		"--gas-prices", te.gasPrices, "--node", te.rpc, "--output", "json", "--keyring-backend", "test", "-y")

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to send tx MsgSend for %s: %v, output: %s", te.senderKeyName, err, string(outputBytes))
	}

	var txResponse CliTxResponse
	err = json.Unmarshal(outputBytes, &txResponse)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal JSON: %v", err))
	}

	return &txResponse, nil
}
