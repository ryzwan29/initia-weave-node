package types

// Option defines a type for the choices
type Option string

// Define the options as string constants
const (
	RunL1Node        Option = "Run a L1 Node"
	LaunchNewMinitia Option = "Launch a New Minitia"
	RunOPinitBots    Option = "Run OPinit Bots"
	RunRelayer       Option = "Run a Relayer"
)

func (o Option) String() string {
	return string(o)
}

// Options returns a slice of all options for display
func Options() []Option {
	return []Option{
		RunL1Node,
		LaunchNewMinitia,
		RunOPinitBots,
		RunRelayer,
	}
}
