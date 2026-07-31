package gossip

import (
	"github.com/0xlength/sidewinder/cmd/sidewinder/gossip/ping"
	"github.com/0xlength/sidewinder/cmd/sidewinder/gossip/pull"
	"github.com/spf13/cobra"
)

var Cmd = cobra.Command{
	Use:   "gossip",
	Short: "Interact with Solana gossip networks",
}

func init() {
	Cmd.AddCommand(
		&ping.Cmd,
		&pull.Cmd,
	)
}
