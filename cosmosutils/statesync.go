package cosmosutils

import (
	"fmt"
	"strconv"

	"github.com/initia-labs/weave/client"
)

type BlockResponse struct {
	Result struct {
		Block struct {
			Header struct {
				Height string `json:"height"`
			} `json:"header"`
		} `json:"block"`
	} `json:"result"`
}

type HashResponse struct {
	Result struct {
		BlockID struct {
			Hash string `json:"hash"`
		} `json:"block_id"`
	} `json:"result"`
}

type StateSyncInfo struct {
	TrustHeight int
	TrustHash   string
}

func GetStateSyncInfo(url string) (*StateSyncInfo, error) {
	httpClient := client.NewHTTPClient()
	var latestBlock BlockResponse
	_, err := httpClient.Get(url, "/block", nil, &latestBlock)
	if err != nil {
		return nil, fmt.Errorf("Error fetching latest block height: %v\n", err)
	}

	latestHeight, err := strconv.Atoi(latestBlock.Result.Block.Header.Height)
	if err != nil {
		return nil, fmt.Errorf("Error converting block height to integer: %v\n", err)
	}
	blockHeight := latestHeight - 2000

	var trustHashResp HashResponse
	_, err = httpClient.Get(url, "/block", map[string]string{"height": strconv.Itoa(blockHeight)}, &trustHashResp)
	if err != nil {
		return nil, fmt.Errorf("Error fetching trust hash: %v\n", err)
	}

	return &StateSyncInfo{
		TrustHeight: blockHeight,
		TrustHash:   trustHashResp.Result.BlockID.Hash,
	}, nil
}
