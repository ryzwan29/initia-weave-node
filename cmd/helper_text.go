package cmd

import "fmt"

const (
	DocsURLPrefix = "https://github.com/initia-labs/weave/blob/main/"
)

func SubHelperText(docPath string) string {
	return fmt.Sprintf("See %s%s for more information about the setup process and potential issues.", DocsURLPrefix, docPath)
}

var (
	WeaveHelperText      = fmt.Sprintf("Weave is the CLI for managing Initia deployments.\n\nSee %sREADME.md for more information.", DocsURLPrefix)
	L1NodeHelperText     = SubHelperText("docs/initia_node.md")
	RollupHelperText     = SubHelperText("docs/rollup_launch.md")
	OPinitBotsHelperText = SubHelperText("docs/opinit_bots.md")
	RelayerHelperText    = SubHelperText("docs/relayer.md")
	GasStationHelperText = SubHelperText("docs/gas_station.md")
)
