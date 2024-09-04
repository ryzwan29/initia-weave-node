package states

import (
	"sync"
)

type GlobalState struct {
	mu sync.RWMutex

	mnemonic string
}

var globalStateInstance *GlobalState
var once sync.Once

func GetGlobalState() *GlobalState {
	once.Do(func() {
		globalStateInstance = &GlobalState{}
	})
	return globalStateInstance
}

func (gs *GlobalState) GetMnemonic() string {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return gs.mnemonic
}

func (gs *GlobalState) SetMnemonic(mnemonic string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.mnemonic = mnemonic
}
