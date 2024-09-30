package flags

import (
	"testing"

	"github.com/test-go/testify/assert"
)

func TestIsEnabled(t *testing.T) {
	testCases := []struct {
		name          string
		enabledFlags  string
		flag          FeatureFlag
		expectedValue bool
	}{
		{
			name:          "Single flag enabled",
			enabledFlags:  "minitia_launch",
			flag:          MinitiaLaunch,
			expectedValue: true,
		},
		{
			name:          "Multiple flags enabled, target flag present",
			enabledFlags:  "minitia_launch,opinit_bots,relayer",
			flag:          OPInitBots,
			expectedValue: true,
		},
		{
			name:          "Multiple flags enabled, target flag not present",
			enabledFlags:  "minitia_launch,opinit_bots",
			flag:          Relayer,
			expectedValue: false,
		},
		{
			name:          "No flags enabled",
			enabledFlags:  "",
			flag:          MinitiaLaunch,
			expectedValue: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			EnabledFlags = tc.enabledFlags
			result := IsEnabled(tc.flag)
			assert.Equal(t, tc.expectedValue, result)
		})
	}
}
