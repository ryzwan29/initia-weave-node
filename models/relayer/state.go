package relayer

import (
	"github.com/initia-labs/weave/types"
)

type RelayerState struct {
	weave       types.WeaveState
	Config      map[string]string
	IBCChannels []types.IBCChannelPair

	l1KeyMethod       string
	l1RelayerAddress  string
	l1RelayerMnemonic string
	l1NeedsFunding    bool

	l2KeyMethod       string
	l2RelayerAddress  string
	l2RelayerMnemonic string
	l2NeedsFunding    bool

	hermesBinaryPath string
}

func NewRelayerState() RelayerState {
	return RelayerState{
		weave:       types.NewWeaveState(),
		Config:      make(map[string]string),
		IBCChannels: make([]types.IBCChannelPair, 0),
	}
}

func (state RelayerState) Clone() RelayerState {
	config := make(map[string]string)
	for k, v := range state.Config {
		config[k] = v
	}
	clone := RelayerState{
		weave:       state.weave,
		Config:      config,
		IBCChannels: state.IBCChannels,
	}

	return clone
}
