package weaveinit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWeaveInitOptions(t *testing.T) {
	options := GetWeaveInitOptions()

	assert.Equal(t, 4, len(options))
	// order of options is important
	assert.Equal(t, RunL1NodeOption, options[0])
	assert.Equal(t, LaunchNewRollupOption, options[1])
	assert.Equal(t, RunOPBotsOption, options[2])
	assert.Equal(t, RunRelayerOption, options[3])
}
