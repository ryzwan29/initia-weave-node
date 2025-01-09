//go:build !test
// +build !test

package analytics

import "github.com/amplitude/analytics-go/amplitude"

// Define a No-Op Client for tests

// NoOpClient is a no-op implementation of the amplitude.Client interface
type NoOpClient struct{}

// Track does nothing in the NoOpClient
func (n *NoOpClient) Track(event amplitude.Event) {}

// Identify does nothing in the NoOpClient
func (n *NoOpClient) Identify(identify amplitude.Identify, eventOptions amplitude.EventOptions) {}

// GroupIdentify does nothing in the NoOpClient
func (n *NoOpClient) GroupIdentify(groupType string, groupName string, identify amplitude.Identify, eventOptions amplitude.EventOptions) {
}

// SetGroup does nothing in the NoOpClient
func (n *NoOpClient) SetGroup(groupType string, groupName []string, eventOptions amplitude.EventOptions) {
}

// Revenue does nothing in the NoOpClient
func (n *NoOpClient) Revenue(revenue amplitude.Revenue, eventOptions amplitude.EventOptions) {}

// Flush does nothing in the NoOpClient
func (n *NoOpClient) Flush() {}

// Shutdown does nothing in the NoOpClient
func (n *NoOpClient) Shutdown() {}

// Add does nothing in the NoOpClient
func (n *NoOpClient) Add(plugin amplitude.Plugin) {}

// Remove does nothing in the NoOpClient
func (n *NoOpClient) Remove(pluginName string) {}

// Config returns an empty Config in the NoOpClient
func (n *NoOpClient) Config() amplitude.Config {
	return amplitude.Config{}
}
