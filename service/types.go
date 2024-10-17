package service

type CommandName string

const (
	Initia           CommandName = "initia"
	Minitia          CommandName = "minitia"
	OPinitExecutor   CommandName = "executor"
	OPinitChallenger CommandName = "challenger"
)

func (cmd CommandName) MustGetBinaryName() string {
	switch cmd {
	case Initia:
		return "initiad"
	case Minitia:
		return "minitiad"
	case OPinitExecutor, OPinitChallenger:
		return "opinitd"
	default:
		panic("unsupported command")
	}
}

func (cmd CommandName) MustGetServiceSlug() string {
	switch cmd {
	case Initia:
		return "initiad"
	case Minitia:
		return "minitiad"
	case OPinitExecutor:
		return "opinitd.executor"
	case OPinitChallenger:
		return "opinitd.challenger"
	default:
		panic("unsupported command")
	}
}
