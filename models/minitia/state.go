package minitia

import "github.com/initia-labs/weave/types"

type LaunchState struct {
	weave types.WeaveState
}

func NewLaunchState() *LaunchState {
	return &LaunchState{
		weave: types.NewWeaveState(),
	}
}
