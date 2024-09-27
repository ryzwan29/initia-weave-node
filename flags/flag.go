package flags

import "strings"

var EnabledFlags string

type FeatureFlag string

const (
	MinitiaLaunch      FeatureFlag = "minitia_launch"
	OPInitBotsExecutor FeatureFlag = "opinit-bots-executor"
	Relayer            FeatureFlag = "relayer"
)

func IsEnabled(flag FeatureFlag) bool {
	return strings.Contains(EnabledFlags, string(flag))
}
