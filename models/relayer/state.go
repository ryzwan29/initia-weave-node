package relayer

import "github.com/initia-labs/weave/types"

type RelayerState struct {
	weave  types.WeaveState
	Config map[string]string
}

func NewRelayerState() RelayerState {
	return RelayerState{
		weave:  types.NewWeaveState(),
		Config: make(map[string]string),
	}
}

func (state RelayerState) Clone() RelayerState {
	config := make(map[string]string)
	for k, v := range state.Config {
		config[k] = v
	}
	clone := RelayerState{
		weave:  state.weave,
		Config: config,
	}

	return clone
}
