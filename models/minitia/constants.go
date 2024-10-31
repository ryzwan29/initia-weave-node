package minitia

const (
	AppName string = "minitiad"

	OperatorKeyName        string = "weave.Operator"
	BridgeExecutorKeyName  string = "weave.BridgeExecutor"
	OutputSubmitterKeyName string = "weave.OutputSubmitter"
	BatchSubmitterKeyName  string = "weave.BatchSubmitter"
	ChallengerKeyName      string = "weave.Challenger"

	TmpTxFilename string = "weave.minitia.tx.json"

	DefaultL1GasPrices string = "0.015uinit"

	MaxMonikerLength int = 70
	MaxChainIDLength int = 50

	LaunchConfigFilename = "minitia.config.json"

	CelestiaAppName string = "celestia-appd"

	InitiaScanURL         string = "https://scan.testnet.initia.xyz"
	DefaultMinitiaLCD     string = "http://localhost:1317"
	DefaultMinitiaRPC     string = "http://localhost:26657"
	DefaultMinitiaJsonRPC string = "http://localhost:8545"
)
