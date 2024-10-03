package minitia

import (
	"encoding/json"
	"fmt"
	"github.com/initia-labs/weave/utils"
	"os"
	"os/exec"
	"path/filepath"
)

type L1SystemKeys struct {
	Operator        *GenesisAccount
	BridgeExecutor  *GenesisAccount
	OutputSubmitter *GenesisAccount
	BatchSubmitter  *GenesisAccount
	Challenger      *GenesisAccount
}

func NewL1SystemKeys(operator, bridgeExecutor, outputSubmitter, batchSubmitter, challenger *GenesisAccount) *L1SystemKeys {
	return &L1SystemKeys{
		Operator:        operator,
		BridgeExecutor:  bridgeExecutor,
		OutputSubmitter: outputSubmitter,
		BatchSubmitter:  batchSubmitter,
		Challenger:      challenger,
	}
}

func (lsk *L1SystemKeys) FundAccountsWithGasStation(appName, rpc, chainId string) (*utils.CliTxResponse, error) {
	gasStationMnemonic := utils.GetConfig("common.gas_station_mnemonic").(string)
	rawKey, err := utils.RecoverKeyFromMnemonic(appName, utils.WeaveGasStationKeyName, gasStationMnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to recover gas station key: %v", err)
	}
	defer utils.DeleteKey(appName, utils.WeaveGasStationKeyName)

	gasStationKey := utils.MustUnmarshalKeyInfo(rawKey)
	rawTxContent := fmt.Sprintf(
		FundMinitiaAccountsTxInterface,
		gasStationKey.Address,
		lsk.Operator.Address,
		lsk.Operator.Coins,
		lsk.BridgeExecutor.Address,
		lsk.BridgeExecutor.Coins,
		lsk.OutputSubmitter.Address,
		lsk.OutputSubmitter.Coins,
		lsk.BatchSubmitter.Address,
		lsk.BatchSubmitter.Coins,
		lsk.Challenger.Address,
		lsk.Challenger.Coins,
	)
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home: %v", err)
	}
	rawTxPath := filepath.Join(userHome, utils.WeaveDataDirectory, TmpTxFilename)
	if err = utils.WriteFile(rawTxPath, rawTxContent); err != nil {
		return nil, fmt.Errorf("failed to write raw tx file: %v", err)
	}
	defer utils.DeleteFile(rawTxPath)

	signCmd := exec.Command(appName, "tx", "sign", rawTxPath, "--from", utils.WeaveGasStationKeyName, "--node", rpc,
		"--chain-id", chainId, "--keyring-backend", "test", "--output-document", rawTxPath)
	err = signCmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	broadcastCmd := exec.Command(appName, "tx", "broadcast", rawTxPath, "--node", rpc, "--output", "json")
	broadcastRes, err := broadcastCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %v", err)
	}

	var txResponse utils.CliTxResponse
	err = json.Unmarshal(broadcastRes, &txResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &txResponse, nil
}

const FundMinitiaAccountsTxInterface = `
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
      },
      {
        "@type":"/cosmos.bank.v1beta1.MsgSend",
        "from_address":"%[1]s",
        "to_address":"%[10]s",
        "amount":[
          {
            "denom":"uinit",
            "amount":"%[11]s"
          }
        ]
      }
    ],
    "memo":"",
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
