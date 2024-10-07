package types

type MinitiaConfig struct {
	L1Config        L1NetworkConfig  `json:"l1_config"`
	L2Config        L2NetworkConfig  `json:"l2_config"`
	BridgeConfig    BridgeConfig     `json:"op_bridge"`
	SystemKeys      ValidatorKeys    `json:"system_keys"`
	GenesisAccounts []GenesisAccount `json:"genesis_accounts"`
}

type L1NetworkConfig struct {
	ChainID   string `json:"chain_id"`
	RPCURL    string `json:"rpc_url"`
	GasPrices string `json:"gas_prices"`
}

type L2NetworkConfig struct {
	ChainID string `json:"chain_id"`
	Denom   string `json:"denom"`
	Moniker string `json:"moniker"`
}

type BridgeConfig struct {
	OutputSubmissionStartTime string `json:"output_submission_start_time"`
	OutputSubmissionInterval  int64  `json:"output_submission_interval"`
	OutputFinalizationPeriod  int64  `json:"output_finalization_period"`
	BatchSubmissionTarget     string `json:"batch_submission_target"`
}

type ValidatorKeys struct {
	Validator       KeyDetails `json:"validator"`
	BridgeExecutor  KeyDetails `json:"bridge_executor"`
	OutputSubmitter KeyDetails `json:"output_submitter"`
	BatchSubmitter  KeyDetails `json:"batch_submitter"`
	Challenger      KeyDetails `json:"challenger"`
}

type KeyDetails struct {
	L1Address string `json:"l1_address"`
	L2Address string `json:"l2_address"`
	Mnemonic  string `json:"mnemonic"`
}

type GenesisAccount struct {
	Address string `json:"address"`
}
