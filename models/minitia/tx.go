package minitia

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/initia-labs/weave/types"
	"github.com/initia-labs/weave/utils"
)

type L1SystemKeys struct {
	Operator        *types.GenesisAccount
	BridgeExecutor  *types.GenesisAccount
	OutputSubmitter *types.GenesisAccount
	BatchSubmitter  *types.GenesisAccount
	Challenger      *types.GenesisAccount
}

func NewL1SystemKeys(operator, bridgeExecutor, outputSubmitter, batchSubmitter, challenger *types.GenesisAccount) *L1SystemKeys {
	return &L1SystemKeys{
		Operator:        operator,
		BridgeExecutor:  bridgeExecutor,
		OutputSubmitter: outputSubmitter,
		BatchSubmitter:  batchSubmitter,
		Challenger:      challenger,
	}
}

type FundAccountsResponse struct {
	CelestiaTx *utils.CliTxResponse
	InitiaTx   *utils.CliTxResponse
}

func (lsk *L1SystemKeys) FundAccountsWithGasStation(state *LaunchState) (*FundAccountsResponse, error) {
	var resp FundAccountsResponse

	gasStationMnemonic := utils.GetConfig("common.gas_station_mnemonic").(string)
	rawKey, err := utils.RecoverKeyFromMnemonic(state.binaryPath, utils.WeaveGasStationKeyName, gasStationMnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to recover gas station key: %v", err)
	}
	defer utils.MustDeleteKey(state.binaryPath, utils.WeaveGasStationKeyName)

	gasStationKey := utils.MustUnmarshalKeyInfo(rawKey)
	var rawTxContent string
	if state.batchSubmissionIsCelestia {
		rawTxContent = fmt.Sprintf(
			FundMinitiaAccountsWithoutBatchTxInterface,
			gasStationKey.Address,
			lsk.Operator.Address,
			lsk.Operator.Coins,
			lsk.BridgeExecutor.Address,
			lsk.BridgeExecutor.Coins,
			lsk.OutputSubmitter.Address,
			lsk.OutputSubmitter.Coins,
			lsk.Challenger.Address,
			lsk.Challenger.Coins,
		)
		_ = utils.MustRecoverKeyFromMnemonic(state.celestiaBinaryPath, utils.WeaveGasStationKeyName, gasStationMnemonic)
		defer utils.MustDeleteKey(state.celestiaBinaryPath, utils.WeaveGasStationKeyName)

		// TODO: Choose DA layer based on the chosen L1 network
		celestiaRpc := utils.GetConfig("constants.da_layer.celestia_testnet.rpc").(string)
		celestiaChainId := utils.GetConfig("constants.da_layer.celestia_testnet.chain_id").(string)
		sendCmd := exec.Command(state.celestiaBinaryPath, "tx", "bank", "send", utils.WeaveGasStationKeyName,
			lsk.BatchSubmitter.Address, fmt.Sprintf("%sutia", lsk.BatchSubmitter.Coins), "--node", celestiaRpc,
			"--chain-id", celestiaChainId, "--gas", "100000", "--gas-prices", "0.1utia", "--output", "json", "-y",
		)
		broadcastRes, err := sendCmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to broadcast transaction: %v", err)
		}

		var txResponse utils.CliTxResponse
		err = json.Unmarshal(broadcastRes, &txResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		if txResponse.Code != 0 {
			return nil, fmt.Errorf("celestia tx failed with error: %v", txResponse.RawLog)
		}
		resp.CelestiaTx = &txResponse
	} else {
		rawTxContent = fmt.Sprintf(
			FundMinitiaAccountsDefaultTxInterface,
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
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home: %v", err)
	}
	rawTxPath := filepath.Join(userHome, utils.WeaveDataDirectory, TmpTxFilename)
	if err = utils.WriteFile(rawTxPath, rawTxContent); err != nil {
		return nil, fmt.Errorf("failed to write raw tx file: %v", err)
	}
	defer utils.DeleteFile(rawTxPath)

	signCmd := exec.Command(state.binaryPath, "tx", "sign", rawTxPath, "--from", utils.WeaveGasStationKeyName, "--node", state.l1RPC,
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

	var txResponse utils.CliTxResponse
	err = json.Unmarshal(broadcastRes, &txResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	if txResponse.Code != 0 {
		return nil, fmt.Errorf("initia l1 tx failed with error: %v", txResponse.RawLog)
	}
	resp.InitiaTx = &txResponse

	return &resp, nil
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
