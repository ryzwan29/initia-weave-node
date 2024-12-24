package cosmosutils

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/initia-labs/weave/client"
)

type InitiadBankBalancesQueryResponse struct {
	Balances Coins `json:"balances"`
}

type InitiadQuerier struct {
	binaryPath string
}

func NewInitiadQuerier(rest string) (*InitiadQuerier, error) {
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

	return &InitiadQuerier{
		binaryPath: binaryPath,
	}, nil
}

func (iq *InitiadQuerier) QueryBankBalances(address, rpc string) (*Coins, error) {
	cmd := exec.Command(iq.binaryPath, "query", "bank", "balances", address, "--node", rpc, "--output", "json")

	outputBytes, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query bank balances for %s: %v, output: %s", address, err, string(outputBytes))
	}

	var queryResponse InitiadBankBalancesQueryResponse
	err = json.Unmarshal(outputBytes, &queryResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &queryResponse.Balances, nil
}
