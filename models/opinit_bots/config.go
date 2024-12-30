package opinit_bots

import "github.com/initia-labs/weave/crypto"

type NodeConfig struct {
	ChainID      string `json:"chain_id"`
	Bech32Prefix string `json:"bech32_prefix"`
	RPCAddress   string `json:"rpc_address"`
}

type ChallengerConfig struct {
	Version                int          `json:"version"`
	Server                 ServerConfig `json:"server"`
	L1Node                 NodeConfig   `json:"l1_node"`
	L2Node                 NodeConfig   `json:"l2_node"`
	L1StartHeight          int          `json:"l1_start_height"`
	L2StartHeight          int          `json:"l2_start_height"`
	DisableAutoSetL1Height bool         `json:"disable_auto_set_l1_height"`
}

type NodeSettings struct {
	ChainID       string  `json:"chain_id"`
	Bech32Prefix  string  `json:"bech32_prefix"`
	RPCAddress    string  `json:"rpc_address"`
	GasPrice      string  `json:"gas_price"`
	GasAdjustment float64 `json:"gas_adjustment"`
	TxTimeout     int     `json:"tx_timeout"`
}

type ServerConfig struct {
	Address      string `json:"address"`
	AllowOrigins string `json:"allow_origins"`
	AllowHeaders string `json:"allow_headers"`
	AllowMethods string `json:"allow_methods"`
}

type ExecutorConfig struct {
	Version                       int          `json:"version"`
	Server                        ServerConfig `json:"server"`
	L1Node                        NodeSettings `json:"l1_node"`
	L2Node                        NodeSettings `json:"l2_node"`
	DANode                        NodeSettings `json:"da_node"`
	BridgeExecutor                string       `json:"bridge_executor"`
	OracleBridgeExecutor          string       `json:"oracle_bridge_executor"`
	DisableOutputSubmitter        bool         `json:"disable_output_submitter"`
	DisableBatchSubmitter         bool         `json:"disable_batch_submitter"`
	MaxChunks                     int          `json:"max_chunks"`
	MaxChunkSize                  int          `json:"max_chunk_size"`
	MaxSubmissionTime             int          `json:"max_submission_time"`
	DisableAutoSetL1Height        bool         `json:"disable_auto_set_l1_height"`
	L1StartHeight                 int          `json:"l1_start_height"`
	L2StartHeight                 int          `json:"l2_start_height"`
	BatchStartHeight              int          `json:"batch_start_height"`
	DisableDeleteFutureWithdrawal bool         `json:"disable_delete_future_withdrawal"`
}

type KeyFile struct {
	BridgeExecutor       string `json:"bridge_executor,omitempty"`
	OutputSubmitter      string `json:"output_submitter,omitempty"`
	Challenger           string `json:"challenger,omitempty"`
	BatchSubmitter       string `json:"batch_submitter,omitempty"`
	OracleBridgeExecutor string `json:"oracle_bridge_executor,omitempty"`
}

func GenerateMnemonicKefile() (KeyFile, error) {
	mnemonics := make([]string, 0)

	for idx := 0; idx < len(BotNames); idx++ {
		mnemonic, err := crypto.GenerateMnemonic()
		if err != nil {
			return KeyFile{}, err
		}
		mnemonics = append(mnemonics, mnemonic)
	}

	return KeyFile{
		BridgeExecutor:       mnemonics[0],
		OutputSubmitter:      mnemonics[1],
		Challenger:           mnemonics[2],
		BatchSubmitter:       mnemonics[3],
		OracleBridgeExecutor: mnemonics[4],
	}, nil
}
