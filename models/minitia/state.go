package minitia

import "github.com/initia-labs/weave/types"

type LaunchState struct {
	weave              types.WeaveState
	existingMinitiaApp bool
	l1Network          string
	vmType             string
}

func NewLaunchState() *LaunchState {
	return &LaunchState{
		weave: types.NewWeaveState(),
	}
}
