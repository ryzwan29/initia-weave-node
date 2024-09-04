// global_storage.go
package states

import (
	"sync"
)

type GlobalStorage struct {
	mu         sync.RWMutex
	textInputs map[string]string // Holds text data for various input fields
}

var globalStorageInstance *GlobalStorage
var globalStorageOnce sync.Once

// GetGlobalStorage returns the singleton instance of GlobalStorage
func GetGlobalStorage() *GlobalStorage {
	globalStorageOnce.Do(func() {
		globalStorageInstance = &GlobalStorage{
			textInputs: make(map[string]string),
		}
	})
	return globalStorageInstance
}

func (gs *GlobalStorage) SetText(fieldID, text string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.textInputs[fieldID] = text
}

func (gs *GlobalStorage) GetText(fieldID string) (string, bool) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	text, ok := gs.textInputs[fieldID]
	return text, ok
}
