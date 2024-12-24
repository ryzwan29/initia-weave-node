package cosmosutils

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/initia-labs/weave/client"
)

const (
	DefaultGasAdjustment = "1.4"
	TmpKeyName           = "weave.clitx.executor"
)

type InitiadTxResponse struct {
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

type InitiadTxExecutor struct {
	binaryPath string
}

func NewInitiadTxExecutor(rest string) (*InitiadTxExecutor, error) {
	httpClient := client.NewHTTPClient()
	nodeVersion, url, err := GetInitiaBinaryUrlFromLcd(httpClient, rest)
	if err != nil {
		return nil, err
	}
	binaryPath, err := GetInitiaBinaryPath(nodeVersion)
	if err != nil {
		return nil, err
	}
	err = InstallInitiaBinary(nodeVersion, url, binaryPath)
	if err != nil {
		return nil, err
	}

	return &InitiadTxExecutor{
		binaryPath: binaryPath,
	}, nil
}

func (te *InitiadTxExecutor) BroadcastMsgSend(senderMnemonic, recipientAddress, amount, gasPrices, rpc, chainId string) (*InitiadTxResponse, error) {
	_, err := RecoverKeyFromMnemonic(te.binaryPath, TmpKeyName, senderMnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to recover gas station key: %v", err)
	}
	defer func() {
		_ = DeleteKey(te.binaryPath, TmpKeyName)
	}()

	cmd := exec.Command(te.binaryPath, "tx", "bank", "send", TmpKeyName, recipientAddress, amount, "--from",
		TmpKeyName, "--chain-id", chainId, "--gas", "auto", "--gas-adjustment", DefaultGasAdjustment,
		"--gas-prices", gasPrices, "--node", rpc, "--output", "json", "--keyring-backend", "test", "-y")

	outputBytes, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to send tx MsgSend for %s: %v, output: %s", TmpKeyName, err, string(outputBytes))
	}

	var txResponse InitiadTxResponse
	err = json.Unmarshal(outputBytes, &txResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if txResponse.Code != 0 {
		return nil, fmt.Errorf("tx failed with error: %v", txResponse.RawLog)
	}

	return &txResponse, nil
}

type HermesTxExecutor struct {
	binaryPath string
}

func NewHermesTxExecutor(binaryPath string) *HermesTxExecutor {
	return &HermesTxExecutor{
		binaryPath: binaryPath,
	}
}

func (te *HermesTxExecutor) UpdateClient(clientId, chainId string) (string, error) {
	cmd := exec.Command(te.binaryPath, "update", "client", "--host-chain", chainId, "--client", clientId)

	outputBytes, err := cmd.Output()
	if err != nil {
		return string(outputBytes), fmt.Errorf("failed to update client: %v, output: %s", err, string(outputBytes))
	}

	return string(outputBytes), err
}
