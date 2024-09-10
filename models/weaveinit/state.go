package weaveinit

type RunL1NodeState struct {
	network            string
	initiadVersion     string
	chainId            string
	moniker            string
	existingApp        bool
	replaceExistingApp bool
	minGasPrice        string
	seeds              string
	persistentPeers    string
	existingGenesis    bool
}
