package cmd

import "fmt"

const (
	DocsURLPrefix = "https://github.com/initia-labs/weave/blob/main/"
)

func SubHeplerText(docPath string) string {
	return fmt.Sprintf("See %s%s for more information about the setup process and potential issues.", DocsURLPrefix, docPath)
}

var (
	WeaveHelperText      = fmt.Sprintf("Weave is the CLI for managing Initia deployments.\n\nSee %sREADME.md for more information.", DocsURLPrefix)
	L1NodeHelperText     = SubHeplerText("docs/initia_node.md")
	RollupHelperText     = SubHeplerText("docs/rollup_launch.md")
	OPinitBotsHelperText = SubHeplerText("docs/opinit_bots.md")
	RelayerHelperText    = SubHeplerText("docs/relayer.md")
	GasStationHelperText = SubHeplerText("docs/gas_station.md")
)
