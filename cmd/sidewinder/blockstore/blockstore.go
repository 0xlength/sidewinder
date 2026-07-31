//go:build !lite

package blockstore

import (
	"github.com/0xlength/sidewinder/cmd/sidewinder/blockstore/compact"
	"github.com/0xlength/sidewinder/cmd/sidewinder/blockstore/dumpbatches"
	"github.com/0xlength/sidewinder/cmd/sidewinder/blockstore/dumpshreds"
	"github.com/0xlength/sidewinder/cmd/sidewinder/blockstore/statdatarate"
	"github.com/0xlength/sidewinder/cmd/sidewinder/blockstore/statentries"
	"github.com/0xlength/sidewinder/cmd/sidewinder/blockstore/tarblocks"
	"github.com/0xlength/sidewinder/cmd/sidewinder/blockstore/verifydata"
	"github.com/0xlength/sidewinder/cmd/sidewinder/blockstore/yaml"
	"github.com/spf13/cobra"
)

var Cmd = cobra.Command{
	Use:   "blockstore",
	Short: "Access blockstore database",
}

func init() {
	Cmd.AddCommand(
		&compact.Cmd,
		&dumpshreds.Cmd,
		&dumpbatches.Cmd,
		&statdatarate.Cmd,
		&statentries.Cmd,
		&tarblocks.Cmd,
		&verifydata.Cmd,
		&yaml.Cmd,
	)
}
