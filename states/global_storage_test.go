package states_test

import (
	"testing"

	"github.com/initia-labs/weave/states"
)

func TestGlobalStorageSingleton(t *testing.T) {
	instance1 := states.GetGlobalStorage()
	instance2 := states.GetGlobalStorage()

	if instance1 != instance2 {
		t.Errorf("Expected singleton instances to be the same, got different instances")
	}
}
