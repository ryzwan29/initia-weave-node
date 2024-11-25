package relayer

import "github.com/initia-labs/weave/types"

type RelayerState struct {
	weave types.WeaveState
}

func NewRelayerState() RelayerState {
	return RelayerState{
		weave: types.NewWeaveState(),
	}
}

func (state RelayerState) Clone() RelayerState {
	clone := RelayerState{
		weave: state.weave,
	}

	return clone
}
