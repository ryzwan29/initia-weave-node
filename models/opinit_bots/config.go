package opinit_bots

type NodeConfig struct {
	ChainID      string `json:"chain_id"`
	Bech32Prefix string `json:"bech32_prefix"`
	RPCAddress   string `json:"rpc_address"`
}

type ChallengerConfig struct {
	Version       int        `json:"version"`
	ListenAddress string     `json:"listen_address"`
	L1Node        NodeConfig `json:"l1_node"`
	L2Node        NodeConfig `json:"l2_node"`
	L2StartHeight int        `json:"l2_start_height"`
}

// Renaming NodeConfig to something more descriptive, such as NodeSettings, NodeConfiguration, or similar
type NodeSettings struct {
	ChainID       string  `json:"chain_id"`
	Bech32Prefix  string  `json:"bech32_prefix"`
	RPCAddress    string  `json:"rpc_address"`
	GasPrice      string  `json:"gas_price"`
	GasAdjustment float64 `json:"gas_adjustment"`
	TxTimeout     int     `json:"tx_timeout"`
}

type ExecutorConfig struct {
	Version               int          `json:"version"`
	ListenAddress         string       `json:"listen_address"`
	L1Node                NodeSettings `json:"l1_node"`
	L2Node                NodeSettings `json:"l2_node"`
	DANode                NodeSettings `json:"da_node"`
	OutputSubmitter       string       `json:"output_submitter"`
	BridgeExecutor        string       `json:"bridge_executor"`
	BatchSubmitterEnabled bool         `json:"enable_batch_submitter"`
	MaxChunks             int          `json:"max_chunks"`
	MaxChunkSize          int          `json:"max_chunk_size"`
	MaxSubmissionTime     int          `json:"max_submission_time"`
	L2StartHeight         int          `json:"l2_start_height"`
	BatchStartHeight      int          `json:"batch_start_height"`
}
