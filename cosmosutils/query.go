package cosmosutils

import (
	"encoding/json"
	"fmt"

	"github.com/initia-labs/weave/client"
)

const (
	NoBalancesText string = "No Balances"
)

func QueryBankBalances(rest, address string) (*Coins, error) {
	httpClient := client.NewHTTPClient()
	var result map[string]interface{}
	_, err := httpClient.Get(
		rest,
		fmt.Sprintf("/cosmos/bank/v1beta1/balances/%s", address),
		map[string]string{"pagination.limit": "100"},
		&result,
	)
	if err != nil {
		return nil, err
	}

	rawBalances, ok := result["balances"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse balances field")
	}

	balancesJSON, err := json.Marshal(rawBalances)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal balances: %w", err)
	}

	var balances Coins
	err = json.Unmarshal(balancesJSON, &balances)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal balances into Coins: %w", err)
	}

	return &balances, nil
}

func QueryOPChildParams(address string) (params OPChildParams, err error) {
	httpClient := client.NewHTTPClient()

	var res OPChildParamsResoponse
	if _, err := httpClient.Get(address, "/opinit/opchild/v1/params", nil, &res); err != nil {
		return params, err
	}

	return res.Params, nil
}
