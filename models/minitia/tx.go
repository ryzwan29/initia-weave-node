package minitia

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/initia-labs/weave/common"
	"github.com/initia-labs/weave/config"
	"github.com/initia-labs/weave/cosmosutils"
	"github.com/initia-labs/weave/io"
	"github.com/initia-labs/weave/registry"
	"github.com/initia-labs/weave/types"
)

type L1SystemKeys struct {
	BridgeExecutor  *types.GenesisAccount
	OutputSubmitter *types.GenesisAccount
	BatchSubmitter  *types.GenesisAccount
	Challenger      *types.GenesisAccount
}

func NewL1SystemKeys(bridgeExecutor, outputSubmitter, batchSubmitter, challenger *types.GenesisAccount) *L1SystemKeys {
	return &L1SystemKeys{
		BridgeExecutor:  bridgeExecutor,
		OutputSubmitter: outputSubmitter,
		BatchSubmitter:  batchSubmitter,
		Challenger:      challenger,
	}
}

type FundAccountsResponse struct {
	CelestiaTx *cosmosutils.InitiadTxResponse
	InitiaTx   *cosmosutils.InitiadTxResponse
}

func (lsk *L1SystemKeys) FundAccountsWithGasStation(state *LaunchState) (*FundAccountsResponse, error) {
	var resp FundAccountsResponse

	gasStationMnemonic := config.GetGasStationMnemonic()
	rawKey, err := cosmosutils.RecoverKeyFromMnemonic(state.binaryPath, common.WeaveGasStationKeyName, gasStationMnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to recover gas station key: %v", err)
	}
	defer func() {
		_ = cosmosutils.DeleteKey(state.binaryPath, common.WeaveGasStationKeyName)
	}()

	gasStationKey, err := cosmosutils.UnmarshalKeyInfo(rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal gas station key: %v", err)
	}
	var rawTxContent string
	if state.batchSubmissionIsCelestia {
		rawTxContent = fmt.Sprintf(
			FundMinitiaAccountsWithoutBatchTxInterface,
			gasStationKey.Address,
			lsk.BridgeExecutor.Address,
			lsk.BridgeExecutor.Coins,
			lsk.OutputSubmitter.Address,
			lsk.OutputSubmitter.Coins,
			lsk.Challenger.Address,
			lsk.Challenger.Coins,
		)
		_, err = cosmosutils.RecoverKeyFromMnemonic(state.celestiaBinaryPath, common.WeaveGasStationKeyName, gasStationMnemonic)
		if err != nil {
			return nil, fmt.Errorf("failed to recover celestia gas station key: %v", err)
		}
		defer func() {
			_ = cosmosutils.DeleteKey(state.celestiaBinaryPath, common.WeaveGasStationKeyName)
		}()

		// TODO: Choose DA layer based on the chosen L1 network
		celestiaRegistry, err := registry.GetChainRegistry(registry.CelestiaTestnet)
		if err != nil {
			return nil, fmt.Errorf("failed to get celestia registry: %v", err)
		}

		celestiaRpc, err := celestiaRegistry.GetActiveRpc()
		if err != nil {
			return nil, fmt.Errorf("failed to get active rpc for celestia: %v", err)
		}

		//celestiaMinGasPrice, err := celestiaRegistry.GetMinGasPriceByDenom(DefaultCelestiaGasDenom)
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get celestia minimum gas price: %v", err)
		//}

		celestiaChainId := celestiaRegistry.GetChainId()
		sendCmd := exec.Command(state.celestiaBinaryPath, "tx", "bank", "send", common.WeaveGasStationKeyName,
			lsk.BatchSubmitter.Address, fmt.Sprintf("%sutia", lsk.BatchSubmitter.Coins), "--node", celestiaRpc,
			"--chain-id", celestiaChainId, "--gas", "200000", "--gas-prices", "0.1utia", "--output", "json", "-y",
		)
		broadcastRes, err := sendCmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to broadcast transaction: %v", err)
		}

		var txResponse cosmosutils.InitiadTxResponse
		err = json.Unmarshal(broadcastRes, &txResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		if txResponse.Code != 0 {
			return nil, fmt.Errorf("celestia tx failed with error: %v", txResponse.RawLog)
		}
		err = lsk.waitForTransactionInclusion(state.celestiaBinaryPath, celestiaRpc, txResponse.TxHash)
		if err != nil {
			return nil, err
		}
		resp.CelestiaTx = &txResponse
	} else {
		rawTxContent = fmt.Sprintf(
			FundMinitiaAccountsDefaultTxInterface,
			gasStationKey.Address,
			lsk.BridgeExecutor.Address,
			lsk.BridgeExecutor.Coins,
			lsk.OutputSubmitter.Address,
			lsk.OutputSubmitter.Coins,
			lsk.BatchSubmitter.Address,
			lsk.BatchSubmitter.Coins,
			lsk.Challenger.Address,
			lsk.Challenger.Coins,
		)
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home: %v", err)
	}
	rawTxPath := filepath.Join(userHome, common.WeaveDataDirectory, TmpTxFilename)
	if err = io.WriteFile(rawTxPath, rawTxContent); err != nil {
		return nil, fmt.Errorf("failed to write raw tx file: %v", err)
	}
	defer func() {
		if err := io.DeleteFile(rawTxPath); err != nil {
			fmt.Printf("failed to delete raw tx file: %v", err)
		}
	}()

	signCmd := exec.Command(state.binaryPath, "tx", "sign", rawTxPath, "--from", common.WeaveGasStationKeyName, "--node", state.l1RPC,
		"--chain-id", state.l1ChainId, "--keyring-backend", "test", "--output-document", rawTxPath)
	err = signCmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	broadcastCmd := exec.Command(state.binaryPath, "tx", "broadcast", rawTxPath, "--node", state.l1RPC, "--output", "json")
	broadcastRes, err := broadcastCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %v", err)
	}

	var txResponse cosmosutils.InitiadTxResponse
	err = json.Unmarshal(broadcastRes, &txResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	if txResponse.Code != 0 {
		return nil, fmt.Errorf("initia l1 tx failed with error: %v", txResponse.RawLog)
	}

	err = lsk.waitForTransactionInclusion(state.binaryPath, state.l1RPC, txResponse.TxHash)
	if err != nil {
		return nil, err
	}
	resp.InitiaTx = &txResponse

	return &resp, nil
}

// waitForTransactionInclusion polls for the transaction inclusion in a block
func (lsk *L1SystemKeys) waitForTransactionInclusion(binaryPath, rpcURL, txHash string) error {
	// Poll for transaction status until it's included in a block
	timeout := time.After(15 * time.Second)   // Example timeout for polling
	ticker := time.NewTicker(3 * time.Second) // Poll every 3 seconds
	defer ticker.Stop()                       // Ensure cleanup of ticker resource

	for {
		select {
		case <-timeout:
			return fmt.Errorf("transaction not included in block within timeout")
		case <-ticker.C:
			// Query transaction status
			statusCmd := exec.Command(binaryPath, "query", "tx", txHash, "--node", rpcURL, "--output", "json")
			statusRes, err := statusCmd.CombinedOutput()
			// If the transaction is not included in a block yet, just continue polling
			if err != nil {
				// You can add more detailed error handling here if needed,
				// but for now, we continue polling if it returns an error (i.e., "not found" or similar).
				continue
			}

			var txResponse cosmosutils.MinimalTxResponse
			err = json.Unmarshal(statusRes, &txResponse)
			if err != nil {
				return fmt.Errorf("failed to unmarshal transaction JSON response: %v", err)
			}
			if txResponse.Code == 0 { // Successful transaction
				// Transaction successfully included in block
				return nil
			} else {
				return fmt.Errorf("tx failed with error: %v", txResponse.RawLog)
			}

			// If the transaction is not in a block yet, continue polling
		}
	}
}

const FundMinitiaAccountsDefaultTxInterface = `
{
  "body":{
    "messages":[
      {
        "@type":"/cosmos.bank.v1beta1.MsgSend",
        "from_address":"%[1]s",
        "to_address":"%[2]s",
        "amount":[
          {
            "denom":"uinit",
            "amount":"%[3]s"
          }
        ]
      },
      {
        "@type":"/cosmos.bank.v1beta1.MsgSend",
        "from_address":"%[1]s",
        "to_address":"%[4]s",
        "amount":[
          {
            "denom":"uinit",
            "amount":"%[5]s"
          }
        ]
      },
      {
        "@type":"/cosmos.bank.v1beta1.MsgSend",
        "from_address":"%[1]s",
        "to_address":"%[6]s",
        "amount":[
          {
            "denom":"uinit",
            "amount":"%[7]s"
          }
        ]
      },
      {
        "@type":"/cosmos.bank.v1beta1.MsgSend",
        "from_address":"%[1]s",
        "to_address":"%[8]s",
        "amount":[
          {
            "denom":"uinit",
            "amount":"%[9]s"
          }
        ]
      }
    ],
    "memo":"Sent from Weave Gas Station!",
    "timeout_height":"0",
    "extension_options":[],
    "non_critical_extension_options":[]
  },
  "auth_info":{
    "signer_infos":[],
    "fee":{
      "amount":[
        {
          "denom":"uinit",
          "amount":"12000"
        }
      ],
      "gas_limit":"800000",
      "payer":"",
      "granter":""
    },
    "tip":null
  },
  "signatures":[]
}
`

const FundMinitiaAccountsWithoutBatchTxInterface = `
{
  "body":{
    "messages":[
      {
        "@type":"/cosmos.bank.v1beta1.MsgSend",
        "from_address":"%[1]s",
        "to_address":"%[2]s",
        "amount":[
          {
            "denom":"uinit",
            "amount":"%[3]s"
          }
        ]
      },
      {
        "@type":"/cosmos.bank.v1beta1.MsgSend",
        "from_address":"%[1]s",
        "to_address":"%[4]s",
        "amount":[
          {
            "denom":"uinit",
            "amount":"%[5]s"
          }
        ]
      },
      {
        "@type":"/cosmos.bank.v1beta1.MsgSend",
        "from_address":"%[1]s",
        "to_address":"%[6]s",
        "amount":[
          {
            "denom":"uinit",
            "amount":"%[7]s"
          }
        ]
      }
    ],
    "memo":"Sent from Weave Gas Station!",
    "timeout_height":"0",
    "extension_options":[],
    "non_critical_extension_options":[]
  },
  "auth_info":{
    "signer_infos":[],
    "fee":{
      "amount":[
        {
          "denom":"uinit",
          "amount":"10500"
        }
      ],
      "gas_limit":"700000",
      "payer":"",
      "granter":""
    },
    "tip":null
  },
  "signatures":[]
}
`
