package tpu_quic

import (
	"github.com/0xlength/sidewinder/cmd/sidewinder/tpu_quic/ping"
	"github.com/spf13/cobra"
)

var Cmd = cobra.Command{
	Use:   "tpu-quic",
	Short: "TPU/QUIC tools",
}

func init() {
	Cmd.AddCommand(
		&ping.Cmd,
	)
}
