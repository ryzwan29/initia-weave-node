package analytics

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/initia-labs/weave/config"
	"github.com/spf13/cobra"
)

var (
	Client                amplitude.Client
	SessionID             int64
	GlobalEventProperties map[string]interface{}
)

type disabledLogger struct{}

func (d disabledLogger) Debugf(format string, v ...interface{}) {}

func (d disabledLogger) Errorf(format string, v ...interface{}) {}

func (d disabledLogger) Infof(format string, v ...interface{}) {}

func (d disabledLogger) Warnf(format string, v ...interface{}) {}

func Initialize(weaveVersion string) {
	c := amplitude.NewConfig(AmplitudeKey)
	c.OptOut = config.AnalyticsOptOut()
	c.Logger = disabledLogger{}

	Client = amplitude.NewClient(c)
	identify := amplitude.Identify{}
	identify.Set("Arch", runtime.GOARCH)
	identify.Set("Go Version", runtime.Version())
	Client.Identify(identify, amplitude.EventOptions{
		DeviceID:   config.GetAnalyticsDeviceID(),
		OSName:     runtime.GOOS,
		AppVersion: weaveVersion,
	})

	SessionID = time.Now().Unix()
}

type EventAttributes map[string]interface{}

// Event represents an event with some attributes
type AmplitudeEvent struct {
	Attributes EventAttributes
}

func AppendGlobalEventProperties(properties EventAttributes) {
	if GlobalEventProperties == nil {
		GlobalEventProperties = make(EventAttributes)
	}

	for k, v := range properties {
		GlobalEventProperties[k] = v
	}
}

// NewEmptyEvent creates and returns an empty event
func NewEmptyEvent() *AmplitudeEvent {
	return &AmplitudeEvent{
		Attributes: make(EventAttributes),
	}
}

func TrackEvent(eventType Event, overrideProperties *AmplitudeEvent) {
	eventProperties := make(EventAttributes)
	for k, v := range GlobalEventProperties {
		eventProperties[k] = v
	}

	for k, v := range overrideProperties.Attributes {
		eventProperties[k] = v
	}

	Client.Track(amplitude.Event{
		EventType: string(eventType),
		EventOptions: amplitude.EventOptions{
			DeviceID:  config.GetAnalyticsDeviceID(),
			SessionID: int(SessionID),
		},
		EventProperties: eventProperties,
	})
}

func TrackRunEvent(cmd *cobra.Command, args []string, component Component) {
	AppendGlobalEventProperties(EventAttributes{
		ComponentEventKey: component,
		CommandEventKey:   cmd.CommandPath(),
	})
	for idx, arg := range args {
		AppendGlobalEventProperties(EventAttributes{
			fmt.Sprintf("arg-%d", idx): arg,
		})
	}
	TrackEvent(RunEvent, NewEmptyEvent())
}

func TrackCompletedEvent(cmd *cobra.Command, component Component) {
	AppendGlobalEventProperties(EventAttributes{
		ComponentEventKey: component,
		CommandEventKey:   cmd.CommandPath(),
	})
	TrackEvent(CompletedEvent, NewEmptyEvent())
}

// Add adds a key-value pair to the event's attributes
func (e *AmplitudeEvent) Add(key string, value interface{}) *AmplitudeEvent {
	if key != string(ModelNameKey) {
		if str, ok := value.(string); ok {
			value = strings.ToLower(str) // Convert string value to lowercase
		}
	}
	e.Attributes[key] = value
	return e
}
