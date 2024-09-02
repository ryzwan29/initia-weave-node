package types_test

import (
	"testing"

	"github.com/initia-labs/weave/types"
)

func TestOptionString(t *testing.T) {
	tests := []struct {
		name   string
		option types.Option
		want   string
	}{
		{"Run L1 Node", types.RunL1Node, "Run a L1 Node"},
		{"Launch New Minitia", types.LaunchNewMinitia, "Launch a New Minitia"},
		{"Run OPinit Bots", types.RunOPinitBots, "Run OPinit Bots"},
		{"Run Relayer", types.RunRelayer, "Run a Relayer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.option.String(); got != tt.want {
				t.Errorf("Option.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	want := []types.Option{
		types.RunL1Node,
		types.LaunchNewMinitia,
		types.RunOPinitBots,
		types.RunRelayer,
	}

	got := types.Options()

	if len(got) != len(want) {
		t.Fatalf("Options() returned %d items, want %d", len(got), len(want))
	}

	for i, option := range got {
		if option != want[i] {
			t.Errorf("Options() item %d = %v, want %v", i, option, want[i])
		}
	}
}
