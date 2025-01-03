package analytics

import (
	"runtime"
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

type EventAtrriutes map[string]interface{}

func AppendGlobalEventProperties(properties EventAtrriutes) {
	if GlobalEventProperties == nil {
		GlobalEventProperties = make(EventAtrriutes)
	}

	for k, v := range properties {
		GlobalEventProperties[k] = v
	}
}

func NewEmptyEvent() EventAtrriutes {
	return make(EventAtrriutes)
}

func TrackEvent(eventType Event, overrideProperties EventAtrriutes) {
	eventProperties := make(EventAtrriutes)
	for k, v := range GlobalEventProperties {
		eventProperties[k] = v
	}

	for k, v := range overrideProperties {
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

func TrackRunEvent(cmd *cobra.Command, component Component) {
	AppendGlobalEventProperties(EventAtrriutes{
		ComponentEventKey: component,
		CommandEventKey:   cmd.CommandPath(),
	})
	TrackEvent(RunEvent, nil)
}

func TrackCompletedEvent(cmd *cobra.Command, component Component) {
	AppendGlobalEventProperties(EventAtrriutes{
		ComponentEventKey: component,
		CommandEventKey:   cmd.CommandPath(),
	})
	TrackEvent(CompletedEvent, nil)
}

func (ea EventAtrriutes) Add(key string, value interface{}) {
	ea[key] = value
}
