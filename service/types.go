package service

type CommandName string

const (
	UpgradableInitia    CommandName = "upgradable_initia"
	NonUpgradableInitia CommandName = "non_upgradable_initia"
	Minitia             CommandName = "minitia"
	OPinitExecutor      CommandName = "executor"
	OPinitChallenger    CommandName = "challenger"
	Relayer             CommandName = "relayer"
)

func (cmd CommandName) MustGetBinaryName() string {
	switch cmd {
	case UpgradableInitia, NonUpgradableInitia:
		return "cosmovisor"
	case Minitia:
		return "minitiad"
	case OPinitExecutor, OPinitChallenger:
		return "opinitd"
	case Relayer:
		return "hermes"
	default:
		panic("unsupported command")
	}
}

func (cmd CommandName) MustGetServiceSlug() string {
	switch cmd {
	case UpgradableInitia:
		return "cosmovisor"
	case NonUpgradableInitia:
		return "cosmovisor"
	case Minitia:
		return "minitiad"
	case OPinitExecutor:
		return "opinitd.executor"
	case OPinitChallenger:
		return "opinitd.challenger"
	case Relayer:
		return "hermes"
	default:
		panic("unsupported command")
	}
}
